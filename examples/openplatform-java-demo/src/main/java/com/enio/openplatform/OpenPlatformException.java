package com.enio.openplatform;

/**
 * 开放平台网关业务异常。
 */
public class OpenPlatformException extends RuntimeException {

    private final String code;

    public OpenPlatformException(String code, String message) {
        super(message);
        this.code = code;
    }

    public String getCode() {
        return code;
    }

    public boolean isTokenInvalid() {
        return String.valueOf(OpenPlatformClient.CODE_TOKEN_INVALID).equals(code);
    }

    public boolean isQuotaExceeded() {
        return String.valueOf(OpenPlatformClient.CODE_QUOTA_EXCEEDED).equals(code);
    }
}
