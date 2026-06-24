package parseid

import (
	"bytes"
	"encoding/json"
	"strings"
)

// FlexInt64 兼容 JSON 数字或字符串（前端 string id）
type FlexInt64 int64

func (f *FlexInt64) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*f = 0
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			*f = 0
			return nil
		}
		n, err := Path(s)
		if err != nil {
			return err
		}
		*f = FlexInt64(n)
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*f = FlexInt64(n)
	return nil
}

func (f FlexInt64) Int64() int64 {
	return int64(f)
}
