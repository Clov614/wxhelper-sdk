package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var i int = 0

// 创建临时目录并设置环境变量 TEMP_DIR
// 每次生成唯一的临时目录名
func setupTempDir(t *testing.T) string {
	i++
	t.Helper()
	// 生成唯一的临时目录名
	tempDir := filepath.Join(t.TempDir(), "testdata_"+time.Now().Format("20060102_150405")+fmt.Sprintf("_%d", i))
	t.Logf("tempDir: %s", tempDir)
	// 创建目录
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	os.Setenv("TEMP_DIR", tempDir)
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	return tempDir
}

// TestCacheManagerSave 测试 Save 方法
func TestCacheManagerSave(t *testing.T) {
	manager := GetCacheManager()
	_ = setupTempDir(t)

	// 测试保存文件
	fileName := "testfile1.txt"

	fileData := []byte("test data")
	fileInfo, err := manager.Save(fileName, false, fileData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 验证 FileInfo
	if fileInfo.FileName != fileName {
		t.Errorf("expected fileName %s, got %s", fileName, fileInfo.FileName)
	}
	if fileInfo.FileExt != "txt" {
		t.Errorf("expected fileExt txt, got %s", fileInfo.FileExt)
	}
}

// TestCacheManagerGetData 测试 GetDataByFileName 方法
func TestCacheManagerGetData(t *testing.T) {
	manager := GetCacheManager()
	_ = setupTempDir(t)

	// 保存测试文件
	fileName := "testfile2.txt"

	fileData := []byte("Hello, world!")
	_, err := manager.Save(fileName, false, fileData)

	// 测试读取数据
	data, err := manager.GetDataByFileName(fileName)
	if err != nil {
		t.Fatalf("GetDataByFileName failed: %v", err)
	}
	expectedData := "Hello, world!"
	if string(data) != expectedData {
		t.Errorf("expected data %s, got %s", expectedData, string(data))
	}
}

// TestCacheManagerFilePath 测试 GetFilePathByFileName 方法
func TestCacheManagerFilePath(t *testing.T) {
	manager := GetCacheManager()
	tempDir := setupTempDir(t)

	// 保存测试文件
	fileName := "testfile3.txt"
	fileData := []byte("test data")
	_, err := manager.Save(fileName, false, fileData)

	// 测试获取文件路径
	filePath, err := manager.GetFilePathByFileName(fileName)
	if err != nil {
		t.Fatalf("GetFilePathByFileName failed: %v", err)
	}
	// 期望路径应该包括临时目录和文件名
	expectedPath := filepath.Join(tempDir, fileName)
	if filePath != expectedPath {
		t.Errorf("expected filePath %s, got %s", expectedPath, filePath)
	}
}
