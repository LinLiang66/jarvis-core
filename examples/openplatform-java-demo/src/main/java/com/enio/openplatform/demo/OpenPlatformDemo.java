package com.enio.openplatform.demo;

import cn.hutool.core.lang.TypeReference;
import com.enio.openplatform.OpenPlatformClient;
import com.enio.openplatform.model.EchoResponse;

import java.util.Map;

/**
 * Demo: handshake + encrypted Echo (open.demo.echo).
 *
 * <pre>
 * mvn -q exec:java \
 *   -Dexec.args="http://127.0.0.1:8000/api/v1/open/gateway app_xxx signSecret appSecretBase64"
 * </pre>
 *
 * Env: OPEN_GATEWAY_URL, OPEN_APP_ID, OPEN_SIGN_SECRET, OPEN_APP_SECRET
 */
public class OpenPlatformDemo {

    public static void main(String[] args) throws Exception {
        String gateway = args.length > 0 ? args[0] : envOr("OPEN_GATEWAY_URL", "http://127.0.0.1:8000/api/v1/open/gateway");
        String appId = args.length > 1 ? args[1] : envOr("OPEN_APP_ID", "");
        String signSecret = args.length > 2 ? args[2] : envOr("OPEN_SIGN_SECRET", "");
        String appSecret = args.length > 3 ? args[3] : envOr("OPEN_APP_SECRET", "");

        if (appId.isEmpty() || signSecret.isEmpty() || appSecret.isEmpty()) {
            System.err.println("usage: OpenPlatformDemo [gateway] [appId] [signSecret] [appSecret]");
            System.err.println("or set OPEN_APP_ID / OPEN_SIGN_SECRET / OPEN_APP_SECRET");
            System.exit(1);
        }

        OpenPlatformClient client = new OpenPlatformClient(gateway, appId, signSecret, appSecret);

        client.handshake();
        System.out.println("token: " + client.getToken());
        System.out.println("3des session ready");

        EchoResponse resp = client.callEncrypted(
                OpenPlatformClient.ACTION_ECHO,
                Map.of("hello", "jarvis demo"),
                new TypeReference<EchoResponse>() {}
        );

        System.out.printf("echo response: action=%s message=%s echo=%s%n",
                resp.getAction(), resp.getMessage(), resp.getEcho());
    }

    private static String envOr(String key, String def) {
        String v = System.getenv(key);
        return v != null && !v.isBlank() ? v : def;
    }
}
