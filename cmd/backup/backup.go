package backup

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const BACKUP_DIR_NAME = "dtbotool-backups"

// 获取备份目录的跨平台实现
func GetBackupDir() (string, error) {
	var baseDir string

	switch {
	case os.Getenv("XDG_CACHE_HOME") != "":
		baseDir = os.Getenv("XDG_CACHE_HOME")
	case os.Getenv("HOME") != "":
		baseDir = filepath.Join(os.Getenv("HOME"), ".cache")
	case os.Getenv("LOCALAPPDATA") != "": // Windows
		baseDir = os.Getenv("LOCALAPPDATA")
	case os.Getenv("TEMP") != "":
		baseDir = os.Getenv("TEMP")
	default:
		return "", fmt.Errorf("无法确定备份目录位置")
	}

	backupDir := filepath.Join(baseDir, BACKUP_DIR_NAME)
	return backupDir, os.MkdirAll(backupDir, 0755)
}

// 创建备份
func CreateBackup(inputFile string) (string, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return "", err
	}

	// 读取输入文件
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 生成唯一的备份文件名
	timestamp := time.Now().Format("20060102_150405")
	hash := md5.Sum(inputData)
	hashStr := hex.EncodeToString(hash[:])[:8]
	backupFile := filepath.Join(backupDir,
		fmt.Sprintf("%s_%s_%s.bak", filepath.Base(inputFile), timestamp, hashStr))

	// 复制���件
	if err := copyFile(inputFile, backupFile); err != nil {
		return "", fmt.Errorf("创建备份失败: %v", err)
	}

	return backupFile, nil
}

// 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
