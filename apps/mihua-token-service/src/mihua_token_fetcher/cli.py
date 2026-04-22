from __future__ import annotations

import argparse
import base64
import hashlib
import importlib
import json
import os
import re
import secrets
import shutil
import sqlite3
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.parse import parse_qs
from urllib.request import Request, urlopen

from .helpers import (
    LocatedToken,
    collect_allowed_hosts,
    extract_token_from_cookies,
    extract_token_from_headers,
    extract_token_from_json_text,
    extract_token_from_storage,
    find_repo_root,
    is_url_host_allowed,
    load_env_file,
    mask_token,
    write_env_value,
)


TOKEN_ENV_KEY = "MIHUA_CALL_RECORD_TOKEN"
ORIGIN_ENV_KEY = "MIHUA_CALL_RECORD_SOURCE_ORIGIN"
DEFAULT_MIHUA_USERNAME = "zzyjishu_spxxjj"
DEFAULT_MIHUA_PASSWORD = "Aa@123456"
SYSTEM_SETTING_TOKEN_KEY = "mihua_call_record_token"
SYSTEM_SETTING_TOKEN_DESCRIPTION = "米话通话记录 token"
URL_ENV_KEYS = (
    "MIHUA_CALL_RECORD_LIST_URL",
    "MIHUA_TELEMARKETING_RECORDING_LIST_URL",
    "MIHUA_TELEMARKETING_RECORDING_DETAIL_URL",
)
DEFAULT_CAPTCHA_LENGTH = 4
USERNAME_INPUT_SELECTORS = (
    'input[placeholder="用户名"]',
    'input[placeholder="账号"]',
    'input[placeholder*="用户名"]',
    'input[placeholder*="账号"]',
    'input[name="username"]',
    'input[name="account"]',
    'input[autocomplete="username"]',
)
PASSWORD_INPUT_SELECTORS = (
    'input[placeholder="密码"]',
    'input[placeholder*="密码"]',
    'input[name="password"]',
    'input[autocomplete="current-password"]',
    'input[type="password"]',
)
CAPTCHA_INPUT_SELECTORS = (
    'input[placeholder="验证码"]',
    'input[placeholder*="验证码"]',
    'input[name="captcha"]',
    'input[name="verify_code"]',
    'input[name="verifyCode"]',
    'input[maxlength="4"]',
)
LOGIN_BUTTON_SELECTORS = (
    'button:has-text("登录")',
    'button:has-text("登 录")',
    'input[type="submit"]',
)
_DDDD_OCR_INSTANCE: Any | None = None
_DDDD_OCR_INITIALIZED = False


@dataclass(slots=True)
class RuntimeConfig:
    repo_root: Path | None
    env_file: Path | None
    env_values: dict[str, str]
    origin: str
    open_url: str
    api_base_url: str
    allowed_hosts: set[str]


@dataclass(slots=True)
class DatabaseConfig:
    driver: str
    sqlite_path: Path | None = None
    mysql_host: str = ""
    mysql_port: int = 3306
    mysql_user: str = ""
    mysql_password: str = ""
    mysql_db: str = ""
    mysql_charset: str = "utf8mb4"


class MihuaApiError(RuntimeError):
    def __init__(self, message: str, *, retryable: bool = True) -> None:
        super().__init__(message)
        self.retryable = retryable


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="mihua-token-fetcher",
        description="打开米话页面并自动抓取 token。",
    )
    parser.add_argument("--origin", help="米话页面 origin，例如 https://spxxjj.emicloudcc.com")
    parser.add_argument("--open-url", help="浏览器启动后打开的页面地址，默认与 --origin 一致")
    parser.add_argument("--api-base-url", default="https://cmb.emicloudcc.com", help="米话 API 基础地址，默认 https://cmb.emicloudcc.com")
    parser.add_argument("--env-file", help="读取或回写的 env 文件路径，默认自动定位仓库根目录的 .env")
    parser.add_argument("--write-env", action="store_true", help="抓到 token 后回写 env 文件中的 MIHUA_CALL_RECORD_TOKEN")
    parser.add_argument("--headless", action="store_true", help="以无头模式运行浏览器")
    parser.add_argument("--timeout-seconds", type=int, default=180, help="等待 token 的秒数，默认 180")
    parser.add_argument("--poll-interval-ms", type=int, default=1200, help="轮询存储的间隔毫秒数，默认 1200")
    parser.add_argument("--browser-channel", help="可选浏览器通道，如 chrome 或 msedge")
    parser.add_argument("--username", default=DEFAULT_MIHUA_USERNAME, help="可选：自动填充登录用户名")
    parser.add_argument("--password", default=DEFAULT_MIHUA_PASSWORD, help="可选：自动填充登录密码")
    parser.add_argument("--captcha", help="可选：手动指定验证码")
    parser.add_argument("--auto-login-retries", type=int, default=3, help="自动登录重试次数，默认 3；同时用于 API 和浏览器登录重试")
    parser.add_argument("--captcha-image-path", help="将当前验证码图片保存到指定路径，便于人工识别")
    parser.add_argument("--force-browser", action="store_true", help="即使提供账号密码也强制走浏览器模式")
    parser.add_argument(
        "--write-db",
        "--write-system-setting",
        dest="write_db",
        action="store_true",
        help="抓到 token 后按仓库根目录 .env 的数据库配置回写 system_settings.mihua_call_record_token",
    )
    parser.add_argument("--print-token", action="store_true", help="显式输出完整 token")
    return parser


def locate_default_env_file(repo_root: Path | None) -> Path | None:
    if repo_root is None:
        return None
    candidate = repo_root / ".env"
    if candidate.exists():
        return candidate
    return candidate


def load_runtime_config(args: argparse.Namespace) -> RuntimeConfig:
    repo_root = find_repo_root(Path.cwd())
    env_file = Path(args.env_file).expanduser().resolve() if args.env_file else locate_default_env_file(repo_root)
    env_values = load_env_file(env_file)

    origin = (args.origin or os.getenv(ORIGIN_ENV_KEY) or env_values.get(ORIGIN_ENV_KEY, "")).strip()
    if not origin:
        raise SystemExit(
            "缺少米话 origin。请通过 --origin 传入，或者在仓库根目录 .env 中配置 MIHUA_CALL_RECORD_SOURCE_ORIGIN。"
        )

    open_url = (args.open_url or origin).strip()
    api_base_url = (args.api_base_url or "https://cmb.emicloudcc.com").strip().rstrip("/")
    host_inputs = [origin, open_url]
    for key in URL_ENV_KEYS:
        host_inputs.append(os.getenv(key, "").strip() or env_values.get(key, "").strip())
    host_inputs.append(api_base_url)
    allowed_hosts = collect_allowed_hosts(host_inputs)

    return RuntimeConfig(
        repo_root=repo_root,
        env_file=env_file,
        env_values=env_values,
        origin=origin,
        open_url=open_url,
        api_base_url=api_base_url,
        allowed_hosts=allowed_hosts,
    )


def import_playwright() -> Any:
    try:
        from playwright.sync_api import sync_playwright
    except ImportError as exc:  # pragma: no cover - runtime dependency branch
        raise SystemExit(
            "未安装 playwright。\n"
            "请先执行：\n"
            "1. pip install -e .\n"
            "2. python3 -m playwright install chromium"
        ) from exc
    return sync_playwright


def iter_local_venv_site_packages() -> list[Path]:
    tool_root = Path(__file__).resolve().parents[2]
    venv_dir = tool_root / ".venv"
    if not venv_dir.exists():
        return []
    candidates = [venv_dir / "Lib" / "site-packages", *venv_dir.glob("lib/python*/site-packages")]
    return [candidate for candidate in candidates if candidate.exists()]


def import_optional_module(module_name: str) -> Any | None:
    try:
        return importlib.import_module(module_name)
    except ImportError:
        pass

    for site_packages in iter_local_venv_site_packages():
        site_packages_str = str(site_packages)
        if site_packages_str not in sys.path:
            sys.path.insert(0, site_packages_str)
        try:
            return importlib.import_module(module_name)
        except ImportError:
            continue
    return None


def import_ddddocr() -> Any | None:
    return import_optional_module("ddddocr")


def get_ddddocr_instance() -> Any | None:
    global _DDDD_OCR_INITIALIZED, _DDDD_OCR_INSTANCE

    if _DDDD_OCR_INITIALIZED:
        return _DDDD_OCR_INSTANCE

    _DDDD_OCR_INITIALIZED = True
    ddddocr = import_ddddocr()
    if ddddocr is None:
        return None

    try:
        _DDDD_OCR_INSTANCE = ddddocr.DdddOcr(show_ad=False)
    except TypeError:
        _DDDD_OCR_INSTANCE = ddddocr.DdddOcr()
    except Exception:
        _DDDD_OCR_INSTANCE = None
    return _DDDD_OCR_INSTANCE


def capture_storage_snapshot(page: Any) -> dict[str, dict[str, str]]:
    try:
        return page.evaluate(
            """() => {
                const readStorage = (bucket) => {
                    const out = {};
                    for (let i = 0; i < bucket.length; i += 1) {
                        const key = bucket.key(i);
                        out[key] = bucket.getItem(key);
                    }
                    return out;
                };
                return {
                    localStorage: readStorage(window.localStorage),
                    sessionStorage: readStorage(window.sessionStorage),
                };
            }"""
        )
    except Exception:
        return {"localStorage": {}, "sessionStorage": {}}


def extract_captcha_data_uri(page: Any) -> str:
    try:
        return page.evaluate(
            """() => {
                const images = [...document.querySelectorAll('img')];
                const target = images.find((img) => {
                    return img.clientWidth >= 80 &&
                        img.clientWidth <= 140 &&
                        img.clientHeight >= 30 &&
                        img.clientHeight <= 60 &&
                        String(img.src || '').startsWith('data:image');
                });
                return target ? String(target.src || '') : '';
            }"""
        )
    except Exception:
        return ""


def decode_data_uri_image(data_uri: str) -> bytes:
    if not data_uri.startswith("data:image") or "," not in data_uri:
        return b""
    encoded = data_uri.split(",", 1)[1].strip()
    try:
        return base64.b64decode(encoded)
    except Exception:
        return b""


def normalize_ocr_text(raw: str) -> str:
    value = re.sub(r"[^A-Za-z0-9]", "", raw or "")
    normalized = value.strip()
    if len(normalized) >= DEFAULT_CAPTCHA_LENGTH:
        return normalized[:DEFAULT_CAPTCHA_LENGTH]
    return normalized


def is_directory_writable(directory: Path) -> bool:
    try:
        directory.mkdir(parents=True, exist_ok=True)
    except Exception:
        return False

    probe_path = directory / f".mihua-write-test-{os.getpid()}"
    try:
        probe_path.write_bytes(b"ok")
        probe_path.unlink(missing_ok=True)
        return True
    except Exception:
        try:
            probe_path.unlink(missing_ok=True)
        except Exception:
            pass
        return False


def build_captcha_output_path(args: argparse.Namespace) -> Path:
    if args.captcha_image_path:
        return Path(args.captcha_image_path).expanduser().resolve()

    candidates = [
        Path(tempfile.gettempdir()),
        Path.cwd() / ".cache",
        Path.home() / ".cache" / "mihua-token-fetcher",
    ]
    seen: set[Path] = set()
    for candidate in candidates:
        resolved = candidate.expanduser().resolve()
        if resolved in seen:
            continue
        seen.add(resolved)
        if is_directory_writable(resolved):
            return resolved / "mihua-captcha-latest.jpg"

    return (Path.cwd() / "mihua-captcha-latest.jpg").resolve()


def build_now_string() -> str:
    return datetime.now(UTC).strftime("%Y-%m-%d %H:%M:%S")


def solve_captcha_with_ddddocr(image_bytes: bytes) -> str:
    if not image_bytes:
        return ""
    ocr = get_ddddocr_instance()
    if ocr is None:
        return ""

    try:
        result = ocr.classification(image_bytes, png_fix=True)
    except TypeError:
        try:
            result = ocr.classification(image_bytes)
        except Exception:
            return ""
    except Exception:
        return ""

    if isinstance(result, dict):
        return normalize_ocr_text(str(result.get("text") or ""))
    return normalize_ocr_text(str(result or ""))


def solve_captcha_with_tesseract(image_bytes: bytes) -> str:
    if not image_bytes:
        return ""
    with tempfile.NamedTemporaryFile(prefix="mihua-captcha-", suffix=".jpg", delete=False) as handle:
        image_path = Path(handle.name)
        handle.write(image_bytes)

    try:
        try:
            result = subprocess.run(
                [
                    "tesseract",
                    str(image_path),
                    "stdout",
                    "--psm",
                    "7",
                    "-c",
                    "tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
                ],
                capture_output=True,
                text=True,
                check=False,
            )
        except FileNotFoundError:
            return ""
        if result.returncode != 0:
            return ""
        return normalize_ocr_text(result.stdout)
    finally:
        try:
            image_path.unlink(missing_ok=True)
        except Exception:
            pass


def build_random_nonce(length: int = 32) -> str:
    alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
    return "".join(secrets.choice(alphabet) for _ in range(max(length, 1)))


def write_captcha_image(image_bytes: bytes, target: Path) -> Path:
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(image_bytes)
    return target


def parse_json_object(text: str) -> dict[str, Any]:
    try:
        payload = json.loads(text)
    except json.JSONDecodeError:
        return {}
    if not isinstance(payload, dict):
        return {}
    return payload


def get_config_value(config: RuntimeConfig, key: str, default: str = "") -> str:
    return (os.getenv(key, "").strip() or config.env_values.get(key, "").strip() or default).strip()


def resolve_relative_path(base_dir: Path, raw_path: str) -> Path:
    candidate = Path(raw_path).expanduser()
    if candidate.is_absolute():
        return candidate.resolve()
    return (base_dir / candidate).resolve()


def parse_mysql_dsn(raw_dsn: str) -> dict[str, str]:
    dsn = raw_dsn.strip()
    match = re.match(
        r"^(?P<user>.*?):(?P<password>.*?)@tcp\((?P<host>[^:)]+)(?::(?P<port>\d+))?\)/(?P<db>[^?]+)(?:\?(?P<query>.*))?$",
        dsn,
    )
    if not match:
        return {}

    parsed: dict[str, str] = {
        "MYSQL_USER": match.group("user") or "",
        "MYSQL_PASSWORD": match.group("password") or "",
        "MYSQL_HOST": match.group("host") or "",
        "MYSQL_PORT": match.group("port") or "3306",
        "MYSQL_DB": match.group("db") or "",
    }
    query_text = match.group("query") or ""
    query = parse_qs(query_text, keep_blank_values=True)
    charset_values = query.get("charset") or query.get("CHARSET") or []
    if charset_values:
        parsed["MYSQL_CHARSET"] = str(charset_values[0]).strip()
    return parsed


def load_database_config(config: RuntimeConfig) -> DatabaseConfig:
    driver = get_config_value(config, "DB_DRIVER", "sqlite").lower() or "sqlite"
    backend_dir = config.env_file.parent if config.env_file is not None else (config.repo_root / "backend" if config.repo_root else Path.cwd())

    if driver == "sqlite":
        db_path = get_config_value(config, "DB_PATH", "data.db")
        if not db_path:
            raise SystemExit("写入数据库失败：缺少 DB_PATH 配置。")
        return DatabaseConfig(
            driver="sqlite",
            sqlite_path=resolve_relative_path(backend_dir, db_path),
        )

    if driver != "mysql":
        raise SystemExit(f"写入数据库失败：暂不支持的 DB_DRIVER={driver}")

    dsn_values = parse_mysql_dsn(get_config_value(config, "MYSQL_DSN"))
    host = get_config_value(config, "MYSQL_HOST", dsn_values.get("MYSQL_HOST", "127.0.0.1"))
    port_text = get_config_value(config, "MYSQL_PORT", dsn_values.get("MYSQL_PORT", "3306"))
    user = get_config_value(config, "MYSQL_USER", dsn_values.get("MYSQL_USER", ""))
    password = get_config_value(config, "MYSQL_PASSWORD", dsn_values.get("MYSQL_PASSWORD", ""))
    database = get_config_value(config, "MYSQL_DB", dsn_values.get("MYSQL_DB", ""))
    charset = get_config_value(config, "MYSQL_CHARSET", dsn_values.get("MYSQL_CHARSET", "utf8mb4")) or "utf8mb4"

    if not host or not user or not database:
        raise SystemExit("写入数据库失败：MYSQL_HOST / MYSQL_USER / MYSQL_DB 配置不完整。")
    try:
        port = int(port_text)
    except ValueError as exc:
        raise SystemExit(f"写入数据库失败：MYSQL_PORT 非法: {port_text}") from exc

    return DatabaseConfig(
        driver="mysql",
        mysql_host=host,
        mysql_port=port,
        mysql_user=user,
        mysql_password=password,
        mysql_db=database,
        mysql_charset=charset,
    )


def upsert_system_setting_sqlite(
    sqlite_path: Path,
    *,
    key: str,
    value: str,
    description: str,
) -> None:
    sqlite_path.parent.mkdir(parents=True, exist_ok=True)
    connection = sqlite3.connect(str(sqlite_path))
    try:
        cursor = connection.cursor()
        now = build_now_string()
        cursor.execute(
            'UPDATE system_settings SET value = ?, updated_at = ? WHERE "key" = ?',
            (value, now, key),
        )
        if cursor.rowcount == 0:
            cursor.execute(
                'INSERT INTO system_settings ("key", value, description, updated_at) VALUES (?, ?, ?, ?)',
                (key, value, description, now),
            )
        connection.commit()
    finally:
        connection.close()


def upsert_system_setting_mysql(
    db_config: DatabaseConfig,
    *,
    key: str,
    value: str,
    description: str,
) -> None:
    pymysql = import_optional_module("pymysql")
    if pymysql is None:
        upsert_system_setting_mysql_via_cli(
            db_config,
            key=key,
            value=value,
            description=description,
        )
        return

    connection = pymysql.connect(
        host=db_config.mysql_host,
        port=db_config.mysql_port,
        user=db_config.mysql_user,
        password=db_config.mysql_password,
        database=db_config.mysql_db,
        charset=db_config.mysql_charset,
        autocommit=True,
    )
    try:
        with connection.cursor() as cursor:
            now = build_now_string()
            cursor.execute(
                "UPDATE system_settings SET value = %s, updated_at = %s WHERE `key` = %s",
                (value, now, key),
            )
            if cursor.rowcount == 0:
                cursor.execute(
                    "INSERT INTO system_settings (`key`, value, description, updated_at) VALUES (%s, %s, %s, %s)",
                    (key, value, description, now),
                )
    finally:
        connection.close()


def quote_mysql_literal(value: str) -> str:
    escaped = value.replace("\\", "\\\\").replace("'", "\\'")
    return f"'{escaped}'"


def upsert_system_setting_mysql_via_cli(
    db_config: DatabaseConfig,
    *,
    key: str,
    value: str,
    description: str,
) -> None:
    mysql_binary = shutil.which("mysql")
    if not mysql_binary:
        raise SystemExit(
            "写入 MySQL 失败：既未安装 PyMySQL，也未找到 mysql 客户端。"
            " 请先执行 `pip install -e . --no-build-isolation`，或确认系统安装了 mysql 命令。"
        )

    now = build_now_string()
    sql = " ".join(
        [
            "UPDATE system_settings",
            f"SET value = {quote_mysql_literal(value)}, updated_at = {quote_mysql_literal(now)}",
            f"WHERE `key` = {quote_mysql_literal(key)};",
            "INSERT INTO system_settings (`key`, value, description, updated_at)",
            "SELECT",
            quote_mysql_literal(key) + ",",
            quote_mysql_literal(value) + ",",
            quote_mysql_literal(description) + ",",
            quote_mysql_literal(now),
            "FROM DUAL",
            "WHERE NOT EXISTS (",
            "SELECT 1 FROM system_settings",
            f"WHERE `key` = {quote_mysql_literal(key)}",
            ");",
        ]
    )
    env = os.environ.copy()
    env["MYSQL_PWD"] = db_config.mysql_password
    result = subprocess.run(
        [
            mysql_binary,
            "-h",
            db_config.mysql_host,
            "-P",
            str(db_config.mysql_port),
            "-u",
            db_config.mysql_user,
            db_config.mysql_db,
            f"--default-character-set={db_config.mysql_charset}",
            "-e",
            sql,
        ],
        capture_output=True,
        text=True,
        check=False,
        env=env,
    )
    if result.returncode != 0:
        detail = (result.stderr or result.stdout or "unknown error").strip()
        raise SystemExit(f"写入 MySQL 失败：{detail}")


def write_token_to_database_if_needed(
    config: RuntimeConfig,
    args: argparse.Namespace,
    token: str,
) -> int:
    if not args.write_db:
        return 0

    db_config = load_database_config(config)
    try:
        if db_config.driver == "sqlite":
            assert db_config.sqlite_path is not None
            upsert_system_setting_sqlite(
                db_config.sqlite_path,
                key=SYSTEM_SETTING_TOKEN_KEY,
                value=token,
                description=SYSTEM_SETTING_TOKEN_DESCRIPTION,
            )
            print(f"已更新 SQLite system_settings.{SYSTEM_SETTING_TOKEN_KEY}: {db_config.sqlite_path}")
            return 0

        upsert_system_setting_mysql(
            db_config,
            key=SYSTEM_SETTING_TOKEN_KEY,
            value=token,
            description=SYSTEM_SETTING_TOKEN_DESCRIPTION,
        )
        print(
            "已更新 MySQL system_settings."
            f"{SYSTEM_SETTING_TOKEN_KEY}: {db_config.mysql_user}@{db_config.mysql_host}:{db_config.mysql_port}/{db_config.mysql_db}"
        )
        return 0
    except SystemExit:
        raise
    except Exception as exc:
        print(f"写入数据库失败：{exc}", file=sys.stderr)
        return 1


def extract_api_error_message(payload: dict[str, Any]) -> str:
    for key in ("info", "message", "msg", "error"):
        value = str(payload.get(key) or "").strip()
        if value:
            return value
    return "unknown error"


def is_non_retryable_login_error(message: str) -> bool:
    normalized = re.sub(r"\s+", "", message or "").lower()
    hard_fail_markers = (
        "用户名或密码错误",
        "账号已被禁用",
        "账号不存在",
        "userdoesnotexist",
        "invalidcredential",
        "invalidpassword",
        "wrongpassword",
        "disabled",
    )
    return any(marker in normalized for marker in hard_fail_markers)


def open_json_request(request: Request, *, action: str) -> dict[str, Any]:
    try:
        with urlopen(request, timeout=30) as response:
            payload = json.load(response)
    except HTTPError as exc:
        payload = parse_json_object(exc.read().decode("utf-8", errors="ignore"))
        message = extract_api_error_message(payload) or f"HTTP {exc.code}"
        raise MihuaApiError(f"{action}失败: {message}") from exc
    except URLError as exc:
        raise MihuaApiError(f"{action}失败: 网络错误 {exc.reason}") from exc
    except Exception as exc:
        raise MihuaApiError(f"{action}失败: {exc}") from exc

    if not isinstance(payload, dict):
        raise MihuaApiError(f"{action}失败: 返回内容不是有效 JSON")
    return payload


def fetch_captcha_via_api(config: RuntimeConfig) -> tuple[str, bytes]:
    timestamp = str(int(time.time()))
    nonce = build_random_nonce()
    request = Request(
        f"{config.api_base_url}/captcha",
        headers={
            "source": "client.web",
            "timestamp": timestamp,
            "auth-type": "token",
            "nonce": nonce,
            "accept": "application/json, text/plain, */*",
            "origin": config.origin,
            "referer": config.origin.rstrip("/") + "/",
            "user-agent": "Mozilla/5.0",
        },
    )
    payload = open_json_request(request, action="获取验证码")

    if payload.get("code") != 200:
        raise MihuaApiError(f"获取验证码失败: {extract_api_error_message(payload)}")

    data = payload.get("data") or {}
    session_id = str(data.get("session_id") or "").strip()
    image_uri = str(data.get("image") or "").strip()
    image_bytes = decode_data_uri_image(image_uri)
    if not session_id or not image_bytes:
        raise MihuaApiError("获取验证码失败：返回中缺少 session_id 或 image。")
    return session_id, image_bytes


def try_resolve_captcha_code(
    image_bytes: bytes,
    args: argparse.Namespace,
) -> tuple[str, str]:
    if args.captcha:
        return normalize_ocr_text(args.captcha), "manual"

    candidates: list[tuple[str, str]] = []
    for engine, solver in (
        ("ddddocr", solve_captcha_with_ddddocr),
        ("tesseract", solve_captcha_with_tesseract),
    ):
        result = solver(image_bytes)
        candidates.append((engine, result))
        if len(result) >= DEFAULT_CAPTCHA_LENGTH:
            return result, engine
    for engine, result in candidates:
        if result:
            return result, engine
    return "", "unavailable"


def login_via_api(
    config: RuntimeConfig,
    *,
    username: str,
    password: str,
    session_id: str,
    verify_code: str,
) -> tuple[str, str]:
    username_header = f"{username}@{config.origin.split('//', 1)[-1].strip('/')}"
    timestamp = str(int(time.time()))
    nonce = build_random_nonce()
    password_sha = hashlib.sha256(password.encode("utf-8")).hexdigest()
    signature_key = hashlib.sha256(password_sha.encode("utf-8")).hexdigest()
    sign_text = (
        f"auth-type=signature&nonce={nonce}&session_id={session_id}"
        f"&source=client.web&timestamp={timestamp}&username={username_header}"
        f"&verify_code={verify_code}&signatureKey={signature_key}"
    )
    signature = hashlib.sha256(sign_text.encode("utf-8")).hexdigest().upper()
    payload = json.dumps(
        {"session_id": session_id, "verify_code": verify_code}
    ).encode("utf-8")
    request = Request(
        f"{config.api_base_url}/user/token",
        data=payload,
        headers={
            "source": "client.web",
            "nonce": nonce,
            "auth-type": "signature",
            "timestamp": timestamp,
            "username": username_header,
            "signature": signature,
            "content-type": "application/json;charset=UTF-8",
            "accept": "application/json, text/plain, */*",
            "origin": config.origin,
            "referer": config.origin.rstrip("/") + "/",
            "user-agent": "Mozilla/5.0",
        },
        method="POST",
    )
    payload = open_json_request(request, action="米话登录")

    if payload.get("code") != 200:
        message = extract_api_error_message(payload)
        raise MihuaApiError(f"米话登录失败: {message}", retryable=not is_non_retryable_login_error(message))

    token = str(((payload.get("data") or {}).get("token")) or "").strip()
    if not token:
        raise MihuaApiError("米话登录失败：响应中缺少 token。")
    return token, username_header


def write_token_to_env_if_needed(
    config: RuntimeConfig,
    args: argparse.Namespace,
    token: str,
) -> int:
    if not args.write_env:
        return 0
    if config.env_file is None:
        print("无法回写 env：没有定位到 env 文件。", file=sys.stderr)
        return 1
    write_env_value(config.env_file, TOKEN_ENV_KEY, token)
    print(f"已更新 {config.env_file} 中的 {TOKEN_ENV_KEY}")
    return 0


def persist_token_if_needed(
    config: RuntimeConfig,
    args: argparse.Namespace,
    token: str,
) -> int:
    env_result = write_token_to_env_if_needed(config, args, token)
    if env_result != 0:
        return env_result
    return write_token_to_database_if_needed(config, args, token)


def run_api_capture(args: argparse.Namespace, config: RuntimeConfig) -> int:
    retries = max(args.auto_login_retries, 1)
    captcha_path = build_captcha_output_path(args)
    last_error = "未拿到 token"
    manual_captcha = bool((args.captcha or "").strip())

    for attempt in range(1, retries + 1):
        try:
            session_id, image_bytes = fetch_captcha_via_api(config)
            write_captcha_image(image_bytes, captcha_path)

            captcha_code, engine = try_resolve_captcha_code(image_bytes, args)
            if len(captcha_code) < DEFAULT_CAPTCHA_LENGTH:
                last_error = f"验证码识别失败，已保存图片到 {captcha_path}"
                print(
                    f"第 {attempt} 次 API 登录：验证码识别失败，识别引擎 {engine}。",
                    file=sys.stderr,
                )
                if manual_captcha:
                    break
                continue

            token, username_header = login_via_api(
                config,
                username=args.username,
                password=args.password,
                session_id=session_id,
                verify_code=captcha_code,
            )
            captured = LocatedToken(token, "api_login", f"{username_header} via {config.api_base_url}/user/token")
            print_capture(captured, print_token=args.print_token)
            print(f"验证码图片: {captcha_path}")
            print(f"验证码识别引擎: {engine}")
            return persist_token_if_needed(config, args, token)
        except MihuaApiError as exc:
            last_error = str(exc)
            print(f"第 {attempt} 次 API 登录失败: {exc}", file=sys.stderr)
            if manual_captcha or not exc.retryable:
                break
            time.sleep(0.8)

    print(f"直连 API 未获取到 token：{last_error}", file=sys.stderr)
    if captcha_path.exists():
        print(f"最后一次验证码图片已保存到: {captcha_path}", file=sys.stderr)
    return 1


def refresh_captcha(page: Any) -> None:
    try:
        page.evaluate(
            """() => {
                const images = [...document.querySelectorAll('img')];
                const target = images.find((img) => {
                    return img.clientWidth >= 80 &&
                        img.clientWidth <= 140 &&
                        img.clientHeight >= 30 &&
                        img.clientHeight <= 60;
                });
                if (target) {
                    target.click();
                    return;
                }
                const switcher = [...document.querySelectorAll('*')].find((el) => {
                    return (el.innerText || '').trim() === '换一张';
                });
                if (switcher) {
                    switcher.click();
                }
            }"""
        )
    except Exception:
        return


def is_login_page_url(url: str) -> bool:
    normalized = (url or "").lower()
    return "/login" in normalized or "#/login" in normalized


def has_login_form(page: Any) -> bool:
    selector_groups = (
        USERNAME_INPUT_SELECTORS,
        PASSWORD_INPUT_SELECTORS,
        CAPTCHA_INPUT_SELECTORS,
    )
    for selectors in selector_groups:
        for selector in selectors:
            try:
                if page.locator(selector).first.is_visible():
                    return True
            except Exception:
                continue
    return False


def contains_login_error_text(body_text: str) -> bool:
    normalized = re.sub(r"\s+", "", body_text or "")
    error_markers = (
        "验证码错误",
        "验证码已过期",
        "验证码错误次数过多",
        "用户名或密码错误",
        "账号已被禁用",
        "登录失败",
    )
    return any(marker in normalized for marker in error_markers)


def wait_for_login_result(page: Any, timeout_ms: int) -> bool:
    deadline = time.monotonic() + max(timeout_ms, 1000) / 1000
    while time.monotonic() < deadline:
        try:
            if not is_login_page_url(page.url) and not has_login_form(page):
                return True
            body_text = str(page.evaluate("() => document.body ? document.body.innerText : ''") or "")
            if contains_login_error_text(body_text):
                return False
        except Exception:
            pass
        page.wait_for_timeout(400)
    return not is_login_page_url(page.url) and not has_login_form(page)


def fill_first_matching_input(page: Any, selectors: tuple[str, ...], value: str) -> bool:
    for selector in selectors:
        try:
            locator = page.locator(selector).first
            locator.wait_for(state="visible", timeout=1200)
            locator.fill(value)
            return True
        except Exception:
            continue
    return False


def fill_login_form(page: Any, *, username: str, password: str, captcha_text: str) -> bool:
    username_ok = fill_first_matching_input(page, USERNAME_INPUT_SELECTORS, username)
    password_ok = fill_first_matching_input(page, PASSWORD_INPUT_SELECTORS, password)
    captcha_ok = fill_first_matching_input(page, CAPTCHA_INPUT_SELECTORS, captcha_text)
    return username_ok and password_ok and captcha_ok


def click_login_button(page: Any) -> bool:
    for selector in LOGIN_BUTTON_SELECTORS:
        try:
            locator = page.locator(selector).first
            locator.wait_for(state="visible", timeout=1200)
            locator.click()
            return True
        except Exception:
            continue

    try:
        return bool(
            page.evaluate(
                """() => {
                    const clickable = [...document.querySelectorAll('button, [role="button"], input[type="submit"], a, div, span')];
                    const target = clickable.find((el) => {
                        const text = String(el.innerText || el.value || '').replace(/\\s+/g, '');
                        return text.includes('登录');
                    });
                    if (!target) {
                        return false;
                    }
                    target.click();
                    return true;
                }"""
            )
        )
    except Exception:
        return False


def attempt_auto_login(page: Any, args: argparse.Namespace) -> None:
    if not args.username or not args.password:
        return
    if not is_login_page_url(page.url) and not has_login_form(page):
        return

    retries = max(args.auto_login_retries, 1)
    manual_captcha = bool((args.captcha or "").strip())
    captcha_path = build_captcha_output_path(args)
    last_error = "验证码识别未通过"

    for attempt in range(1, retries + 1):
        image_bytes = decode_data_uri_image(extract_captcha_data_uri(page))
        if image_bytes:
            write_captcha_image(image_bytes, captcha_path)
        captcha_text, engine = try_resolve_captcha_code(image_bytes, args)

        if len(captcha_text) < DEFAULT_CAPTCHA_LENGTH:
            last_error = f"验证码识别失败，识别引擎 {engine}"
            print(
                f"第 {attempt} 次浏览器自动登录：验证码识别失败，识别引擎 {engine}。",
                file=sys.stderr,
            )
            if manual_captcha:
                break
            refresh_captcha(page)
            page.wait_for_timeout(800)
            continue

        if not fill_login_form(page, username=args.username, password=args.password, captcha_text=captcha_text):
            raise SystemExit("自动登录失败：未找到完整的登录表单。")
        if not click_login_button(page):
            raise SystemExit("自动登录失败：未找到登录按钮。")

        if wait_for_login_result(page, 6000):
            print(f"自动登录成功，验证码识别结果: {captcha_text}，识别引擎: {engine}")
            return

        last_error = f"第 {attempt} 次自动登录未成功，最后一次识别结果 {captcha_text} ({engine})"
        print(f"第 {attempt} 次自动登录未成功，准备刷新验证码重试。", file=sys.stderr)
        if manual_captcha:
            break
        refresh_captcha(page)
        page.wait_for_timeout(1000)

    message = f"自动登录失败：{last_error}。"
    if captcha_path.exists():
        message += f" 当前验证码图片已保存到 {captcha_path}。"
    raise SystemExit(message)


def try_store_capture(current: LocatedToken | None, candidate: LocatedToken | None) -> LocatedToken | None:
    if current is not None:
        return current
    if candidate is None:
        return None
    return candidate


def try_capture_from_request(request: Any, allowed_hosts: set[str]) -> LocatedToken | None:
    if not is_url_host_allowed(request.url, allowed_hosts):
        return None
    try:
        headers = request.all_headers()
    except Exception:
        headers = getattr(request, "headers", {})
    return extract_token_from_headers(headers, detail=f"request {request.method} {request.url}")


def try_capture_from_response(response: Any, allowed_hosts: set[str]) -> LocatedToken | None:
    if not is_url_host_allowed(response.url, allowed_hosts):
        return None
    try:
        headers = response.all_headers()
    except Exception:
        headers = getattr(response, "headers", {})
    captured = extract_token_from_headers(headers, detail=f"response {response.status} {response.url}")
    if captured:
        return captured

    content_type = str(headers.get("content-type", "")).lower()
    if "json" not in content_type:
        return None
    try:
        text = response.text()
    except Exception:
        return None
    return extract_token_from_json_text(text, detail=f"response {response.status} {response.url}")


def try_capture_from_storage(page: Any) -> LocatedToken | None:
    snapshot = capture_storage_snapshot(page)
    local_storage = snapshot.get("localStorage") or {}
    session_storage = snapshot.get("sessionStorage") or {}
    captured = extract_token_from_storage(local_storage, bucket="localStorage")
    if captured:
        return captured
    captured = extract_token_from_storage(session_storage, bucket="sessionStorage")
    if captured:
        return captured
    return None


def try_capture_from_cookies(context: Any, url: str) -> LocatedToken | None:
    try:
        cookies = context.cookies([url])
    except Exception:
        return None
    return extract_token_from_cookies(cookies, detail=url)


def print_intro(config: RuntimeConfig, args: argparse.Namespace) -> None:
    print("准备启动浏览器抓取米话 token。")
    print(f"打开地址: {config.open_url}")
    print(f"目标 origin: {config.origin}")
    print(f"允许监听 host: {', '.join(sorted(config.allowed_hosts))}")
    if args.username and args.password:
        print("已启用账号密码模式。")
    if args.write_env:
        print(f"抓取成功后将回写: {config.env_file}")
    if args.write_db:
        print("抓取成功后将回写数据库: system_settings.mihua_call_record_token")
    if args.username and args.password and not args.force_browser:
        print(f"优先走直连 API，失败后自动回退浏览器登录: {config.api_base_url}")
    elif not (args.username and args.password):
        print("请在打开的浏览器中完成米话登录。")
    print()


def print_capture(token: LocatedToken, *, print_token: bool) -> None:
    print("已抓到米话 token。")
    print(f"来源: {token.source}")
    print(f"细节: {token.detail}")
    print(f"掩码: {mask_token(token.value)}")
    print(f"时间: {datetime.now().isoformat(timespec='seconds')}")
    if print_token:
        print("TOKEN_BEGIN")
        print(token.value)
        print("TOKEN_END")


def run_capture(args: argparse.Namespace, config: RuntimeConfig) -> int:
    sync_playwright = import_playwright()
    print_intro(config, args)

    captured: LocatedToken | None = None
    deadline = time.monotonic() + max(args.timeout_seconds, 1)
    poll_interval_ms = max(args.poll_interval_ms, 200)

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(
            headless=args.headless,
            channel=args.browser_channel or None,
        )
        context = browser.new_context(ignore_https_errors=True)
        page = context.new_page()

        def on_request(request: Any) -> None:
            nonlocal captured
            captured = try_store_capture(captured, try_capture_from_request(request, config.allowed_hosts))

        def on_response(response: Any) -> None:
            nonlocal captured
            captured = try_store_capture(captured, try_capture_from_response(response, config.allowed_hosts))

        context.on("request", on_request)
        context.on("response", on_response)

        page.goto(config.open_url, wait_until="domcontentloaded")
        attempt_auto_login(page, args)
        captured = try_store_capture(captured, try_capture_from_storage(page))
        captured = try_store_capture(captured, try_capture_from_cookies(context, page.url))

        while time.monotonic() < deadline and captured is None:
            page.wait_for_timeout(poll_interval_ms)
            captured = try_store_capture(captured, try_capture_from_storage(page))
            captured = try_store_capture(captured, try_capture_from_cookies(context, page.url))

        if captured is None:
            browser.close()
            print("超时：在指定时间内没有抓到 token。", file=sys.stderr)
            print("建议：确认登录后是否已进入业务页，并手动点击一次列表或录音相关页面触发接口请求。", file=sys.stderr)
            return 1

        print_capture(captured, print_token=args.print_token)

        write_result = persist_token_if_needed(config, args, captured.value)
        if write_result != 0:
            browser.close()
            return write_result

        browser.close()
        return 0


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    config = load_runtime_config(args)
    if args.username and args.password and not args.force_browser:
        api_result = run_api_capture(args, config)
        if api_result == 0:
            return 0
        print("直连 API 未获取到 token，自动回退到浏览器登录模式。", file=sys.stderr)
    return run_capture(args, config)
