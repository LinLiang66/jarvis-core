"""开放平台 Python 客户端：握手 + 3DES 业务调用。"""

from __future__ import annotations

import json
import time
from typing import Any
from urllib.parse import quote_plus, urlencode

import requests

from openplatform.crypto import rsa_util, sign, triple_des
from openplatform.exception import OpenPlatformError

ACTION_GET_PUBLIC_KEY = "open.session.publickey"
ACTION_CREATE_SECRET_KEY = "microSession.create.secretkey"
ACTION_ECHO = "open.demo.echo"

CODE_SUCCESS = 200
CODE_TOKEN_INVALID = 40001
CODE_QUOTA_EXCEEDED = 40002

_APP_VER = "1.0.0"
_VERSION = "V1.0"


class OpenPlatformClient:
    def __init__(
        self,
        gateway_url: str,
        app_id: str,
        sign_secret: str,
        app_secret_base64: str,
        timeout: float = 30.0,
    ) -> None:
        self.gateway_url = gateway_url
        self.app_id = app_id
        self.sign_secret = sign_secret
        self.private_key = rsa_util.load_private_key_from_der_base64(app_secret_base64)
        self.timeout = timeout
        self.token: str | None = None
        self._tdes: triple_des.TripleDesCipher | None = None

    def get_public_key_and_token(self) -> str:
        params = self._base_params(ACTION_GET_PUBLIC_KEY)
        params["timestamp"] = str(int(time.time() * 1000))
        params["data"] = "{}"
        body = self._call_gateway(params)
        self.token = body["token"]
        return self.token

    def init_3des_key(self) -> None:
        if not self.token:
            raise RuntimeError("call get_public_key_and_token() first")
        random_num = triple_des.random_digits(12)
        encrypted = rsa_util.encrypt_by_private_key(self.private_key, random_num)
        params = self._base_params(ACTION_CREATE_SECRET_KEY)
        params["req_time"] = str(int(time.time() * 1000))
        params["token"] = self.token
        params["data"] = quote_plus(encrypted)
        body = self._call_gateway(params)
        server_part = rsa_util.decrypt_by_private_key(self.private_key, body["serverPart"])
        final_key = random_num + server_part
        self._tdes = triple_des.TripleDesCipher(final_key)

    def call_encrypted(self, action: str, json_plain: str) -> str:
        if self._tdes is None:
            raise RuntimeError("call init_3des_key() first")
        cipher_data = self._tdes.encrypt(json_plain.encode("utf-8"))
        params = self._base_params(action)
        params["req_time"] = str(int(time.time() * 1000))
        params["token"] = self.token or ""
        params["data"] = cipher_data
        body = self._call_gateway_raw(params)
        if isinstance(body, str):
            cipher_body = body
        else:
            cipher_body = json.dumps(body, ensure_ascii=False, separators=(",", ":"))
        plain = self._tdes.decrypt(cipher_body)
        return plain.decode("utf-8")

    def _base_params(self, action: str) -> dict[str, str]:
        return {
            "action": action,
            "appid": self.app_id,
            "appver": _APP_VER,
            "version": _VERSION,
            "sign_method": sign.SIGN_METHOD,
        }

    def _call_gateway(self, params: dict[str, str]) -> dict[str, Any]:
        body = self._call_gateway_raw(params)
        if not isinstance(body, dict):
            raise OpenPlatformError(CODE_SUCCESS, f"unexpected body: {body!r}")
        return body

    def _call_gateway_raw(self, params: dict[str, str]) -> Any:
        signed = dict(params)
        signed["sign"] = sign.build_sign(signed, self.sign_secret)
        response = requests.post(
            self.gateway_url,
            data=urlencode(signed),
            headers={"Content-Type": "application/x-www-form-urlencoded"},
            timeout=self.timeout,
        )
        root = response.json()
        code = int(root["code"])
        if code == CODE_TOKEN_INVALID:
            msg = root.get("message") or "token invalid or expired"
            raise OpenPlatformError(CODE_TOKEN_INVALID, msg)
        if code == CODE_QUOTA_EXCEEDED:
            msg = root.get("message") or "quota exceeded"
            raise OpenPlatformError(CODE_QUOTA_EXCEEDED, msg)
        if code != CODE_SUCCESS:
            msg = root.get("message") or ""
            raise OpenPlatformError(code, f"gateway error: code={code}, message={msg}")
        data = root.get("data")
        if data is None:
            raise OpenPlatformError(CODE_SUCCESS, "empty response data")
        return data
