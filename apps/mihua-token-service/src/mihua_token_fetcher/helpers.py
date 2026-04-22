from __future__ import annotations

from dataclasses import dataclass
import json
import re
from pathlib import Path
from typing import Any, Iterable, Mapping
from urllib.parse import urlparse


PREFERRED_TOKEN_KEYS = (
    "token",
    "access_token",
    "accessToken",
    "auth_token",
    "authToken",
    "jwt",
)


@dataclass(slots=True)
class LocatedToken:
    value: str
    source: str
    detail: str


def normalize_token_value(raw: Any) -> str | None:
    if not isinstance(raw, str):
        return None
    value = raw.strip()
    if not value:
        return None
    lowered = value.lower()
    if lowered in {"null", "none", "undefined"}:
        return None
    if lowered.startswith("bearer "):
        value = value[7:].strip()
    if len(value) < 8:
        return None
    return value


def parse_env_text(content: str) -> dict[str, str]:
    result: dict[str, str] = {}
    for line in content.splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        result[key.strip()] = value.strip()
    return result


def load_env_file(path: Path | None) -> dict[str, str]:
    if path is None or not path.exists():
        return {}
    return parse_env_text(path.read_text(encoding="utf-8"))


def upsert_env_text(content: str, key: str, value: str) -> str:
    lines = content.splitlines()
    updated = False
    for index, line in enumerate(lines):
        if "=" not in line:
            continue
        current_key, _ = line.split("=", 1)
        if current_key.strip() != key:
            continue
        lines[index] = f"{key}={value}"
        updated = True
        break
    if not updated:
        if lines and lines[-1] != "":
            lines.append(f"{key}={value}")
        else:
            lines[-1:] = [f"{key}={value}"]
    trailing_newline = "\n" if content.endswith("\n") or not content else ""
    return "\n".join(lines) + trailing_newline


def write_env_value(path: Path, key: str, value: str) -> None:
    existing = ""
    if path.exists():
        existing = path.read_text(encoding="utf-8")
    updated = upsert_env_text(existing, key, value)
    path.write_text(updated, encoding="utf-8")


def extract_token_from_headers(headers: Mapping[str, str], *, detail: str) -> LocatedToken | None:
    for key in ("token", "x-token", "access-token", "authorization"):
        token = normalize_token_value(headers.get(key))
        if token:
            return LocatedToken(token, "headers", f"{detail} -> {key}")
    lowered = {str(key).lower(): value for key, value in headers.items()}
    for key in ("token", "x-token", "access-token", "authorization"):
        token = normalize_token_value(lowered.get(key))
        if token:
            return LocatedToken(token, "headers", f"{detail} -> {key}")
    return None


def extract_token_from_storage(storage: Mapping[str, str], *, bucket: str) -> LocatedToken | None:
    preferred: list[tuple[str, str]] = []
    fallback: list[tuple[str, str]] = []
    for key, raw_value in storage.items():
        token = normalize_token_value(raw_value)
        if not token:
            continue
        lowered = key.lower()
        if lowered in {item.lower() for item in PREFERRED_TOKEN_KEYS}:
            preferred.append((key, token))
            continue
        if "token" in lowered or "auth" in lowered or "jwt" in lowered:
            fallback.append((key, token))
    candidates = preferred or fallback
    if not candidates:
        return None
    key, token = candidates[0]
    return LocatedToken(token, "storage", f"{bucket}:{key}")


def extract_token_from_cookies(
    cookies: Iterable[Mapping[str, Any]],
    *,
    detail: str,
) -> LocatedToken | None:
    for cookie in cookies:
        name = str(cookie.get("name", ""))
        value = normalize_token_value(cookie.get("value"))
        if not value:
            continue
        lowered = name.lower()
        if "token" in lowered or "auth" in lowered or lowered in {"jwt", "session"}:
            return LocatedToken(value, "cookies", f"{detail}:{name}")
    return None


def _flatten_json_for_tokens(payload: Any, path: str = "$") -> Iterable[tuple[str, Any]]:
    if isinstance(payload, Mapping):
        for key, value in payload.items():
            next_path = f"{path}.{key}"
            yield next_path, value
            yield from _flatten_json_for_tokens(value, next_path)
        return
    if isinstance(payload, list):
        for index, value in enumerate(payload):
            next_path = f"{path}[{index}]"
            yield next_path, value
            yield from _flatten_json_for_tokens(value, next_path)


def extract_token_from_json_text(text: str, *, detail: str) -> LocatedToken | None:
    try:
        payload = json.loads(text)
    except json.JSONDecodeError:
        return None
    preferred: list[tuple[str, str]] = []
    fallback: list[tuple[str, str]] = []
    for path, value in _flatten_json_for_tokens(payload):
        token = normalize_token_value(value)
        if not token:
            continue
        lowered_path = path.lower()
        if any(pattern in lowered_path for pattern in (".token", "_token", "accesstoken", "access_token", "authtoken", "auth_token")):
            preferred.append((path, token))
            continue
        if "token" in lowered_path or "auth" in lowered_path or "jwt" in lowered_path:
            fallback.append((path, token))
    candidates = preferred or fallback
    if not candidates:
        return None
    path, token = candidates[0]
    return LocatedToken(token, "response_json", f"{detail}:{path}")


def is_url_host_allowed(raw_url: str, allowed_hosts: set[str]) -> bool:
    host = (urlparse(raw_url).hostname or "").lower()
    if not host:
        return False
    if not allowed_hosts:
        return True
    return host in allowed_hosts


def collect_allowed_hosts(values: Iterable[str]) -> set[str]:
    hosts: set[str] = set()
    for value in values:
        raw = (value or "").strip()
        if not raw:
            continue
        parsed = urlparse(raw)
        host = (parsed.hostname or "").lower()
        if host:
            hosts.add(host)
    return hosts


def find_repo_root(start: Path) -> Path | None:
    current = start.resolve()
    for candidate in (current, *current.parents):
        if (candidate / ".git").exists():
            return candidate
        if (candidate / "backend").exists() and (candidate / "frontend").exists():
            return candidate
    return None


def mask_token(token: str, visible: int = 4) -> str:
    value = token.strip()
    if len(value) <= visible * 2:
        return "*" * len(value)
    return f"{value[:visible]}...{value[-visible:]}"


def slugify_url(raw_url: str) -> str:
    host = (urlparse(raw_url).hostname or "").lower().strip()
    if not host:
        return "unknown"
    return re.sub(r"[^a-z0-9]+", "-", host).strip("-") or "unknown"

