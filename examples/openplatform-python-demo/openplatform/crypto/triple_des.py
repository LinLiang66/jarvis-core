"""3DES/ECB/PKCS5Padding，与会话密钥交换后的 finalKey 一致。"""

from __future__ import annotations

import base64
import secrets

from Crypto.Cipher import DES3


def random_digits(length: int) -> str:
    return "".join(str(secrets.randbelow(10)) for _ in range(length))


def normalize_key24(key: bytes) -> bytes:
    if len(key) == 16:
        return key + key[:8]
    if len(key) == 24:
        return key
    if len(key) > 24:
        return key[:24]
    return key.ljust(24, b"\x00")


class TripleDesCipher:
    def __init__(self, session_key: str) -> None:
        self._key = normalize_key24(session_key.encode("utf-8"))

    def encrypt(self, plain: bytes) -> str:
        cipher = DES3.new(self._key, DES3.MODE_ECB)
        padded = _pkcs5_pad(plain, cipher.block_size)
        encrypted = cipher.encrypt(padded)
        return base64.b64encode(encrypted).decode("ascii")

    def decrypt(self, cipher_base64: str) -> bytes:
        cipher = DES3.new(self._key, DES3.MODE_ECB)
        raw = base64.b64decode(cipher_base64)
        padded = cipher.decrypt(raw)
        return _pkcs5_unpad(padded)


def _pkcs5_pad(data: bytes, block_size: int) -> bytes:
    n = block_size - len(data) % block_size
    return data + bytes([n] * n)


def _pkcs5_unpad(data: bytes) -> bytes:
    if not data:
        raise ValueError("empty data")
    n = data[-1]
    if n <= 0 or n > len(data):
        raise ValueError("invalid padding")
    if data[-n:] != bytes([n] * n):
        raise ValueError("invalid padding bytes")
    return data[:-n]
