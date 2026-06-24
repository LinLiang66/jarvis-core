package jsonexample

import (
	"encoding/json"
	"reflect"
	"strings"
)

// FieldDesc 接口文档字段描述（用于参数表格渲染）。
type FieldDesc struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Required bool        `json:"required"`
	Example  any         `json:"example,omitempty"`
	Children []FieldDesc `json:"children,omitempty"`
}

// Sample 根据类型生成示例值（map/slice/标量）。
func Sample(v any) any {
	if v == nil {
		return nil
	}
	return sampleValue(reflect.TypeOf(v), reflect.ValueOf(v))
}

// Pretty 生成格式化 JSON 示例字符串。
func Pretty(v any) string {
	b, err := json.MarshalIndent(Sample(v), "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Describe 从实体类型生成字段描述列表。
func Describe(v any) []FieldDesc {
	if v == nil {
		return nil
	}
	return describeType(reflect.TypeOf(v), 0)
}

// DescribeJSON 将字段描述序列化为 JSON 字符串。
func DescribeJSON(v any) string {
	b, err := json.Marshal(Describe(v))
	if err != nil {
		return "[]"
	}
	return string(b)
}

// BusinessResponse 生成开放平台业务响应示例 { action, data }。
func BusinessResponse(action string, dataSample any) string {
	m := map[string]any{
		"action": action,
		"data":   Sample(dataSample),
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

func sampleValue(typ reflect.Type, val reflect.Value) any {
	return sampleFieldValue(typ, val, typ.Name())
}

func sampleFieldValue(typ reflect.Type, val reflect.Value, fieldName string) any {
	for typ.Kind() == reflect.Pointer {
		if typ.NumMethod() == 0 && typ.Elem().Kind() != reflect.Pointer {
			typ = typ.Elem()
			if val.IsValid() && !val.IsNil() {
				val = val.Elem()
			} else {
				val = reflect.Value{}
			}
			continue
		}
		break
	}

	switch typ.Kind() {
	case reflect.String:
		return placeholderString(fieldName)
	case reflect.Bool:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return placeholderInt(fieldName)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return placeholderInt(fieldName)
	case reflect.Float32, reflect.Float64:
		return 0.0
	case reflect.Slice, reflect.Array:
		elemType := typ.Elem()
		if elemType.Kind() == reflect.Uint8 {
			return "<base64>"
		}
		item := sampleFieldValue(elemType, reflect.Zero(elemType), elemType.Name())
		if item == nil {
			return []any{}
		}
		return []any{item}
	case reflect.Map:
		if typ.Key().Kind() == reflect.String {
			return map[string]any{"key": "value"}
		}
		return map[string]any{}
	case reflect.Interface:
		return nil
	case reflect.Struct:
		return sampleStruct(typ, val)
	default:
		return nil
	}
}

func sampleStruct(typ reflect.Type, val reflect.Value) map[string]any {
	val = unwrapPointerValue(val)
	out := make(map[string]any)
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if !sf.IsExported() {
			continue
		}
		name, omitEmpty := parseJSONTag(sf)
		if name == "" || name == "-" {
			continue
		}
		ft := sf.Type
		fv := reflect.Zero(ft)
		if val.IsValid() && val.Kind() == reflect.Struct {
			fv = val.Field(i)
		}
		if omitEmpty && isZeroValue(fv) {
			continue
		}
		out[name] = sampleFieldValue(ft, fv, name)
	}
	return out
}

func unwrapPointerValue(val reflect.Value) reflect.Value {
	for val.IsValid() && val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return reflect.Value{}
		}
		val = val.Elem()
	}
	return val
}

func describeType(typ reflect.Type, depth int) []FieldDesc {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	const maxDepth = 4
	if depth > maxDepth {
		return nil
	}
	out := make([]FieldDesc, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if !sf.IsExported() {
			continue
		}
		name, omitEmpty := parseJSONTag(sf)
		if name == "" || name == "-" {
			continue
		}
		ft := sf.Type
		desc := FieldDesc{
			Name:     name,
			Type:     fieldTypeName(ft),
			Required: !omitEmpty,
			Example:  sampleValue(ft, reflect.Zero(ft)),
		}
		if ft.Kind() == reflect.Struct || (ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct) {
			desc.Children = describeType(ft, depth+1)
		} else if ft.Kind() == reflect.Slice || ft.Kind() == reflect.Array {
			elem := ft.Elem()
			if elem.Kind() == reflect.Struct || (elem.Kind() == reflect.Pointer && elem.Elem().Kind() == reflect.Struct) {
				desc.Children = describeType(elem, depth+1)
			}
		}
		out = append(out, desc)
	}
	return out
}

func parseJSONTag(sf reflect.StructField) (name string, omitEmpty bool) {
	tag := sf.Tag.Get("json")
	if tag == "" {
		return sf.Name, false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, p := range parts[1:] {
		if p == "omitempty" {
			omitEmpty = true
		}
	}
	return name, omitEmpty
}

func fieldTypeName(typ reflect.Type) string {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		return "array"
	case reflect.Map:
		return "object"
	case reflect.Interface:
		return "any"
	case reflect.Struct:
		return "object"
	default:
		return "any"
	}
}

func isZeroValue(v reflect.Value) bool {
	return v.IsZero()
}

func placeholderString(field string) string {
	lower := strings.ToLower(field)
	switch {
	case lower == "msg", lower == "message":
		return "ok"
	case lower == "timestamp", lower == "req_time", lower == "reqtime":
		return "1710000000000"
	case lower == "token":
		return "<token>"
	case strings.Contains(lower, "data"):
		if strings.Contains(lower, "base") || lower == "data" {
			return "<base64>"
		}
		return "{}"
	case lower == "filename":
		return "demo.png"
	case lower == "contenttype":
		return "image/png"
	case lower == "publickey":
		return "<RSA公钥DER Base64>"
	case lower == "serverpart":
		return "<RSA加密的serverRandom>"
	case strings.Contains(lower, "orderno"):
		return "运单号"
	case strings.Contains(lower, "orgcode"):
		return "241885"
	case strings.Contains(lower, "orgname"):
		return "网点名称"
	case lower == "userid":
		return "241885022"
	case strings.Contains(lower, "username"):
		return "用户名"
	case strings.Contains(lower, "dicttype"):
		return "order_type_code"
	case strings.Contains(lower, "ordercode"):
		return "问题类型编码"
	case strings.Contains(lower, "dealsite"):
		return "处理网点"
	case strings.Contains(lower, "orderstate"):
		return "状态"
	case strings.Contains(lower, "orderpic"), strings.Contains(lower, "fileid"):
		return "图片fileId"
	case strings.Contains(lower, "hello"):
		return "world"
	case strings.Contains(lower, "time"), strings.Contains(lower, "date"):
		return "2024-01-01 00:00:00"
	case strings.HasSuffix(lower, "url"):
		return "https://example.com/file"
	case strings.HasSuffix(lower, "name"):
		return "名称"
	case strings.HasSuffix(lower, "code"):
		return "编码"
	case strings.HasSuffix(lower, "id"):
		return "..."
	default:
		return field
	}
}

// InferFieldsJSON 从 JSON 示例字符串推断字段描述（手动录入接口用）。
func InferFieldsJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "[]"
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return "[]"
	}
	fields := inferJSONValue(v, "")
	b, err := json.Marshal(fields)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func inferJSONValue(v any, name string) []FieldDesc {
	switch val := v.(type) {
	case map[string]any:
		out := make([]FieldDesc, 0, len(val))
		for k, child := range val {
			desc := FieldDesc{Name: k, Required: true, Example: child}
			switch c := child.(type) {
			case map[string]any:
				desc.Type = "object"
				desc.Children = inferJSONValue(c, k)
			case []any:
				desc.Type = "array"
				if len(c) > 0 {
					desc.Children = inferJSONValue(c[0], k)
				}
			case string:
				desc.Type = "string"
			case float64:
				desc.Type = "number"
			case bool:
				desc.Type = "boolean"
			case nil:
				desc.Type = "any"
			default:
				desc.Type = "any"
			}
			out = append(out, desc)
		}
		return out
	default:
		if name != "" {
			return []FieldDesc{{Name: name, Type: jsonTypeOf(v), Required: true, Example: v}}
		}
		return nil
	}
}

func jsonTypeOf(v any) string {
	switch v.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "any"
	}
}

func placeholderInt(field string) int {
	lower := strings.ToLower(field)
	switch {
	case strings.Contains(lower, "pagesize"):
		return 20
	case strings.Contains(lower, "pagenum"):
		return 1
	case strings.Contains(lower, "tabtype"):
		return 1
	case strings.Contains(lower, "includeext"), strings.Contains(lower, "addallathead"):
		return 1
	default:
		return 0
	}
}

