package serialize

import (
	"strconv"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

// apiJSON 对外 API 序列化：int64 输出为 JSON 字符串，避免前端精度丢失
var apiJSON = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func init() {
	jsoniter.RegisterTypeEncoderFunc("int64", encodeInt64, nil)
	jsoniter.RegisterTypeEncoderFunc("*int64", encodeInt64Ptr, nil)
}

func encodeInt64(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	num := *((*int64)(ptr))
	stream.WriteString(strconv.FormatInt(num, 10))
}

func encodeInt64Ptr(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	p := *((**int64)(ptr))
	if p == nil {
		stream.WriteNil()
		return
	}
	stream.WriteString(strconv.FormatInt(*p, 10))
}

// MarshalJSON 序列化响应体（int64 字段为字符串）
func MarshalJSON(v any) ([]byte, error) {
	return apiJSON.Marshal(v)
}
