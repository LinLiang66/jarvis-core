package jsonutil

import (
	"encoding/json"
	"log"
)

func ToJSONBytes(v any) []byte {
	jsonByte, err := json.Marshal(v)
	if err != nil {
		log.Printf("toJSONBytes error: %v, value: %v", err, v)
		return nil
	}
	return jsonByte
}

func ToJSONString(v any) string {
	return string(ToJSONBytes(v))
}

func ConvertMapToJSONString(m any) (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
