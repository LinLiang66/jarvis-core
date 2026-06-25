package filecategory

import "strings"

const (
	Other = 1
	Image = 2
	Doc   = 3
	Video = 4
	Audio = 5
)

var (
	imageExtensions = map[string]struct{}{
		"jpg": {}, "jpeg": {}, "png": {}, "gif": {}, "bmp": {}, "webp": {},
	}
	docExtensions = map[string]struct{}{
		"ppt": {}, "pptx": {}, "doc": {}, "docx": {}, "xls": {}, "xlsx": {},
		"txt": {}, "pdf": {}, "html": {}, "css": {}, "js": {},
	}
	videoExtensions = map[string]struct{}{
		"mp4": {}, "avi": {}, "mov": {}, "wmv": {}, "mkv": {}, "flv": {},
	}
	audioExtensions = map[string]struct{}{
		"mp3": {}, "wav": {}, "flac": {}, "aac": {}, "ogg": {}, "m4a": {},
	}
)

func ExtensionsByCategory(category int) []string {
	var set map[string]struct{}
	switch category {
	case Image:
		set = imageExtensions
	case Doc:
		set = docExtensions
	case Video:
		set = videoExtensions
	case Audio:
		set = audioExtensions
	default:
		return nil
	}
	out := make([]string, 0, len(set))
	for ext := range set {
		out = append(out, ext)
	}
	return out
}

func KnownExtensions() []string {
	known := make(map[string]struct{})
	for _, set := range []map[string]struct{}{imageExtensions, docExtensions, videoExtensions, audioExtensions} {
		for ext := range set {
			known[ext] = struct{}{}
		}
	}
	out := make([]string, 0, len(known))
	for ext := range known {
		out = append(out, ext)
	}
	return out
}

func Normalize(ext string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
}
