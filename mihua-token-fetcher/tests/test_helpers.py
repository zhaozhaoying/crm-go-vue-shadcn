from pathlib import Path
import sqlite3
import sys
import tempfile
import unittest
from types import SimpleNamespace
from unittest.mock import patch

TESTS_DIR = Path(__file__).resolve().parent
PROJECT_ROOT = TESTS_DIR.parent
SRC_DIR = PROJECT_ROOT / "src"
if str(SRC_DIR) not in sys.path:
    sys.path.insert(0, str(SRC_DIR))

from mihua_token_fetcher import cli
from mihua_token_fetcher.helpers import (
    collect_allowed_hosts,
    extract_token_from_headers,
    extract_token_from_json_text,
    extract_token_from_storage,
    find_repo_root,
    upsert_env_text,
)


class HelpersTestCase(unittest.TestCase):
    def test_extract_token_from_headers(self) -> None:
        located = extract_token_from_headers(
            {"Token": "abc12345xyz"},
            detail="request GET https://example.com/api",
        )
        self.assertIsNotNone(located)
        self.assertEqual(located.value, "abc12345xyz")
        self.assertIn("request GET", located.detail)

    def test_extract_token_from_json_text(self) -> None:
        located = extract_token_from_json_text(
            '{"code":200,"data":{"token":"mihua-token-001"}}',
            detail="response 200",
        )
        self.assertIsNotNone(located)
        self.assertEqual(located.value, "mihua-token-001")
        self.assertEqual(located.source, "response_json")

    def test_extract_token_from_storage_prefers_explicit_keys(self) -> None:
        located = extract_token_from_storage(
            {
                "profile": "user-12345678",
                "token": "mihua-storage-token",
                "auth_state": "ignore-me",
            },
            bucket="localStorage",
        )
        self.assertIsNotNone(located)
        self.assertEqual(located.value, "mihua-storage-token")
        self.assertEqual(located.detail, "localStorage:token")

    def test_upsert_env_text_updates_existing_key(self) -> None:
        updated = upsert_env_text(
            "FOO=bar\nMIHUA_CALL_RECORD_TOKEN=old-value\n",
            "MIHUA_CALL_RECORD_TOKEN",
            "new-value",
        )
        self.assertEqual(updated, "FOO=bar\nMIHUA_CALL_RECORD_TOKEN=new-value\n")

    def test_collect_allowed_hosts(self) -> None:
        hosts = collect_allowed_hosts(
            [
                "https://spxxjj.emicloudcc.com",
                "https://cmb.emicloudcc.com/api/v1/list",
                "",
            ]
        )
        self.assertEqual(hosts, {"spxxjj.emicloudcc.com", "cmb.emicloudcc.com"})

    def test_find_repo_root(self) -> None:
        repo_root = find_repo_root(Path.cwd())
        self.assertIsNotNone(repo_root)
        self.assertTrue((repo_root / "backend").exists())


class CliTestCase(unittest.TestCase):
    def test_locate_default_env_file_uses_repo_root_env(self) -> None:
        repo_root = find_repo_root(Path.cwd())
        self.assertIsNotNone(repo_root)
        env_file = cli.locate_default_env_file(repo_root)
        self.assertEqual(env_file, repo_root / ".env")

    def test_normalize_ocr_text_keeps_alnum_and_clips_to_4_chars(self) -> None:
        self.assertEqual(cli.normalize_ocr_text(" A-b_12X "), "Ab12")

    def test_contains_login_error_text_does_not_misjudge_login_labels(self) -> None:
        self.assertFalse(cli.contains_login_error_text("用户名 密码 验证码 登录"))
        self.assertTrue(cli.contains_login_error_text("用户名或密码错误"))

    def test_is_login_page_url_supports_hash_route(self) -> None:
        self.assertTrue(cli.is_login_page_url("https://spxxjj.emicloudcc.com/#/login"))
        self.assertFalse(cli.is_login_page_url("https://spxxjj.emicloudcc.com/#/dashboard"))

    def test_try_resolve_captcha_code_prefers_manual_value(self) -> None:
        code, engine = cli.try_resolve_captcha_code(
            b"",
            SimpleNamespace(captcha=" A-b_12 "),
        )
        self.assertEqual(code, "Ab12")
        self.assertEqual(engine, "manual")

    def test_build_captcha_output_path_prefers_explicit_path(self) -> None:
        path = cli.build_captcha_output_path(
            SimpleNamespace(captcha_image_path="./captchas/current.jpg"),
        )
        self.assertEqual(path, (Path("./captchas/current.jpg")).expanduser().resolve())

    @patch("mihua_token_fetcher.cli.is_directory_writable", side_effect=[False, True])
    @patch("mihua_token_fetcher.cli.Path.home", return_value=Path("/home/tester"))
    @patch("mihua_token_fetcher.cli.Path.cwd", return_value=Path("/work/tool"))
    @patch("mihua_token_fetcher.cli.tempfile.gettempdir", return_value="/tmp/restricted")
    def test_build_captcha_output_path_falls_back_when_temp_unwritable(
        self,
        _tempdir_mock,
        _cwd_mock,
        _home_mock,
        _writable_mock,
    ) -> None:
        path = cli.build_captcha_output_path(SimpleNamespace(captcha_image_path=""))
        self.assertEqual(path, (Path("/work/tool/.cache/mihua-captcha-latest.jpg")).resolve())

    @patch("mihua_token_fetcher.cli.solve_captcha_with_ddddocr", return_value="A1B2")
    @patch("mihua_token_fetcher.cli.solve_captcha_with_tesseract")
    def test_try_resolve_captcha_code_prefers_ddddocr_before_tesseract(self, tesseract_mock, *_args) -> None:
        code, engine = cli.try_resolve_captcha_code(
            b"image-bytes",
            SimpleNamespace(captcha=""),
        )
        self.assertEqual(code, "A1B2")
        self.assertEqual(engine, "ddddocr")
        tesseract_mock.assert_not_called()

    @patch("mihua_token_fetcher.cli.subprocess.run", side_effect=FileNotFoundError)
    def test_solve_captcha_with_tesseract_returns_empty_when_binary_missing(self, _run_mock) -> None:
        self.assertEqual(cli.solve_captcha_with_tesseract(b"image-bytes"), "")

    def test_parse_mysql_dsn(self) -> None:
        parsed = cli.parse_mysql_dsn(
            "root:secret@tcp(127.0.0.1:3306)/crm?charset=utf8mb4&parseTime=True&loc=Local"
        )
        self.assertEqual(parsed["MYSQL_USER"], "root")
        self.assertEqual(parsed["MYSQL_PASSWORD"], "secret")
        self.assertEqual(parsed["MYSQL_HOST"], "127.0.0.1")
        self.assertEqual(parsed["MYSQL_PORT"], "3306")
        self.assertEqual(parsed["MYSQL_DB"], "crm")
        self.assertEqual(parsed["MYSQL_CHARSET"], "utf8mb4")

    def test_upsert_system_setting_sqlite(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "test.db"
            connection = sqlite3.connect(str(db_path))
            try:
                connection.execute(
                    """
                    CREATE TABLE system_settings (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        key TEXT NOT NULL UNIQUE,
                        value TEXT NOT NULL,
                        description TEXT NOT NULL DEFAULT '',
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                    )
                    """
                )
                connection.commit()
            finally:
                connection.close()

            cli.upsert_system_setting_sqlite(
                db_path,
                key=cli.SYSTEM_SETTING_TOKEN_KEY,
                value="token-1",
                description=cli.SYSTEM_SETTING_TOKEN_DESCRIPTION,
            )
            cli.upsert_system_setting_sqlite(
                db_path,
                key=cli.SYSTEM_SETTING_TOKEN_KEY,
                value="token-2",
                description=cli.SYSTEM_SETTING_TOKEN_DESCRIPTION,
            )

            connection = sqlite3.connect(str(db_path))
            try:
                row = connection.execute(
                    'SELECT key, value, description FROM system_settings WHERE "key" = ?',
                    (cli.SYSTEM_SETTING_TOKEN_KEY,),
                ).fetchone()
            finally:
                connection.close()

            self.assertEqual(
                row,
                (
                    cli.SYSTEM_SETTING_TOKEN_KEY,
                    "token-2",
                    cli.SYSTEM_SETTING_TOKEN_DESCRIPTION,
                ),
            )


if __name__ == "__main__":
    unittest.main()
