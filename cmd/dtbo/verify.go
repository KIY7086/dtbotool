package dtbo

import (
	"fmt"
	"os"
)

func VerifyDtbo(dtboFile string) error {
	data, err := os.ReadFile(dtboFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	_, header, err := verifyMagicAndGetEndian(data)
	if err != nil {
		return err
	}

	if header.TotalSize == 0 || header.TotalSize > uint32(len(data)) {
		return fmt.Errorf("文件大小无效")
	}

	if header.DtEntryCount == 0 {
		return fmt.Errorf("未包含任何设备树")
	}

	return nil
}

func RestoreDtbo(backupFile, outputFile string) error {
	data, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	_, header, err := verifyMagicAndGetEndian(data)
	if err != nil {
		return err
	}

	if uint32(len(data)) < header.TotalSize {
		return fmt.Errorf("备份文件大小不正确")
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("写入恢复文件失败: %v", err)
	}

	return nil
}
