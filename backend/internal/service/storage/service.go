package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/model"
	"jarvis-core/backend/internal/store"
)

var (
	ErrStorageNotFound = errors.New("存储不存在")
	ErrNoDefault       = errors.New("请先配置默认存储")
)

type Service struct {
	cfg    *config.Config
	stores *store.Stores
}

func NewService(cfg *config.Config, stores *store.Stores) *Service {
	return &Service{cfg: cfg, stores: stores}
}

func (s *Service) ResolveStorage(ctx context.Context, storageID int64) (*model.SysStorage, error) {
	if storageID > 0 {
		row, err := s.stores.SysStorage.GetByID(ctx, storageID)
		if err != nil {
			return nil, ErrStorageNotFound
		}
		if row.Status != "0" {
			return nil, fmt.Errorf("存储 [%s] 已禁用", row.Name)
		}
		return row, nil
	}
	row, err := s.stores.SysStorage.GetDefault(ctx)
	if err != nil {
		return nil, ErrNoDefault
	}
	return row, nil
}

func (s *Service) Upload(ctx context.Context, storageID int64, parentPath, originalName string, size int64, reader io.Reader, contentType string) (*model.SysFile, error) {
	st, err := s.ResolveStorage(ctx, storageID)
	if err != nil {
		return nil, err
	}
	parentPath = NormalizeParentPath(parentPath)
	if parentPath == "/" {
		parentPath = DefaultParentPath()
	}
	if err := s.ensureParentDirs(ctx, st, parentPath); err != nil {
		return nil, err
	}

	prepared, err := PrepareUploadContent(s.cfg, originalName, contentType, reader, size)
	if err != nil {
		return nil, err
	}

	engine := NewEngine(s.cfg, st)
	result, err := engine.Upload(ctx, parentPath, prepared.FileName, prepared.Size, prepared.Reader, prepared.ContentType)
	if err != nil {
		return nil, err
	}

	row := &model.SysFile{
		StorageID:    st.ID,
		Name:         filepathBase(result.ObjectKey),
		OriginalName: originalName,
		Path:         "/" + strings.TrimLeft(result.ObjectKey, "/"),
		ParentPath:   parentPath,
		URL:          result.PublicURL,
		Size:         result.Size,
		Extension:    ExtName(prepared.FileName),
		ContentType:  prepared.ContentType,
		Type:         1,
	}
	if err := s.stores.SysFile.Create(ctx, row); err != nil {
		_ = engine.DeleteObject(ctx, result.ObjectKey)
		return nil, err
	}
	return row, nil
}

func (s *Service) CreateDir(ctx context.Context, storageID int64, parentPath, name string) (*model.SysFile, error) {
	st, err := s.ResolveStorage(ctx, storageID)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("文件夹名称不能为空")
	}
	parentPath = NormalizeParentPath(parentPath)
	if parentPath == "/" {
		return nil, fmt.Errorf("上级目录不能为空")
	}
	if err := s.ensureParentDirs(ctx, st, parentPath); err != nil {
		return nil, err
	}
	exists, err := s.stores.SysFile.ExistsDir(ctx, st.ID, parentPath, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("文件夹已存在")
	}
	row := &model.SysFile{
		StorageID:    st.ID,
		Name:         name,
		OriginalName: name,
		Path:         BuildRelativePath(parentPath, name),
		ParentPath:   parentPath,
		Type:         0,
	}
	if err := s.stores.SysFile.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) DeleteFiles(ctx context.Context, ids []string) error {
	for _, id := range ids {
		row, err := s.stores.SysFile.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if row.Type == 0 {
			if err := s.deleteDirTree(ctx, row); err != nil {
				return err
			}
			continue
		}
		if err := s.deleteFileRecord(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) deleteDirTree(ctx context.Context, dir *model.SysFile) error {
	st, err := s.stores.SysStorage.GetByID(ctx, dir.StorageID)
	if err != nil {
		return err
	}
	rows, err := s.stores.SysFile.ListByPathPrefix(ctx, dir.StorageID, dir.Path)
	if err != nil {
		return err
	}
	engine := NewEngine(s.cfg, st)
	for _, row := range rows {
		if row.Type != 1 {
			continue
		}
		if err := engine.DeleteObject(ctx, strings.TrimLeft(row.Path, "/")); err != nil && !osIsNotExist(err) {
			return err
		}
	}
	if st.Type == model.StorageTypeLocal {
		s.cleanupLocalDirPrefix(engine, dir.Path)
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, strconv.FormatInt(row.ID, 10))
	}
	return s.stores.SysFile.DeleteByIDs(ctx, ids)
}

func (s *Service) cleanupLocalDirPrefix(engine *Engine, dirPath string) {
	root := engine.LocalRoot()
	prefix := strings.Trim(strings.TrimSpace(dirPath), "/")
	if prefix == "" {
		return
	}
	full := filepath.Join(root, filepath.FromSlash(prefix))
	_ = os.RemoveAll(full)
}

func (s *Service) deleteFileRecord(ctx context.Context, row *model.SysFile) error {
	st, err := s.stores.SysStorage.GetByID(ctx, row.StorageID)
	if err != nil {
		return err
	}
	engine := NewEngine(s.cfg, st)
	if err := engine.DeleteObject(ctx, strings.TrimLeft(row.Path, "/")); err != nil && !osIsNotExist(err) {
		return err
	}
	return s.stores.SysFile.DeleteByIDs(ctx, []string{strconv.FormatInt(row.ID, 10)})
}

func (s *Service) ensureParentDirs(ctx context.Context, st *model.SysStorage, parentPath string) error {
	parentPath = NormalizeParentPath(parentPath)
	if parentPath == "/" {
		return nil
	}
	parts := strings.Split(strings.Trim(parentPath, "/"), "/")
	current := "/"
	for _, part := range parts {
		if part == "" {
			continue
		}
		next := BuildRelativePath(current, part)
		exists, err := s.stores.SysFile.ExistsDir(ctx, st.ID, current, part)
		if err != nil {
			return err
		}
		if !exists {
			row := &model.SysFile{
				StorageID:    st.ID,
				Name:         part,
				OriginalName: part,
				Path:         next,
				ParentPath:   current,
				Type:         0,
			}
			if err := s.stores.SysFile.Create(ctx, row); err != nil {
				return err
			}
		}
		current = next
	}
	return nil
}

func filepathBase(objectKey string) string {
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")
	if i := strings.LastIndex(objectKey, "/"); i >= 0 {
		return objectKey[i+1:]
	}
	return objectKey
}

func osIsNotExist(err error) bool {
	return os.IsNotExist(err)
}
