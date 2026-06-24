package com.enio.openplatform.crypto;

import javax.crypto.Cipher;
import java.nio.charset.StandardCharsets;
import java.security.KeyFactory;
import java.security.PrivateKey;
import java.security.spec.PKCS8EncodedKeySpec;
import java.util.Base64;

/**
 * RSA/ECB/PKCS1Padding，与 Go 服务端及常见 Java RSAUtil 对齐。
 */
public final class RsaUtil {

    private static final String RSA = "RSA/ECB/PKCS1Padding";

    private RsaUtil() {
    }

    /** 从管理端下发的 AppSecret（PKCS#8 DER Base64）加载私钥。 */
    public static PrivateKey loadPrivateKeyFromDerBase64(String appSecretBase64) throws Exception {
        byte[] der = Base64.getDecoder().decode(appSecretBase64.trim());
        PKCS8EncodedKeySpec spec = new PKCS8EncodedKeySpec(der);
        return KeyFactory.getInstance("RSA").generatePrivate(spec);
    }

    /** 私钥加密（密钥交换时加密 clientRandom）。 */
    public static String encryptByPrivateKey(PrivateKey privateKey, String plain) throws Exception {
        Cipher cipher = Cipher.getInstance(RSA);
        cipher.init(Cipher.ENCRYPT_MODE, privateKey);
        byte[] encrypted = cipher.doFinal(plain.getBytes(StandardCharsets.UTF_8));
        return Base64.getEncoder().encodeToString(encrypted);
    }

    /** 私钥解密（解密服务端返回的 serverPart）。 */
    public static String decryptByPrivateKey(PrivateKey privateKey, String cipherBase64) throws Exception {
        Cipher cipher = Cipher.getInstance(RSA);
        cipher.init(Cipher.DECRYPT_MODE, privateKey);
        byte[] raw = Base64.getDecoder().decode(cipherBase64);
        byte[] plain = cipher.doFinal(raw);
        return new String(plain, StandardCharsets.UTF_8);
    }
}
