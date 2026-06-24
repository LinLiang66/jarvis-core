package openplatform

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/pkg/jsonexample"
)

// ActionMeta 开放平台接口元数据（代码注册，启动时同步 MySQL 并生成文档）。
type ActionMeta struct {
	Action         string
	Title          string
	Category       string
	Description    string
	Encrypted      bool
	Billable       bool
	Sort           int
	RequestSchema  string
	ResponseSchema string
	RequestFields  string
	ResponseFields string
}

var (
	actionRegistryMu sync.RWMutex
	actionRegistry   = map[string]ActionMeta{}
)

func RegisterActionMeta(meta ActionMeta) {
	if meta.Action == "" {
		return
	}
	actionRegistryMu.Lock()
	defer actionRegistryMu.Unlock()
	actionRegistry[meta.Action] = meta
}

// RegisterActionMetaWithTypes 注册元数据并从实体类型生成 JSON 示例与字段描述。
func RegisterActionMetaWithTypes(meta ActionMeta, reqSample, respSample any) {
	if reqSample != nil {
		meta.RequestSchema = jsonexample.Pretty(reqSample)
		meta.RequestFields = jsonexample.DescribeJSON(reqSample)
	}
	if respSample != nil {
		meta.ResponseSchema = jsonexample.Pretty(respSample)
		meta.ResponseFields = jsonexample.DescribeJSON(respSample)
	}
	RegisterActionMeta(meta)
}

// ApplyActionTypeSchemas 为已注册 action 补充/覆盖实体生成的 schema（handler 注册时调用）。
func ApplyActionTypeSchemas(action string, reqSample, respSample any) {
	actionRegistryMu.Lock()
	defer actionRegistryMu.Unlock()
	meta, ok := actionRegistry[action]
	if !ok {
		meta = ActionMeta{Action: action, Title: action}
	}
	if reqSample != nil {
		meta.RequestSchema = jsonexample.Pretty(reqSample)
		meta.RequestFields = jsonexample.DescribeJSON(reqSample)
	}
	if respSample != nil {
		meta.ResponseSchema = jsonexample.BusinessResponse(action, respSample)
		meta.ResponseFields = describeBusinessResponseFields(action, respSample)
	}
	actionRegistry[action] = meta
}

// EnsureActionMeta 若 action 尚未注册元数据，写入占位信息（便于新接口自动同步 MySQL）。
func EnsureActionMeta(action, category string, encrypted, billable bool) {
	actionRegistryMu.RLock()
	_, ok := actionRegistry[action]
	actionRegistryMu.RUnlock()
	if ok {
		return
	}
	RegisterActionMeta(ActionMeta{
		Action: action, Title: action, Category: category,
		Encrypted: encrypted, Billable: billable, Sort: 999,
	})
}

func ListRegisteredActions() []ActionMeta {
	actionRegistryMu.RLock()
	defer actionRegistryMu.RUnlock()
	out := make([]ActionMeta, 0, len(actionRegistry))
	for _, m := range actionRegistry {
		out = append(out, m)
	}
	return out
}

func GetRegisteredActionMeta(action string) (ActionMeta, bool) {
	actionRegistryMu.RLock()
	defer actionRegistryMu.RUnlock()
	m, ok := actionRegistry[action]
	return m, ok
}

func applyRegistryMetaToRow(row *model.OpenAPIAction, m ActionMeta) {
	row.Title = m.Title
	row.Category = m.Category
	row.Description = m.Description
	row.Encrypted = m.Encrypted
	row.Billable = m.Billable
	if isHandshakeAction(m.Action) {
		row.Billable = false
	}
	row.RequestSchema = m.RequestSchema
	row.ResponseSchema = m.ResponseSchema
	row.RequestFields = m.RequestFields
	row.ResponseFields = m.ResponseFields
	row.Sort = m.Sort
	row.DocMarkdown = GenerateDocMarkdown(m)
}

func RegistryToModels() []model.OpenAPIAction {
	metas := ListRegisteredActions()
	out := make([]model.OpenAPIAction, 0, len(metas))
	for _, m := range metas {
		billable := m.Billable
		if isHandshakeAction(m.Action) {
			billable = false
		}
		doc := GenerateDocMarkdown(m)
		out = append(out, model.OpenAPIAction{
			Action:         m.Action,
			Title:          m.Title,
			Category:       m.Category,
			Description:    m.Description,
			Encrypted:      m.Encrypted,
			Billable:       billable,
			Status:         "0",
			RequestSchema:  m.RequestSchema,
			ResponseSchema: m.ResponseSchema,
			RequestFields:  m.RequestFields,
			ResponseFields: m.ResponseFields,
			DocMarkdown:    doc,
			Sort:           m.Sort,
			Source:         "code",
		})
	}
	return out
}

func GenerateDocMarkdown(m ActionMeta) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", m.Title))
	b.WriteString(fmt.Sprintf("- **Action**: `%s`\n", m.Action))
	b.WriteString(fmt.Sprintf("- **分类**: %s\n", m.Category))
	if m.Encrypted {
		b.WriteString("- **传输**: 3DES 加密业务请求（需先完成握手）\n")
	} else {
		b.WriteString("- **传输**: 明文 form-urlencoded\n")
	}
	if m.Billable {
		b.WriteString("- **计费**: 是（消耗可用配额）\n")
	} else {
		b.WriteString("- **计费**: 否\n")
	}
	if isHandshakeAction(m.Action) {
		b.WriteString("- **统计**: 否（握手接口不计入调用次数）\n")
	}
	b.WriteString("\n## 说明\n\n")
	if m.Description != "" {
		b.WriteString(m.Description + "\n\n")
	}
	b.WriteString("## 网关调用\n\n")
	b.WriteString("统一入口 `POST /api/v1/open/gateway`，Content-Type: `application/x-www-form-urlencoded`。\n\n")
	if m.Encrypted {
		b.WriteString("业务参数放在 `data` 字段（3DES 加密后的 JSON 字符串），并携带 `token`。\n\n")
	}
	b.WriteString("## 请求体\n\n")
	if m.RequestSchema != "" {
		b.WriteString("```json\n")
		b.WriteString(strings.TrimSpace(m.RequestSchema))
		b.WriteString("\n```\n\n")
	} else {
		b.WriteString("_无_\n\n")
	}
	b.WriteString("## 响应体\n\n")
	b.WriteString("网关外层响应：\n\n```json\n")
	b.WriteString(jsonexample.Pretty(docGatewayWrapResp{
		Code: 200, Message: "success", Data: "<3DES 密文或 JSON 对象>", Success: true,
	}))
	b.WriteString("\n```\n\n")
	if m.Encrypted {
		b.WriteString("3DES 解密后的业务 JSON：\n\n```json\n")
		if m.ResponseSchema != "" {
			b.WriteString(strings.TrimSpace(m.ResponseSchema))
		} else {
			b.WriteString(`{
  "action": "` + m.Action + `",
  "data": { }
}`)
		}
		b.WriteString("\n```\n")
	} else if m.ResponseSchema != "" {
		b.WriteString("```json\n")
		b.WriteString(strings.TrimSpace(m.ResponseSchema))
		b.WriteString("\n```\n")
	}
	return b.String()
}

func RegenerateDoc(row *model.OpenAPIAction) {
	if row.RequestFields == "" && row.RequestSchema != "" {
		row.RequestFields = jsonexample.InferFieldsJSON(row.RequestSchema)
	}
	if row.ResponseFields == "" && row.ResponseSchema != "" {
		row.ResponseFields = jsonexample.InferFieldsJSON(row.ResponseSchema)
	}
	meta := ActionMeta{
		Action:         row.Action,
		Title:          row.Title,
		Category:       row.Category,
		Description:    row.Description,
		Encrypted:      row.Encrypted,
		Billable:       row.Billable,
		RequestSchema:  row.RequestSchema,
		ResponseSchema: row.ResponseSchema,
		RequestFields:  row.RequestFields,
		ResponseFields: row.ResponseFields,
	}
	row.DocMarkdown = GenerateDocMarkdown(meta)
}

func describeBusinessResponseFields(action string, respSample any) string {
	fields := []jsonexample.FieldDesc{
		{Name: "action", Type: "string", Required: true, Example: action},
		{
			Name: "data", Type: "object", Required: true,
			Children: jsonexample.Describe(respSample),
		},
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return "[]"
	}
	return string(b)
}
