package serialize

import (
	"strconv"
	"time"

	"jarvis-core/backend/internal/model"
)

func IDStr(id int64) string {
	return strconv.FormatInt(id, 10)
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func userRoleMeta(u model.SysUser) (roleIDs, roleNames, roleCodes []string) {
	for _, r := range u.Roles {
		roleIDs = append(roleIDs, IDStr(r.ID))
		roleNames = append(roleNames, r.Name)
		roleCodes = append(roleCodes, r.Code)
	}
	if roleIDs == nil {
		roleIDs = []string{}
	}
	if roleNames == nil {
		roleNames = []string{}
	}
	if roleCodes == nil {
		roleCodes = []string{}
	}
	return roleIDs, roleNames, roleCodes
}

func userDeptID(u model.SysUser) any {
	if u.DeptID == nil {
		return nil
	}
	return IDStr(*u.DeptID)
}

func UserDTO(u model.SysUser, roleIDs, roleNames, roles []string) map[string]any {
	return map[string]any{
		"id":           IDStr(u.ID),
		"username":     u.Username,
		"nickname":     u.Nickname,
		"phone":        u.Phone,
		"email":        u.Email,
		"avatar":       u.Avatar,
		"remark":       u.Remark,
		"status":       u.Status,
		"sort":         u.Sort,
		"createTime":   FormatTime(u.CreatedAt),
		"deptId":       userDeptID(u),
		"isSuperAdmin": u.IsSuperAdmin,
		"roleIds":      roleIDs,
		"roleNames":    roleNames,
		"roles":        roles,
	}
}

func RoleDTO(r model.SysRole) map[string]any {
	return map[string]any{
		"id":         IDStr(r.ID),
		"code":       r.Code,
		"name":       r.Name,
		"status":     r.Status,
		"sort":       r.Sort,
		"remark":     r.Remark,
		"isSystem":   r.IsSystem || r.Code == model.RoleAdminCode,
		"createTime": FormatTime(r.CreatedAt),
	}
}

func MenuDTO(m model.SysMenu) map[string]any {
	parent := IDStr(m.ParentID)
	if m.ParentID == 0 {
		parent = "0"
	}
	return map[string]any{
		"id":          IDStr(m.ID),
		"parentId":    parent,
		"type":        m.Type,
		"title":       m.Title,
		"path":        m.Path,
		"component":   m.Component,
		"redirect":    m.Redirect,
		"icon":        m.Icon,
		"permission":  m.Permission,
		"sort":        m.Sort,
		"status":      m.Status,
		"hidden":      m.Hidden,
		"keepAlive":   m.KeepAlive,
		"affix":       m.Affix,
		"alwaysShow":  m.AlwaysShow,
		"breadcrumb":  m.Breadcrumb,
		"showInTabs":  m.ShowInTabs,
		"activeMenu":  m.ActiveMenu,
		"isSystem":    m.IsSystem,
		"roles":       []string{},
		"children":    []map[string]any{},
	}
}

func DictTypeDTO(d model.SysDictType) map[string]any {
	return map[string]any{
		"id":         IDStr(d.ID),
		"name":       d.Name,
		"code":       d.Code,
		"status":     d.Status,
		"sort":       d.Sort,
		"remark":     d.Remark,
		"isSystem":   d.IsSystem,
		"createTime": FormatTime(d.CreatedAt),
		"updateTime": FormatTime(d.UpdatedAt),
	}
}

func DictDataDTO(d model.SysDictData) map[string]any {
	return map[string]any{
		"id":         IDStr(d.ID),
		"typeId":     IDStr(d.TypeID),
		"label":      d.Label,
		"value":      d.Value,
		"status":     d.Status,
		"sort":       d.Sort,
		"remark":     d.Remark,
		"createTime": FormatTime(d.CreatedAt),
	}
}

func DictOption(label, value string) map[string]string {
	return map[string]string{"label": label, "value": value}
}

func StorageDTO(s model.SysStorage, maskSecret bool) map[string]any {
	dto := map[string]any{
		"id":          IDStr(s.ID),
		"name":        s.Name,
		"code":        s.Code,
		"type":        s.Type,
		"accessKey":   s.AccessKey,
		"endpoint":    s.Endpoint,
		"bucketName":  s.BucketName,
		"baseUrl":     s.BaseURL,
		"domain":      s.Domain,
		"description": s.Description,
		"isDefault":   s.IsDefault,
		"sort":        s.Sort,
		"status":      s.Status,
		"createTime":  FormatTime(s.CreatedAt),
		"updateTime":  FormatTime(s.UpdatedAt),
	}
	if maskSecret {
		if s.SecretKey != "" {
			dto["secretKey"] = "******"
		} else {
			dto["secretKey"] = ""
		}
	}
	return dto
}

func fileDTOType(f model.SysFile) int {
	if f.IsDir() {
		return model.FileTypeDir
	}
	return model.FileTypeFile
}

func FileDTO(f model.SysFile) map[string]any {
	return map[string]any{
		"id":           IDStr(f.ID),
		"storageId":    IDStr(f.StorageID),
		"name":         f.Name,
		"originalName": f.OriginalName,
		"path":         f.Path,
		"parentPath":   f.ParentPath,
		"url":          f.URL,
		"size":         f.Size,
		"extension":    f.Extension,
		"contentType":  f.ContentType,
		"type":         fileDTOType(f),
		"createTime":   FormatTime(f.CreatedAt),
		"updateTime":   FormatTime(f.UpdatedAt),
	}
}

func FileDTOEnriched(f model.SysFile, url, storageName string) map[string]any {
	dto := FileDTO(f)
	if url != "" {
		dto["url"] = url
	}
	if storageName != "" {
		dto["storageName"] = storageName
	}
	return dto
}

// LoginUserDTO POST /auth/login 响应中的 user，字段对齐前端 LoginUser
func LoginUserDTO(u model.SysUser, isSuperuser bool) map[string]any {
	dto := map[string]any{
		"id":          int(u.ID),
		"username":    u.Username,
		"name":        u.Nickname,
		"isSuperuser": isSuperuser,
		"status":      u.Status,
		"createTime":  FormatTime(u.CreatedAt),
	}
	if u.Phone != "" {
		dto["mobile"] = u.Phone
	}
	if u.Email != "" {
		dto["email"] = u.Email
	}
	if u.Avatar != "" {
		dto["avatar"] = u.Avatar
	}
	if u.DeptID != nil {
		dto["deptId"] = int(*u.DeptID)
	}
	return dto
}

// UserInfoDTO GET /auth/userinfo 响应，字段对齐前端 UserInfo
func UserInfoDTO(u model.SysUser, permissions []string) map[string]any {
	perms := permissions
	if perms == nil {
		perms = []string{}
	}
	roleIDs, roleNames, _ := userRoleMeta(u)
	enabledRoles := make([]string, 0)
	for _, r := range u.Roles {
		if r.Status == "0" {
			enabledRoles = append(enabledRoles, r.Code)
		}
	}
	dto := map[string]any{
		"id":          IDStr(u.ID),
		"username":    u.Username,
		"nickname":    u.Nickname,
		"phone":       u.Phone,
		"email":       u.Email,
		"avatar":      u.Avatar,
		"remark":      u.Remark,
		"status":      u.Status,
		"sort":        u.Sort,
		"createTime":  FormatTime(u.CreatedAt),
		"deptId":      userDeptID(u),
		"roleIds":     roleIDs,
		"roleNames":   roleNames,
		"roles":       enabledRoles,
		"permissions": perms,
	}
	if dto["avatar"] == "" {
		dto["avatar"] = nil
	}
	return dto
}
