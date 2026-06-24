package com.enio.openplatform.model;

import lombok.Data;

import java.util.Map;

/** Echo demo response (open.demo.echo). */
@Data
public class EchoResponse {
    private String action;
    private Map<String, Object> echo;
    private String message;
}
