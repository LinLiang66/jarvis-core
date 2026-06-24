package com.enio.openplatform;

import cn.hutool.core.lang.TypeReference;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.enio.openplatform.crypto.A2Md5Sign;
import com.enio.openplatform.crypto.RsaUtil;
import com.enio.openplatform.crypto.TripleDesCipher;
import com.enio.openplatform.model.SecretKeyResponse;
import com.enio.openplatform.model.TokenResponse;

import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.security.PrivateKey;
import java.time.Duration;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * 开放平台 Java 客户端：握手 + 3DES 业务调用（JSON 使用 Hutool）。
 * <ul>
 *   <li>{@link #callGateway(Map, TypeReference)} / {@link #callEncrypted(String, String, TypeReference)} 自动将 Data 反序列化为目标类型</li>
 *   <li>token 失效（40001）时自动重新握手并重试一次</li>
 * </ul>
 */
public class OpenPlatformClient {

    private static final TypeReference<TokenResponse> TOKEN_RESPONSE = new TypeReference<>() {};
    private static final TypeReference<SecretKeyResponse> SECRET_KEY_RESPONSE = new TypeReference<>() {};

    public static final String ACTION_GET_PUBLIC_KEY = "open.session.publickey";
    public static final String ACTION_CREATE_SECRET_KEY = "microSession.create.secretkey";
    public static final String ACTION_ECHO = "open.demo.echo";

    /** token 失效或无效，需重新握手 */
    public static final int CODE_TOKEN_INVALID = 40001;
    /** 可用配额不足 */
    public static final int CODE_QUOTA_EXCEEDED = 40002;
    public static final int CODE_SUCCESS = 200;

    private static final String APP_VER = "1.0.0";
    private static final String VERSION = "V1.0";

    private final String gatewayUrl;
    private final String appId;
    private final String signSecret;
    private final PrivateKey privateKey;
    private final HttpClient httpClient;
    private final Object sessionLock = new Object();

    private String token;
    private TripleDesCipher tdesCipher;

    public OpenPlatformClient(String gatewayUrl, String appId, String signSecret, String appSecretBase64)
            throws Exception {
        this.gatewayUrl = gatewayUrl;
        this.appId = appId;
        this.signSecret = signSecret;
        this.privateKey = RsaUtil.loadPrivateKeyFromDerBase64(appSecretBase64);
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    /** 完整握手：获取 token + 3DES 密钥交换（可重复调用以刷新会话）。 */
    public void handshake() throws Exception {
        synchronized (sessionLock) {
            clearSession();
            getPublicKeyAndTokenInternal();
            init3DesKeyInternal();
        }
    }

    /** 若尚未握手则自动执行 {@link #handshake()}。 */
    public void ensureSession() throws Exception {
        synchronized (sessionLock) {
            if (token != null && !token.isEmpty() && tdesCipher != null) {
                return;
            }
            getPublicKeyAndTokenInternal();
            init3DesKeyInternal();
        }
    }

    /** 步骤 1：获取 token。 */
    public String getPublicKeyAndToken() throws Exception {
        synchronized (sessionLock) {
            return getPublicKeyAndTokenInternal();
        }
    }

    /** 步骤 2：3DES 密钥交换。 */
    public void init3DesKey() throws Exception {
        synchronized (sessionLock) {
            init3DesKeyInternal();
        }
    }

    /** 明文网关调用并将 Data 转为指定类型（握手阶段使用）。 */
    public <T> T callGateway(Map<String, String> params, TypeReference<T> typeReference) throws Exception {
        return callGateway(params, typeReference, true);
    }

    /** 加密业务调用，返回解密后的 JSON 字符串。 */
    public String callEncrypted(String action, String jsonPlain) throws Exception {
        return callEncrypted(action, jsonPlain, true);
    }

    /** 加密业务调用，自动将解密 JSON 转为目标类型；token 失效时自动重新握手并重试。 */
    public <T> T callEncrypted(String action, String jsonPlain, TypeReference<T> typeReference) throws Exception {
        ensureSession();
        try {
            return doCallEncrypted(action, jsonPlain, typeReference);
        } catch (OpenPlatformException e) {
            if (e.isTokenInvalid()) {
                rehandshake();
                return doCallEncrypted(action, jsonPlain, typeReference);
            }
            throw e;
        }
    }

    /** 加密业务调用：请求体为对象，自动 JSON 序列化并反序列化响应。 */
    public <T> T callEncrypted(String action, Object requestBody, TypeReference<T> typeReference) throws Exception {
        String jsonPlain = requestBody instanceof String s ? s : JSONUtil.toJsonStr(requestBody);
        return callEncrypted(action, jsonPlain, typeReference);
    }

    public String getToken() {
        return token;
    }

    // --- internal ---

    private String getPublicKeyAndTokenInternal() throws Exception {
        Map<String, String> params = baseParams(ACTION_GET_PUBLIC_KEY);
        params.put("timestamp", String.valueOf(System.currentTimeMillis()));
        params.put("data", "{}");

        TokenResponse body = callGateway(params, TOKEN_RESPONSE, false);
        if (body.getToken() == null || body.getToken().isEmpty()) {
            throw new IllegalStateException("empty token in response");
        }
        this.token = body.getToken();
        return token;
    }

    private void init3DesKeyInternal() throws Exception {
        if (token == null || token.isEmpty()) {
            throw new IllegalStateException("call getPublicKeyAndToken() first");
        }
        try {
            exchange3DesKeyOnce();
        } catch (OpenPlatformException e) {
            if (e.isTokenInvalid()) {
                getPublicKeyAndTokenInternal();
                exchange3DesKeyOnce();
                return;
            }
            throw e;
        }
    }

    private void exchange3DesKeyOnce() throws Exception {
        String randomNum = TripleDesCipher.randomDigits(12);
        String encrypted = RsaUtil.encryptByPrivateKey(privateKey, randomNum);

        Map<String, String> params = baseParams(ACTION_CREATE_SECRET_KEY);
        params.put("req_time", String.valueOf(System.currentTimeMillis()));
        params.put("token", token);
        params.put("data", URLEncoder.encode(encrypted, StandardCharsets.UTF_8));

        SecretKeyResponse body = callGateway(params, SECRET_KEY_RESPONSE, false);
        applySecretKeyResponse(body, randomNum);
    }

    private void applySecretKeyResponse(SecretKeyResponse body, String randomNum) throws Exception {
        String serverPart = RsaUtil.decryptByPrivateKey(privateKey, body.getServerPart());
        this.tdesCipher = new TripleDesCipher(randomNum + serverPart);
    }

    private void rehandshake() throws Exception {
        synchronized (sessionLock) {
            clearSession();
            getPublicKeyAndTokenInternal();
            init3DesKeyInternal();
        }
    }

    private void clearSession() {
        token = null;
        tdesCipher = null;
    }

    private String callEncrypted(String action, String jsonPlain, boolean canRetryOnToken) throws Exception {
        ensureSession();
        try {
            return doCallEncryptedPlain(action, jsonPlain);
        } catch (OpenPlatformException e) {
            if (canRetryOnToken && e.isTokenInvalid()) {
                rehandshake();
                return doCallEncryptedPlain(action, jsonPlain);
            }
            throw e;
        }
    }

    private <T> T doCallEncrypted(String action, String jsonPlain, TypeReference<T> typeReference) throws Exception {
        String plain = doCallEncryptedPlain(action, jsonPlain);
        return JSONUtil.toBean(plain, typeReference, false);
    }

    private String doCallEncryptedPlain(String action, String jsonPlain) throws Exception {
        if (tdesCipher == null) {
            throw new IllegalStateException("call init3DesKey() first");
        }
        String cipherData = tdesCipher.encrypt(jsonPlain.getBytes(StandardCharsets.UTF_8));
        Map<String, String> params = baseParams(action);
        params.put("req_time", String.valueOf(System.currentTimeMillis()));
        params.put("token", token);
        params.put("data", cipherData);

        Object body = callGatewayRawBody(params, true);
        String cipherBody = dataToJsonString(body);
        byte[] plain = tdesCipher.decrypt(cipherBody);
        return new String(plain, StandardCharsets.UTF_8);
    }

    private Map<String, String> baseParams(String action) {
        Map<String, String> params = new LinkedHashMap<>();
        params.put("action", action);
        params.put("appid", appId);
        params.put("appver", APP_VER);
        params.put("version", VERSION);
        params.put("sign_method", A2Md5Sign.SIGN_METHOD);
        return params;
    }

    private <T> T callGateway(Map<String, String> params, TypeReference<T> typeReference, boolean retryOnTokenInvalid) throws Exception {
        Object data = callGatewayRawBody(params, retryOnTokenInvalid);
        return convertData(data, typeReference);
    }

    private Object callGatewayRawBody(Map<String, String> params, boolean retryOnTokenInvalid) throws Exception {
        try {
            return postGateway(params);
        } catch (OpenPlatformException e) {
            if (!retryOnTokenInvalid || !e.isTokenInvalid()) {
                throw e;
            }
            synchronized (sessionLock) {
                if (ACTION_GET_PUBLIC_KEY.equals(params.get("action"))) {
                    throw e;
                }
                getPublicKeyAndTokenInternal();
                if (params.containsKey("token")) {
                    Map<String, String> retryParams = new LinkedHashMap<>(params);
                    retryParams.put("token", token);
                    return postGateway(retryParams);
                }
            }
            throw e;
        }
    }

    private Object postGateway(Map<String, String> params) throws Exception {
        Map<String, String> signed = new LinkedHashMap<>(params);
        signed.put("sign", A2Md5Sign.buildSign(signed, signSecret));

        StringBuilder form = new StringBuilder();
        for (Map.Entry<String, String> e : signed.entrySet()) {
            if (form.length() > 0) {
                form.append('&');
            }
            form.append(urlEncode(e.getKey())).append('=').append(urlEncode(e.getValue()));
        }

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(gatewayUrl))
                .timeout(Duration.ofSeconds(30))
                .header("Content-Type", "application/x-www-form-urlencoded")
                .POST(HttpRequest.BodyPublishers.ofString(form.toString()))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

        JSONObject root = JSONUtil.parseObj(response.body());
        int code = root.getInt("code");
        if (code == CODE_TOKEN_INVALID) {
            String msg = root.getStr("message", "token invalid or expired");
            throw new OpenPlatformException(String.valueOf(CODE_TOKEN_INVALID), msg);
        }
        if (code == CODE_QUOTA_EXCEEDED) {
            String msg = root.getStr("message", "quota exceeded");
            throw new OpenPlatformException(String.valueOf(CODE_QUOTA_EXCEEDED), msg);
        }
        if (code != CODE_SUCCESS) {
            String msg = root.getStr("message", "");
            throw new OpenPlatformException(String.valueOf(code), "gateway error: code=" + code + ", message=" + msg);
        }
        Object data = root.get("data");
        if (data == null || (data instanceof CharSequence cs && cs.toString().isBlank())) {
            throw new OpenPlatformException(String.valueOf(CODE_SUCCESS), "empty response data");
        }
        return data;
    }

    private static <T> T convertData(Object data, TypeReference<T> typeReference) {
        if (data instanceof JSONObject jo) {
            return JSONUtil.toBean(jo, typeReference, false);
        }
        if (data instanceof CharSequence) {
            return JSONUtil.toBean(data.toString(), typeReference, false);
        }
        return JSONUtil.toBean(JSONUtil.toJsonStr(data), typeReference, false);
    }

    private static String dataToJsonString(Object data) {
        if (data instanceof CharSequence cs) {
            return cs.toString();
        }
        return JSONUtil.toJsonStr(data);
    }

    private static String urlEncode(String value) {
        return URLEncoder.encode(value, StandardCharsets.UTF_8);
    }
}
