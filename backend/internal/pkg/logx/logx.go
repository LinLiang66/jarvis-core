package logx

import (
	"encoding/json"
	"log"
	"strings"
)

// Level 应用日志级别（debug 最详细）
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var current = LevelInfo

// Init 从环境变量解析：debug | info | warn | error
func Init(level string) {
	current = parseLevel(level)
	log.Printf("[logx] 日志级别=%s", levelName(current))
}

func parseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug", "trace":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error", "fatal":
		return LevelError
	default:
		return LevelInfo
	}
}

func levelName(l Level) string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

// IsDebug 当前是否为 debug 级别
func IsDebug() bool {
	return current <= LevelDebug
}

func Debugf(format string, args ...any) {
	if current > LevelDebug {
		return
	}
	log.Printf("[DEBUG] "+format, args...)
}

func Infof(format string, args ...any) {
	if current > LevelInfo {
		return
	}
	log.Printf("[INFO] "+format, args...)
}

func Warnf(format string, args ...any) {
	if current > LevelWarn {
		return
	}
	log.Printf("[WARN] "+format, args...)
}

func Errorf(format string, args ...any) {
	if current > LevelError {
		return
	}
	log.Printf("[ERROR] "+format, args...)
}

// JSON 格式化对象为 JSON 字符串
func JSON(v any) string {
	if v == nil {
		return "null"
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}
