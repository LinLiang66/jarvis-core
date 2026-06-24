"""开放平台网关业务异常。"""

from __future__ import annotations


class OpenPlatformError(Exception):
    def __init__(self, code: int, message: str) -> None:
        super().__init__(message)
        self.code = code
        self.message = message

    def is_token_invalid(self) -> bool:
        return self.code == 40001

    def is_quota_exceeded(self) -> bool:
        return self.code == 40002
