#!/usr/bin/env python3
"""Open platform demo: handshake + encrypted Echo (open.demo.echo).

Create an open-platform app in admin UI first, then:

    pip install -r requirements.txt
    python demo.py \\
      --gateway http://127.0.0.1:8000/api/v1/open/gateway \\
      --appid app_xxx --sign signSecret --secret appSecretBase64

Env: OPEN_GATEWAY_URL, OPEN_APP_ID, OPEN_SIGN_SECRET, OPEN_APP_SECRET
"""

from __future__ import annotations

import argparse
import json
import os
import sys

from openplatform.client import ACTION_ECHO, OpenPlatformClient


def env_or(key: str, default: str = "") -> str:
    return os.environ.get(key, default)


def main() -> None:
    parser = argparse.ArgumentParser(description="Open platform Python demo (Echo only)")
    parser.add_argument(
        "--gateway",
        default=env_or("OPEN_GATEWAY_URL", "http://127.0.0.1:8000/api/v1/open/gateway"),
    )
    parser.add_argument("--appid", default=env_or("OPEN_APP_ID"))
    parser.add_argument("--sign", default=env_or("OPEN_SIGN_SECRET"))
    parser.add_argument("--secret", default=env_or("OPEN_APP_SECRET"))
    args = parser.parse_args()

    if not args.appid or not args.sign or not args.secret:
        parser.print_help()
        sys.exit(1)

    client = OpenPlatformClient(args.gateway, args.appid, args.sign, args.secret)

    token = client.get_public_key_and_token()
    print("token:", token)

    client.init_3des_key()
    print("3des session ready")

    echo_payload = json.dumps({"hello": "jarvis demo"}, ensure_ascii=False)
    echo_response = client.call_encrypted(ACTION_ECHO, echo_payload)
    print("echo response:", echo_response)


if __name__ == "__main__":
    main()
