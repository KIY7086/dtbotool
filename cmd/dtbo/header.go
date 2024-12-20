package dtbo

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// DTBO 头部结构
type DtboHeader struct {
	Magic           uint32
	TotalSize       uint32
	HeaderSize      uint32
	DtEntrySize     uint32
	DtEntryCount    uint32
	DtEntriesOffset uint32
	PageSize        uint32
	Version         uint32
}

// DTBO 设备树条目结构
type DtEntry struct {
	DtSize   uint32
	DtOffset uint32
	Id       uint32
	Rev      uint32
	Custom   [4]uint32
}

func printHeaderInfo(h *DtboHeader) {
	fmt.Printf("DTBO 头部信息:\n")
	fmt.Printf("  魔数: 0x%08X\n", h.Magic)
	fmt.Printf("  总大小: %d 字节\n", h.TotalSize)
	fmt.Printf("  头部大小: %d 字节\n", h.HeaderSize)
	fmt.Printf("  条目数量: %d\n", h.DtEntryCount)
	fmt.Printf("  版本: %d\n\n", h.Version)
}

// 验证并获取字节序
func verifyMagicAndGetEndian(data []byte) (binary.ByteOrder, *DtboHeader, error) {
	header := &DtboHeader{}
	reader := bytes.NewReader(data)

	// 尝试大端序
	if err := binary.Read(reader, binary.BigEndian, header); err == nil && header.Magic == 0xD7B7AB1E {
		return binary.BigEndian, header, nil
	}

	// 尝试小端序
	reader.Seek(0, 0)
	if err := binary.Read(reader, binary.LittleEndian, header); err == nil && header.Magic == 0xD7B7AB1E {
		return binary.LittleEndian, header, nil
	}

	return nil, nil, fmt.Errorf("无效的DTBO文件格式")
}
