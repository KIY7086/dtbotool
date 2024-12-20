package recovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiy7086/dtbotool/cmd/backup"
	"github.com/kiy7086/dtbotool/cmd/dtbo"
)

// 列出所有备份
func ListBackups() error {
	backupDir, err := backup.GetBackupDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("无法读取备份")
	}

	if len(files) == 0 {
		fmt.Println("没有找到任何备份")
		return nil
	}

	fmt.Println("可用的备份文件:")
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".bak") {
			info, _ := file.Info()
			origName := strings.Split(file.Name(), "_")[0]
			origName = strings.TrimSuffix(origName, ".bak")
			fmt.Printf("- %s (备份时间: %s)\n",
				origName,
				info.ModTime().Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

// 删除所有备份
func PurgeBackups() error {
	backupDir, err := backup.GetBackupDir()
	if err != nil {
		return err
	}

	return os.RemoveAll(backupDir)
}

// 恢复备份
func RestoreBackup() error {
	backupDir, err := backup.GetBackupDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("无法读取备份")
	}

	var backups []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".bak") {
			backups = append(backups, file)
		}
	}

	if len(backups) == 0 {
		return fmt.Errorf("没有找到可用的备份")
	}

	fmt.Println("\n可用的备份:")
	for i, file := range backups {
		info, _ := file.Info()
		origName := strings.Split(file.Name(), "_")[0]
		origName = strings.TrimSuffix(origName, ".bak")
		fmt.Printf("[%d] %s (备份时间: %s)\n",
			i+1, origName,
			info.ModTime().Format("2006-01-02 15:04:05"))
	}

	fmt.Print("\n请选择要恢复的备份编号 (输入 0 取消): ")
	var choice int
	fmt.Scanln(&choice)

	if choice <= 0 || choice > len(backups) {
		return fmt.Errorf("操作已取消")
	}

	selectedFile := backups[choice-1]
	origName := strings.Split(selectedFile.Name(), "_")[0]
	if !strings.HasSuffix(origName, ".img") {
		origName = strings.TrimSuffix(origName, ".bak") + ".img"
	} else {
		origName = strings.TrimSuffix(origName, ".bak")
	}

	fmt.Printf("\n即将恢复备份到: %s\n", origName)
	fmt.Print("确认恢复? [y/N] ")

	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" {
		return fmt.Errorf("操作已取消")
	}

	backupFile := filepath.Join(backupDir, selectedFile.Name())
	if err := dtbo.RestoreDtbo(backupFile, origName); err != nil {
		return fmt.Errorf("恢复失败: %v", err)
	}

	fmt.Printf("\n备份已恢复到: %s\n", origName)
	fmt.Println("\n如需刷入设备，请确保设备已进入 fastboot 模式，然后执行:")
	fmt.Printf("fastboot flash dtbo %s\n", origName)

	return nil
}
