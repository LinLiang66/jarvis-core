package com.enio.openplatform.model;

import lombok.Data;

/** 获取公钥 / token 响应 */
@Data
public class TokenResponse {
    private String token;
    private String publicKey;
}
