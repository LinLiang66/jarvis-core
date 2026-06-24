"""a2_md5 签名：参数 key 升序，拼接 key+value，末尾追加 signSecret，MD5 小写 hex。"""

from __future__ import annotations

import hashlib
from typing import Mapping

SIGN_METHOD = "a2_md5"


def build_sign(params: Mapping[str, str], sign_secret: str) -> str:
    items = sorted(
        (k, v)
        for k, v in params.items()
        if k != "sign" and v
    )
    raw = "".join(f"{k}{v}" for k, v in items) + sign_secret
    return hashlib.md5(raw.encode("utf-8")).hexdigest()
