package storage

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/model"
)

type UploadResult struct {
	ObjectKey string
	RawURL    string
	PublicURL string
	Size      int64
}

type Engine struct {
	cfg     *config.Config
	storage *model.SysStorage
}

func NewEngine(cfg *config.Config, storage *model.SysStorage) *Engine {
	return &Engine{cfg: cfg, storage: storage}
}

func ReplaceURLDomain(rawURL, baseURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	baseURL = strings.TrimSpace(baseURL)
	if rawURL == "" || baseURL == "" {
		return rawURL
	}
	raw, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return rawURL
	}
	raw.Scheme = base.Scheme
	raw.Host = base.Host
	return raw.String()
}

func BuildPublicURL(storage *model.SysStorage, rawURL string) string {
	if storage == nil {
		return rawURL
	}
	if storage.Type == model.StorageTypeOSS && strings.TrimSpace(storage.BaseURL) != "" {
		return ReplaceURLDomain(rawURL, storage.BaseURL)
	}
	if strings.TrimSpace(storage.Domain) != "" {
		return JoinURL(storage.Domain, objectPathFromURL(rawURL, storage))
	}
	return rawURL
}

func JoinURL(prefix, objectPath string) string {
	prefix = strings.TrimSpace(prefix)
	objectPath = strings.TrimSpace(objectPath)
	if prefix == "" {
		return objectPath
	}
	if objectPath == "" {
		return strings.TrimRight(prefix, "/")
	}
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(objectPath, "/")
}

func objectPathFromURL(rawURL string, storage *model.SysStorage) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return strings.TrimLeft(rawURL, "/")
	}
	p := strings.TrimLeft(u.Path, "/")
	if storage != nil && storage.Type == model.StorageTypeOSS && storage.BucketName != "" {
		prefix := storage.BucketName + "/"
		if strings.HasPrefix(p, prefix) {
			p = strings.TrimPrefix(p, prefix)
		}
	}
	return p
}

func NormalizeParentPath(parentPath string) string {
	parentPath = strings.TrimSpace(parentPath)
	if parentPath == "" || parentPath == "/" {
		return "/"
	}
	if !strings.HasPrefix(parentPath, "/") {
		parentPath = "/" + parentPath
	}
	return strings.TrimRight(parentPath, "/")
}

func BuildRelativePath(parentPath, fileName string) string {
	parentPath = NormalizeParentPath(parentPath)
	if parentPath == "/" {
		return "/" + strings.TrimLeft(fileName, "/")
	}
	return parentPath + "/" + strings.TrimLeft(fileName, "/")
}

func UniqueFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	base := strings.TrimSuffix(filepath.Base(originalName), ext)
	if base == "" {
		base = "file"
	}
	return fmt.Sprintf("%s_%s%s", base, strings.ReplaceAll(uuid.NewString(), "-", ""), ext)
}

func DefaultParentPath() string {
	return "/" + time.Now().Format("2006/01/02")
}

func (e *Engine) Upload(ctx context.Context, parentPath, originalName string, size int64, reader io.Reader, contentType string) (*UploadResult, error) {
	if e.storage.Type == model.StorageTypeLocal {
		return e.uploadLocal(parentPath, originalName, size, reader)
	}
	return e.uploadOSS(ctx, parentPath, originalName, size, reader, contentType)
}

func (e *Engine) DeleteObject(ctx context.Context, objectKey string) error {
	objectKey = strings.TrimLeft(objectKey, "/")
	if e.storage.Type == model.StorageTypeLocal {
		root := e.localRoot()
		full := filepath.Join(root, filepath.FromSlash(objectKey))
		return os.Remove(full)
	}
	client, bucket, err := e.ossClient()
	if err != nil {
		return err
	}
	return client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
}

func (e *Engine) LocalRoot() string {
	return e.localRoot()
}

func (e *Engine) localRoot() string {
	root := strings.TrimSpace(e.storage.BucketName)
	if root == "" {
		root = e.cfg.UploadDir
	}
	if !filepath.IsAbs(root) {
		root = filepath.Clean(root)
	}
	return root
}

func (e *Engine) uploadLocal(parentPath, originalName string, size int64, reader io.Reader) (*UploadResult, error) {
	parentPath = NormalizeParentPath(parentPath)
	if parentPath == "/" {
		parentPath = DefaultParentPath()
	}
	storedName := UniqueFileName(originalName)
	relativePath := BuildRelativePath(parentPath, storedName)
	objectKey := strings.TrimLeft(relativePath, "/")

	root := e.localRoot()
	fullPath := filepath.Join(root, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	written, err := io.Copy(f, reader)
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, err
	}
	if size <= 0 {
		size = written
	}

	rawURL := JoinURL(e.localDomain(), objectKey)
	publicURL := BuildPublicURL(e.storage, rawURL)
	return &UploadResult{
		ObjectKey: objectKey,
		RawURL:    rawURL,
		PublicURL: publicURL,
		Size:      size,
	}, nil
}

func (e *Engine) localDomain() string {
	if strings.TrimSpace(e.storage.Domain) != "" {
		return strings.TrimRight(strings.TrimSpace(e.storage.Domain), "/")
	}
	return strings.TrimRight(e.cfg.PublicBaseURL, "/") + strings.TrimRight(e.cfg.StaticURLPrefix, "/") + "/" + e.storage.Code
}

func (e *Engine) uploadOSS(ctx context.Context, parentPath, originalName string, size int64, reader io.Reader, contentType string) (*UploadResult, error) {
	parentPath = NormalizeParentPath(parentPath)
	if parentPath == "/" {
		parentPath = DefaultParentPath()
	}
	storedName := UniqueFileName(originalName)
	relativePath := BuildRelativePath(parentPath, storedName)
	objectKey := strings.TrimLeft(relativePath, "/")

	client, bucket, err := e.ossClient()
	if err != nil {
		return nil, err
	}
	opts := minio.PutObjectOptions{ContentType: contentType}
	if _, err := client.PutObject(ctx, bucket, objectKey, reader, size, opts); err != nil {
		return nil, err
	}

	rawURL := e.buildOSSRawURL(bucket, objectKey)
	publicURL := BuildPublicURL(e.storage, rawURL)
	return &UploadResult{
		ObjectKey: objectKey,
		RawURL:    rawURL,
		PublicURL: publicURL,
		Size:      size,
	}, nil
}

func (e *Engine) buildOSSRawURL(bucket, objectKey string) string {
	if strings.TrimSpace(e.storage.Domain) != "" {
		return JoinURL(e.storage.Domain, objectKey)
	}
	endpoint := strings.TrimSpace(e.storage.Endpoint)
	if endpoint == "" {
		return JoinURL("", objectKey)
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return JoinURL(endpoint, bucket+"/"+objectKey)
	}
	host := u.Hostname()
	if host == "" {
		host = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	}
	if isIPAddress(host) {
		return fmt.Sprintf("%s://%s/%s/%s", schemeOf(u, endpoint), host, bucket, objectKey)
	}
	return fmt.Sprintf("%s://%s.%s/%s", schemeOf(u, endpoint), bucket, host, objectKey)
}

func schemeOf(u *url.URL, endpoint string) string {
	if u != nil && u.Scheme != "" {
		return u.Scheme
	}
	if strings.HasPrefix(strings.ToLower(endpoint), "http://") {
		return "http"
	}
	return "https"
}

func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

func isCOSEndpoint(host string) bool {
	h := strings.ToLower(host)
	return strings.Contains(h, ".myqcloud.com") || strings.Contains(h, ".tencentcos.cn")
}

func isAliyunOSSEndpoint(host string) bool {
	return strings.Contains(strings.ToLower(host), "aliyuncs.com")
}

// cosRegionFromHost 从 cos.ap-guangzhou.myqcloud.com 解析 ap-guangzhou
func cosRegionFromHost(host string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(host)), ".")
	for i, part := range parts {
		if part == "cos" && i+1 < len(parts) && parts[i+1] != "" {
			return parts[i+1]
		}
	}
	return ""
}

// normalizeOSSHost 若 Endpoint 误填为 bucket.cos.region.myqcloud.com，去掉 bucket 前缀
func normalizeOSSHost(host, bucket string) string {
	host = strings.TrimSpace(host)
	bucket = strings.TrimSpace(bucket)
	if bucket != "" && strings.HasPrefix(strings.ToLower(host), strings.ToLower(bucket)+".") {
		return host[len(bucket)+1:]
	}
	return host
}

func ossBucketLookup(host string) minio.BucketLookupType {
	if isIPAddress(host) {
		return minio.BucketLookupPath
	}
	if isCOSEndpoint(host) || isAliyunOSSEndpoint(host) {
		return minio.BucketLookupDNS
	}
	return minio.BucketLookupAuto
}

func ossRegion(host string) string {
	if r := cosRegionFromHost(host); r != "" {
		return r
	}
	return "us-east-1"
}

func (e *Engine) ossClient() (*minio.Client, string, error) {
	endpoint := strings.TrimSpace(e.storage.Endpoint)
	if endpoint == "" {
		return nil, "", fmt.Errorf("Endpoint 不能为空")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", fmt.Errorf("Endpoint 格式不正确")
	}
	host := u.Host
	if host == "" {
		host = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	}
	secure := u.Scheme == "https" || !strings.HasPrefix(strings.ToLower(endpoint), "http://")
	bucket := strings.TrimSpace(e.storage.BucketName)
	if bucket == "" {
		return nil, "", fmt.Errorf("Bucket 不能为空")
	}
	host = normalizeOSSHost(host, bucket)
	client, err := minio.New(host, &minio.Options{
		Creds:        credentials.NewStaticV4(e.storage.AccessKey, e.storage.SecretKey, ""),
		Secure:       secure,
		Region:       ossRegion(host),
		BucketLookup: ossBucketLookup(host),
	})
	if err != nil {
		return nil, "", err
	}
	return client, bucket, nil
}

func EnsureLocalDir(root string) error {
	return os.MkdirAll(root, 0o755)
}

func ExtName(name string) string {
	ext := strings.TrimPrefix(strings.ToLower(path.Ext(name)), ".")
	return ext
}
