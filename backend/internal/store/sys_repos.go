package store

import (
	"context"
	"sort"
	"strconv"

	"gorm.io/gorm"

	"jarvis-core/backend/internal/infra/base"
	"jarvis-core/backend/internal/model"
)

type SysUserRepository struct{ base.CRUD }

func NewSysUserRepository(db *gorm.DB) *SysUserRepository {
	return &SysUserRepository{CRUD: base.CRUD{DB: db}}
}

func (r *SysUserRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.SysUser{})
}

type SysUserFilter struct {
	Username string
	Phone    string
	Status   string
}

func (r *SysUserRepository) List(ctx context.Context, pq PageQuery, f SysUserFilter) ([]model.SysUser, int64, error) {
	return ListPage[model.SysUser](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if f.Username != "" {
			q = q.Where("username LIKE ?", "%"+f.Username+"%")
		}
		if f.Phone != "" {
			q = q.Where("phone LIKE ?", "%"+f.Phone+"%")
		}
		if f.Status != "" {
			q = q.Where("status = ?", f.Status)
		}
		return q
	})
}

func (r *SysUserRepository) GetByID(ctx context.Context, id any) (*model.SysUser, error) {
	var row model.SysUser
	if err := r.DB.WithContext(ctx).Preload("Roles").First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysUserRepository) GetByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	var row model.SysUser
	if err := r.DB.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysUserRepository) Create(ctx context.Context, row *model.SysUser) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysUserRepository) Save(ctx context.Context, row *model.SysUser) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysUserRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysUser{}, ids).Error
}

func (r *SysUserRepository) ReplaceRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	user := model.SysUser{ID: userID}
	var roles []model.SysRole
	if len(roleIDs) > 0 {
		if err := r.DB.WithContext(ctx).Find(&roles, roleIDs).Error; err != nil {
			return err
		}
	}
	return r.DB.WithContext(ctx).Model(&user).Association("Roles").Replace(roles)
}

type SysRoleRepository struct{ base.CRUD }

func NewSysRoleRepository(db *gorm.DB) *SysRoleRepository {
	return &SysRoleRepository{CRUD: base.CRUD{DB: db}}
}

func (r *SysRoleRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.SysRole{})
}

func (r *SysRoleRepository) List(ctx context.Context, pq PageQuery, code, name, status string) ([]model.SysRole, int64, error) {
	return ListPage[model.SysRole](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if code != "" {
			q = q.Where("code LIKE ?", "%"+code+"%")
		}
		if name != "" {
			q = q.Where("name LIKE ?", "%"+name+"%")
		}
		if status != "" {
			q = q.Where("status = ?", status)
		}
		return q
	})
}

func (r *SysRoleRepository) AllEnabled(ctx context.Context) ([]model.SysRole, error) {
	var rows []model.SysRole
	err := r.DB.WithContext(ctx).Where("status = ?", "0").Order("sort asc, id asc").Find(&rows).Error
	return rows, err
}

func (r *SysRoleRepository) GetByID(ctx context.Context, id any) (*model.SysRole, error) {
	var row model.SysRole
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysRoleRepository) Create(ctx context.Context, row *model.SysRole) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysRoleRepository) Save(ctx context.Context, row *model.SysRole) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysRoleRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysRole{}, ids).Error
}

func (r *SysRoleRepository) MenuIDs(ctx context.Context, roleID int64) ([]string, error) {
	ids, err := r.roleMenuIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = strconv.FormatInt(id, 10)
	}
	return out, nil
}

// MenuLeafIDs 返回角色已分配菜单中的叶子节点，供树组件回显勾选。
func (r *SysRoleRepository) MenuLeafIDs(ctx context.Context, roleID int64) ([]string, error) {
	ids, err := r.roleMenuIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	all, err := NewSysMenuRepository(r.DB).All(ctx)
	if err != nil {
		return nil, err
	}
	leaves := filterLeafMenuIDs(all, ids)
	out := make([]string, len(leaves))
	for i, id := range leaves {
		out[i] = strconv.FormatInt(id, 10)
	}
	return out, nil
}

func (r *SysRoleRepository) roleMenuIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.DB.WithContext(ctx).Table("sys_role_menus").Where("sys_role_id = ?", roleID).Pluck("sys_menu_id", &ids).Error
	return ids, err
}

func filterLeafMenuIDs(all []model.SysMenu, assigned []int64) []int64 {
	if len(assigned) == 0 {
		return nil
	}
	idSet := make(map[int64]struct{}, len(assigned))
	for _, id := range assigned {
		idSet[id] = struct{}{}
	}
	children := make(map[int64][]int64)
	for _, m := range all {
		children[m.ParentID] = append(children[m.ParentID], m.ID)
	}
	var leaves []int64
	for _, mid := range assigned {
		hasAssignedChild := false
		for _, childID := range children[mid] {
			if _, ok := idSet[childID]; ok {
				hasAssignedChild = true
				break
			}
		}
		if !hasAssignedChild {
			leaves = append(leaves, mid)
		}
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
	return leaves
}

func expandMenuIDs(all []model.SysMenu, menuIDs []int64) []int64 {
	if len(menuIDs) == 0 {
		return nil
	}
	byID := make(map[int64]model.SysMenu, len(all))
	children := make(map[int64][]int64)
	for _, m := range all {
		byID[m.ID] = m
		children[m.ParentID] = append(children[m.ParentID], m.ID)
	}
	result := make(map[int64]struct{})
	var collectDescendants func(int64)
	collectDescendants = func(parentID int64) {
		for _, childID := range children[parentID] {
			result[childID] = struct{}{}
			collectDescendants(childID)
		}
	}
	for _, mid := range menuIDs {
		if _, ok := byID[mid]; !ok {
			continue
		}
		result[mid] = struct{}{}
		if byID[mid].Type == 1 {
			collectDescendants(mid)
		}
	}
	out := make([]int64, 0, len(result))
	for id := range result {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func expandMenuAncestors(all []model.SysMenu, menuIDs []int64) []int64 {
	if len(menuIDs) == 0 {
		return nil
	}
	allow := expandAncestors(all, menuIDs)
	out := make([]int64, 0, len(allow))
	for id := range allow {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (r *SysRoleRepository) ReplaceMenus(ctx context.Context, roleID int64, menuIDs []int64) error {
	all, err := NewSysMenuRepository(r.DB).All(ctx)
	if err != nil {
		return err
	}
	withAncestors := expandMenuAncestors(all, menuIDs)
	expanded := expandMenuIDs(all, withAncestors)
	role := model.SysRole{ID: roleID}
	var menus []model.SysMenu
	if len(expanded) > 0 {
		if err := r.DB.WithContext(ctx).Find(&menus, expanded).Error; err != nil {
			return err
		}
	}
	return r.DB.WithContext(ctx).Model(&role).Association("Menus").Replace(menus)
}

// AppendMenuIDs 将菜单增量关联到角色（不覆盖已有分配）。
func (r *SysRoleRepository) AppendMenuIDs(ctx context.Context, roleID int64, menuIDs []int64) error {
	if len(menuIDs) == 0 {
		return nil
	}
	all, err := NewSysMenuRepository(r.DB).All(ctx)
	if err != nil {
		return err
	}
	expanded := expandMenuIDs(all, menuIDs)
	role := model.SysRole{ID: roleID}
	var menus []model.SysMenu
	if err := r.DB.WithContext(ctx).Find(&menus, expanded).Error; err != nil {
		return err
	}
	return r.DB.WithContext(ctx).Model(&role).Association("Menus").Append(menus)
}

type SysMenuRepository struct{ base.CRUD }

func NewSysMenuRepository(db *gorm.DB) *SysMenuRepository {
	return &SysMenuRepository{CRUD: base.CRUD{DB: db}}
}

func (r *SysMenuRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.SysMenu{})
}

func (r *SysMenuRepository) All(ctx context.Context) ([]model.SysMenu, error) {
	var rows []model.SysMenu
	err := r.DB.WithContext(ctx).Order("sort_order asc, id asc").Find(&rows).Error
	return rows, err
}

func (r *SysMenuRepository) GetByID(ctx context.Context, id any) (*model.SysMenu, error) {
	var row model.SysMenu
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysMenuRepository) Create(ctx context.Context, row *model.SysMenu) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysMenuRepository) Save(ctx context.Context, row *model.SysMenu) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysMenuRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysMenu{}, ids).Error
}

// ListForRoutes 对齐上游 MenuCRUD.list + ROLE_BASED 权限：超管全部菜单，否则仅角色直接关联的菜单
func (r *SysMenuRepository) ListForRoutes(ctx context.Context, user model.SysUser, super bool) ([]model.SysMenu, error) {
	if super {
		var rows []model.SysMenu
		err := r.DB.WithContext(ctx).Order("sort_order asc, id asc").Find(&rows).Error
		return rows, err
	}
	var roleIDs []int64
	for _, role := range user.Roles {
		if role.Status == "0" {
			roleIDs = append(roleIDs, role.ID)
		}
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var menuIDs []int64
	if err := r.DB.WithContext(ctx).Table("sys_role_menus").
		Where("sys_role_id IN ?", roleIDs).
		Distinct().
		Pluck("sys_menu_id", &menuIDs).Error; err != nil {
		return nil, err
	}
	if len(menuIDs) == 0 {
		return nil, nil
	}
	var rows []model.SysMenu
	err := r.DB.WithContext(ctx).Where("id IN ?", menuIDs).Order("sort_order asc, id asc").Find(&rows).Error
	return rows, err
}

func (r *SysMenuRepository) RouteMenusForRoles(ctx context.Context, roleIDs []int64, super bool) ([]model.SysMenu, error) {
	q := r.DB.WithContext(ctx).Where("status = ? AND hidden = ? AND type IN ?", "0", false, []int{1, 2})
	if super {
		var rows []model.SysMenu
		err := q.Order("sort_order asc, id asc").Find(&rows).Error
		return rows, err
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var linked []int64
	if err := r.DB.WithContext(ctx).Table("sys_role_menus").Where("sys_role_id IN ?", roleIDs).Distinct().Pluck("sys_menu_id", &linked).Error; err != nil {
		return nil, err
	}
	all, err := r.All(ctx)
	if err != nil {
		return nil, err
	}
	allow := expandAncestors(all, linked)
	var out []model.SysMenu
	for _, m := range all {
		if (m.Type == 1 || m.Type == 2) && allow[m.ID] {
			out = append(out, m)
		}
	}
	return out, nil
}

func expandAncestors(all []model.SysMenu, linked []int64) map[int64]bool {
	byID := map[int64]model.SysMenu{}
	for _, m := range all {
		byID[m.ID] = m
	}
	allow := map[int64]bool{}
	var mark func(id int64)
	mark = func(id int64) {
		if allow[id] {
			return
		}
		allow[id] = true
		if p, ok := byID[id]; ok && p.ParentID > 0 {
			mark(p.ParentID)
		}
	}
	for _, id := range linked {
		mark(id)
	}
	return allow
}

type SysDictRepository struct{ base.CRUD }

func NewSysDictRepository(db *gorm.DB) *SysDictRepository {
	return &SysDictRepository{CRUD: base.CRUD{DB: db}}
}

func (r *SysDictRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.SysDictType{}, &model.SysDictData{})
}

func (r *SysDictRepository) ListTypes(ctx context.Context, name, status string) ([]model.SysDictType, error) {
	q := r.DB.WithContext(ctx).Model(&model.SysDictType{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []model.SysDictType
	return rows, q.Order("sort_order asc, id asc").Find(&rows).Error
}

func (r *SysDictRepository) GetTypeByID(ctx context.Context, id any) (*model.SysDictType, error) {
	var row model.SysDictType
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysDictRepository) GetTypeByCode(ctx context.Context, code string) (*model.SysDictType, error) {
	var row model.SysDictType
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysDictRepository) CreateType(ctx context.Context, row *model.SysDictType) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysDictRepository) SaveType(ctx context.Context, row *model.SysDictType) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysDictRepository) DeleteTypes(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysDictType{}, ids).Error
}

func (r *SysDictRepository) ListData(ctx context.Context, pq PageQuery, typeID, label, status string) ([]model.SysDictData, int64, error) {
	return ListPage[model.SysDictData](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if typeID != "" {
			q = q.Where("type_id = ?", typeID)
		}
		if label != "" {
			q = q.Where("label LIKE ?", "%"+label+"%")
		}
		if status != "" {
			q = q.Where("status = ?", status)
		}
		return q
	})
}

func (r *SysDictRepository) GetDataByID(ctx context.Context, id any) (*model.SysDictData, error) {
	var row model.SysDictData
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysDictRepository) CreateData(ctx context.Context, row *model.SysDictData) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysDictRepository) SaveData(ctx context.Context, row *model.SysDictData) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysDictRepository) DeleteData(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysDictData{}, ids).Error
}

func (r *SysDictRepository) OptionsByCode(ctx context.Context, code string) ([]model.SysDictData, error) {
	if code == "common_status" {
		code = "STATUS"
	}
	t, err := r.GetTypeByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	var rows []model.SysDictData
	err = r.DB.WithContext(ctx).Where("type_id = ? AND status = ?", t.ID, "0").Order("sort_order asc, id asc").Find(&rows).Error
	return rows, err
}

type PhoneCallRepository struct{ base.CRUD }

func NewPhoneCallRepository(db *gorm.DB) *PhoneCallRepository {
	return &PhoneCallRepository{CRUD: base.CRUD{DB: db}}
}

func (r *PhoneCallRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.PhoneCallSession{})
}

func (r *PhoneCallRepository) GetByID(ctx context.Context, id any) (*model.PhoneCallSession, error) {
	var row model.PhoneCallSession
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PhoneCallRepository) Create(ctx context.Context, row *model.PhoneCallSession) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *PhoneCallRepository) Save(ctx context.Context, row *model.PhoneCallSession) error {
	return r.DB.WithContext(ctx).Save(row).Error
}
