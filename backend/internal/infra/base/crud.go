package base

import "gorm.io/gorm"

// CRUD 参考 tiny-pro-go/impl/baseImpl.go 的通用 GORM 操作。
type CRUD struct {
	DB *gorm.DB
}

func (b CRUD) Create(model any) error {
	return b.DB.Create(model).Error
}

func (b CRUD) Save(model any) error {
	return b.DB.Save(model).Error
}

func (b CRUD) Delete(model any) error {
	return b.DB.Delete(model).Error
}

func (b CRUD) First(dest any, conds ...any) error {
	return b.DB.First(dest, conds...).Error
}
