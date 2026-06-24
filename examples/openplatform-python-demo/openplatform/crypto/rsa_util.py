"""RSA/ECB/PKCS1Padding，与 Go/Java 服务端对齐。"""

from __future__ import annotations

import base64

from Crypto.Cipher import PKCS1_v1_5
from Crypto.PublicKey import RSA


def load_private_key_from_der_base64(app_secret_base64: str) -> RSA.RsaKey:
    der = base64.b64decode(app_secret_base64.strip())
    return RSA.import_key(der)


def encrypt_by_private_key(private_key: RSA.RsaKey, plain: str) -> str:
    """私钥加密（密钥交换时加密 clientRandom）。"""
    encrypted = _private_encrypt_pkcs1_type1(private_key, plain.encode("utf-8"))
    return base64.b64encode(encrypted).decode("ascii")


def decrypt_by_private_key(private_key: RSA.RsaKey, cipher_base64: str) -> str:
    """私钥解密（解密服务端公钥加密的 serverPart）。"""
    raw = base64.b64decode(cipher_base64)
    cipher = PKCS1_v1_5.new(private_key)
    plain = cipher.decrypt(raw, sentinel=None)
    if plain is None:
        raise ValueError("RSA decrypt failed")
    return plain.decode("utf-8")


def _private_encrypt_pkcs1_type1(key: RSA.RsaKey, message: bytes) -> bytes:
    k = key.size_in_bytes()
    if len(message) > k - 11:
        raise ValueError("message too long")
    padding_len = k - len(message) - 3
    em = b"\x00\x01" + (b"\xff" * padding_len) + b"\x00" + message
    m = pow(int.from_bytes(em, "big"), key.d, key.n)
    return m.to_bytes(k, "big")
