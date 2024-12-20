package dtbo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func UnpackDtbo(dtboFile, outDir string) error {
	data, err := os.ReadFile(dtboFile)
	if err != nil {
		return fmt.Errorf("读取DTBO文件失败: %v", err)
	}

	endian, header, err := verifyMagicAndGetEndian(data)
	if err != nil {
		return err
	}

	printHeaderInfo(header)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	for i := uint32(0); i < header.DtEntryCount; i++ {
		if err := extractDtEntry(data, header, i, outDir, endian); err != nil {
			fmt.Printf("警告: 处理设备树条目 %d 时出错: %v\n", i, err)
			continue
		}
	}

	fmt.Printf("完成！已提取 %d 个设备树到目录: %s\n", header.DtEntryCount, outDir)
	return nil
}

func extractDtEntry(data []byte, header *DtboHeader, index uint32, outDir string, endian binary.ByteOrder) error {
	entryOffset := header.DtEntriesOffset + (index * header.DtEntrySize)
	entry := &DtEntry{}

	if err := binary.Read(bytes.NewReader(data[entryOffset:]), endian, entry); err != nil {
		return fmt.Errorf("解析设备树条目失败: %v", err)
	}

	if entry.DtOffset+entry.DtSize > uint32(len(data)) {
		return fmt.Errorf("设备树条目范围无效")
	}

	dtbData := data[entry.DtOffset : entry.DtOffset+entry.DtSize]
	outFile := filepath.Join(outDir, fmt.Sprintf("dtbo_%d.dtb", index))

	if err := os.WriteFile(outFile, dtbData, 0644); err != nil {
		return fmt.Errorf("保存设备树文件失败: %v", err)
	}

	fmt.Printf("已提取设备树 %d:\n", index)
	fmt.Printf("  ID: %d, 版本: %d\n", entry.Id, entry.Rev)
	fmt.Printf("  大小: %d 字节\n", entry.DtSize)
	fmt.Printf("  输出: %s\n\n", outFile)

	return nil
}
