package com.enio.openplatform.crypto;

import javax.crypto.Cipher;
import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;
import java.util.Base64;

/**
 * 3DES/ECB/PKCS5Padding，与会话密钥交换后的 finalKey 一致；实例可复用，避免每次初始化。
 */
public final class TripleDesCipher {

    private static final String ALGORITHM = "DESede/ECB/PKCS5Padding";

    private final SecretKey secretKey;

    public TripleDesCipher(String sessionKey) throws Exception {
        this.secretKey = new SecretKeySpec(normalizeKey24(sessionKey.getBytes(StandardCharsets.UTF_8)), "DESede");
    }

    public String encrypt(byte[] plain) throws Exception {
        Cipher cipher = Cipher.getInstance(ALGORITHM);
        cipher.init(Cipher.ENCRYPT_MODE, secretKey);
        byte[] encrypted = cipher.doFinal(plain);
        return Base64.getEncoder().encodeToString(encrypted);
    }

    public byte[] decrypt(String cipherBase64) throws Exception {
        Cipher cipher = Cipher.getInstance(ALGORITHM);
        cipher.init(Cipher.DECRYPT_MODE, secretKey);
        byte[] raw = Base64.getDecoder().decode(cipherBase64);
        return cipher.doFinal(raw);
    }

    private static byte[] normalizeKey24(byte[] key) {
        if (key.length == 16) {
            byte[] out = new byte[24];
            System.arraycopy(key, 0, out, 0, 16);
            System.arraycopy(key, 0, out, 16, 8);
            return out;
        }
        if (key.length == 24) {
            return key;
        }
        if (key.length > 24) {
            byte[] out = new byte[24];
            System.arraycopy(key, 0, out, 0, 24);
            return out;
        }
        byte[] out = new byte[24];
        System.arraycopy(key, 0, out, 0, key.length);
        return out;
    }

    public static String randomDigits(int length) {
        SecureRandom random = new SecureRandom();
        StringBuilder sb = new StringBuilder(length);
        for (int i = 0; i < length; i++) {
            sb.append(random.nextInt(10));
        }
        return sb.toString();
    }
}
