// Package utils
// @Author Clover
// @Data 2025/1/8 下午4:05:00
// @Desc
package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/eatmoreapple/env"
	"os"
	"path/filepath"
)

// TempDir for go:linkname
func TempDir() string {
	return env.Name("TEMP_DIR").StringOrElse(os.TempDir())
}

// ConvertToWindows 转换为windows路径
func ConvertToWindows(filename string) (path string, err error) {
	return filepath.Join(TempDir(), filename), nil
}

// DecodeBase64 解码base64
func DecodeBase64(base64Str string) (data []byte, err error) {
	data, err = base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("DecodeImgBase64: %w", err)
	}
	return data, nil
}
