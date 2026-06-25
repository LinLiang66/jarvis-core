package storage

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"jarvis-core/backend/internal/config"
)

type PreparedUpload struct {
	Reader      io.Reader
	Size        int64
	ContentType string
	FileName    string
}

// PrepareUploadContent 对图片做智能压缩，本地与 OSS 上传共用。
func PrepareUploadContent(cfg *config.Config, fileName, contentType string, reader io.Reader, size int64) (*PreparedUpload, error) {
	out := &PreparedUpload{
		Reader:      reader,
		Size:        size,
		ContentType: contentType,
		FileName:    fileName,
	}
	if !cfg.ImageCompressEnable {
		return out, nil
	}
	if size <= 0 || size > int64(cfg.ImageCompressMaxInput) {
		return out, nil
	}
	if size < int64(cfg.ImageCompressMinBytes) {
		return out, nil
	}
	ct := normalizeImageContentType(contentType, fileName)
	if !isCompressibleImage(ct) {
		return out, nil
	}

	data, err := io.ReadAll(io.LimitReader(reader, int64(cfg.ImageCompressMaxInput)+1))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > cfg.ImageCompressMaxInput {
		return out, nil
	}

	compressed, newCT, changed := compressImage(data, ct, cfg)
	if !changed {
		out.Reader = bytes.NewReader(data)
		out.Size = int64(len(data))
		return out, nil
	}

	newName := fileName
	if newCT == "image/jpeg" && !strings.EqualFold(filepath.Ext(fileName), ".jpg") && !strings.EqualFold(filepath.Ext(fileName), ".jpeg") {
		newName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".jpg"
	}
	out.Reader = bytes.NewReader(compressed)
	out.Size = int64(len(compressed))
	out.ContentType = newCT
	out.FileName = newName
	return out, nil
}

func normalizeImageContentType(contentType, fileName string) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ct != "" && ct != "application/octet-stream" {
		return ct
	}
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	default:
		return ct
	}
}

func isCompressibleImage(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/bmp":
		return true
	default:
		return false
	}
}

func compressImage(data []byte, contentType string, cfg *config.Config) ([]byte, string, bool) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, contentType, false
	}

	maxDim := cfg.ImageCompressMaxDim
	if maxDim <= 0 {
		maxDim = 1920
	}
	bounds := img.Bounds()
	if bounds.Dx() > maxDim || bounds.Dy() > maxDim {
		img = imaging.Fit(img, maxDim, maxDim, imaging.Lanczos)
	}

	quality := cfg.ImageCompressQuality
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	usePNG := format == "png" && imageHasAlpha(img)
	var buf bytes.Buffer
	outCT := "image/jpeg"
	if usePNG {
		if err := png.Encode(&buf, img); err != nil {
			return data, contentType, false
		}
		outCT = "image/png"
	} else {
		if err := jpeg.Encode(&buf, toRGBA(img), &jpeg.Options{Quality: quality}); err != nil {
			return data, contentType, false
		}
	}
	if buf.Len() >= len(data) {
		return data, contentType, false
	}
	return buf.Bytes(), outCT, true
}

func toRGBA(img image.Image) image.Image {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	return imaging.Clone(img)
}

func imageHasAlpha(img image.Image) bool {
	switch raw := img.(type) {
	case *image.NRGBA:
		return raw.Opaque() == false
	case *image.NRGBA64:
		return raw.Opaque() == false
	case *image.RGBA:
		return raw.Opaque() == false
	case *image.RGBA64:
		return raw.Opaque() == false
	default:
		return false
	}
}
