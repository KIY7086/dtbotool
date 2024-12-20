package dtbo

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func PackDtbo(dtbDir, dtboFile string) error {
	files, err := os.ReadDir(dtbDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	var dtbFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dtb") {
			dtbFiles = append(dtbFiles, filepath.Join(dtbDir, file.Name()))
		}
	}

	if len(dtbFiles) == 0 {
		return fmt.Errorf("未找到DTB文件")
	}

	header := &DtboHeader{
		Magic:        0xD7B7AB1E,
		HeaderSize:   uint32(binary.Size(DtboHeader{})),
		DtEntrySize:  uint32(binary.Size(DtEntry{})),
		DtEntryCount: uint32(len(dtbFiles)),
		PageSize:     4096,
		Version:      1,
	}

	header.DtEntriesOffset = header.HeaderSize
	currentOffset := header.DtEntriesOffset + header.DtEntrySize*header.DtEntryCount

	outFile, err := os.Create(dtboFile)
	if err != nil {
		return fmt.Errorf("创建DTBO文件失败: %v", err)
	}
	defer outFile.Close()

	if err := binary.Write(outFile, binary.BigEndian, header); err != nil {
		return fmt.Errorf("写入头部失败: %v", err)
	}

	entries := make([]DtEntry, len(dtbFiles))
	for i, dtbFile := range dtbFiles {
		dtbData, err := os.ReadFile(dtbFile)
		if err != nil {
			return fmt.Errorf("读取DTB文件失败: %v", err)
		}

		entries[i] = DtEntry{
			DtSize:   uint32(len(dtbData)),
			DtOffset: currentOffset,
			Id:       uint32(i),
			Rev:      1,
		}

		if err := binary.Write(outFile, binary.BigEndian, entries[i]); err != nil {
			return fmt.Errorf("写入条目失败: %v", err)
		}

		currentOffset += uint32(len(dtbData))
	}

	for _, dtbFile := range dtbFiles {
		dtbData, err := os.ReadFile(dtbFile)
		if err != nil {
			return fmt.Errorf("读取DTB文件失败: %v", err)
		}

		if _, err := outFile.Write(dtbData); err != nil {
			return fmt.Errorf("写入DTB数据失败: %v", err)
		}
	}

	header.TotalSize = currentOffset
	outFile.Seek(0, 0)
	if err := binary.Write(outFile, binary.BigEndian, header); err != nil {
		return fmt.Errorf("更新头部失败: %v", err)
	}

	fmt.Printf("已成功打包 %d 个DTB文件到: %s\n", len(dtbFiles), dtboFile)
	return nil
}
