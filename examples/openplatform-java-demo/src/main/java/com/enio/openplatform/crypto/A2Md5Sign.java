package com.enio.openplatform.crypto;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.Map;
import java.util.TreeMap;

/**
 * A2 MD5 签名：参数 key 升序，拼接 key+value，末尾追加 signSecret，MD5 小写 hex。
 */
public final class A2Md5Sign {

    public static final String SIGN_METHOD = "a2_md5";

    private A2Md5Sign() {
    }

    public static String buildSign(Map<String, String> params, String signSecret) {
        TreeMap<String, String> sorted = new TreeMap<>();
        for (Map.Entry<String, String> e : params.entrySet()) {
            if ("sign".equals(e.getKey())) {
                continue;
            }
            String v = e.getValue();
            if (v == null || v.isEmpty()) {
                continue;
            }
            sorted.put(e.getKey(), v);
        }
        StringBuilder sb = new StringBuilder();
        for (Map.Entry<String, String> e : sorted.entrySet()) {
            sb.append(e.getKey()).append(e.getValue());
        }
        sb.append(signSecret);
        return md5Hex(sb.toString());
    }

    private static String md5Hex(String input) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] digest = md.digest(input.getBytes(StandardCharsets.UTF_8));
            StringBuilder hex = new StringBuilder(digest.length * 2);
            for (byte b : digest) {
                hex.append(String.format("%02x", b));
            }
            return hex.toString();
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException("MD5 not available", e);
        }
    }
}
