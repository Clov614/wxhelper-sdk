// Package manager
// @Author Clover
// @Date 2025/1/8 下午4:13:00
// @Desc Cache Manager for handling files
package manager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"wxhelper-sdk/inner/utils"
)

const (
	defaultImgDir  = "/img"
	defaultFileDir = "/file"
)

var (
	ErrInvalidDate = errors.New("invalid date")
	ErrFileExists  = errors.New("file already exists")
)

type ICacheManager interface {
	Save(fileName string, isImg bool, data []byte) (*FileInfo, error)
	GetFilePathByFileName(fileName string) (string, error)
	GetDataByFileName(fileName string) ([]byte, error)
}

type FileName2FileInfo map[string]*FileInfo // FileName-to-FileInfo mapping

type FileInfo struct {
	FilePath string // Full file path
	FileName string // File name including extension
	FileExt  string // File extension
	IsImg    bool   // Indicates if the file is an image
}

// CacheManager is the implementation of ICacheManager
type CacheManager struct {
	mu                sync.RWMutex
	fileName2FileInfo FileName2FileInfo
}

var (
	cacheManager     *CacheManager
	cacheManagerOnce sync.Once
)

// newCacheManager creates a new CacheManager with the given cache size.
func newCacheManager(cacheSize int) *CacheManager {
	return &CacheManager{
		fileName2FileInfo: make(FileName2FileInfo, cacheSize),
	}
}

// GetCacheManager returns the singleton instance of CacheManager.
func GetCacheManager() ICacheManager {
	cacheManagerOnce.Do(func() {
		cacheManager = newCacheManager(30)
	})
	return cacheManager
}

// Save saves a file by its fileName and writes data to the file system.
func (cm *CacheManager) Save(fileName string, isImg bool, data []byte) (*FileInfo, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if the file already exists
	if _, exists := cm.fileName2FileInfo[fileName]; exists {
		return nil, fmt.Errorf("%w: %s", ErrFileExists, fileName)
	}

	// Generate file path
	filePath, err := utils.ConvertToWindows(fileName)
	if err != nil {
		return nil, fmt.Errorf("convert to windows file failed: %w", err)
	}

	// Write data to the file system
	if err := writeDataToFile(filePath, data); err != nil {
		return nil, fmt.Errorf("failed to write data to file: %w", err)
	}

	// Create and store FileInfo
	fileInfo := &FileInfo{
		FilePath: filePath,
		FileName: fileName,
		FileExt:  getFileExtension(fileName),
		IsImg:    isImg,
	}
	cm.fileName2FileInfo[fileName] = fileInfo

	return fileInfo, nil
}

// GetFilePathByFileName retrieves the file path by its file name.
func (cm *CacheManager) GetFilePathByFileName(fileName string) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fileInfo, exists := cm.fileName2FileInfo[fileName]
	if !exists {
		return "", os.ErrNotExist
	}
	return fileInfo.FilePath, nil
}

// GetDataByFileName retrieves the file data by its file name.
func (cm *CacheManager) GetDataByFileName(fileName string) ([]byte, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fileInfo, exists := cm.fileName2FileInfo[fileName]
	if !exists {
		return nil, os.ErrNotExist
	}

	// Read the data from the file
	data, err := os.ReadFile(fileInfo.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}
	return data, nil
}

// writeDataToFile writes data to the given file path.
func writeDataToFile(filePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

// getFileExtension extracts the file extension from the file name.
func getFileExtension(fileName string) string {
	if i := strings.LastIndex(fileName, "."); i >= 0 {
		return fileName[i+1:]
	}
	return ""
}
