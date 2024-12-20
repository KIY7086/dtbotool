package unpack

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kiy7086/dtbotool/cmd/backup"
	"github.com/kiy7086/dtbotool/cmd/dtb"
	"github.com/kiy7086/dtbotool/cmd/dtbo"
)

// HandleUnpack 处理解包操作
func HandleUnpack(input, output string, rawOutput bool) error {
	// 检查输入文件/目录是否存在
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("'%s' 不存在", input)
	}

	switch {
	case strings.HasSuffix(input, ".dtbo"), strings.HasSuffix(input, ".img"):
		return handleDtboUnpack(input, output, rawOutput)
	case strings.HasSuffix(input, ".dtb"):
		return handleDtbUnpack(input, output)
	default:
		return handleDirUnpack(input)
	}
}

func handleDtboUnpack(input, output string, rawOutput bool) error {
	_, err := backup.CreateBackup(input)
	if err != nil {
		fmt.Printf("警告: 备份失败: %v\n", err)
		fmt.Printf("是否继续? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("操作已取消")
		}
	}

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "dtbo_*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if rawOutput {
		return handleRawDtboUnpack(input, output)
	}
	return handleDtsDtboUnpack(input, output, tmpDir)
}

func handleRawDtboUnpack(input, output string) error {
	outDir := output
	if outDir == "" {
		inputData, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("读取文件失败: %v", err)
		}
		timestamp := time.Now().Format("20060102_150405")
		hash := md5.Sum(inputData)
		hashStr := hex.EncodeToString(hash[:])[:8]
		outDir = fmt.Sprintf("dtbo_extracted_%s_%s", timestamp, hashStr)
	}
	return dtbo.UnpackDtbo(input, outDir)
}

func handleDtsDtboUnpack(input, output string, tmpDir string) error {
	fmt.Printf("正在解析DTBO文件...\n")
	if err := dtbo.UnpackDtbo(input, tmpDir); err != nil {
		return fmt.Errorf("解析DTBO失败: %v", err)
	}

	outDir := output
	if outDir == "" {
		inputData, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("读取文件失败: %v", err)
		}
		timestamp := time.Now().Format("20060102_150405")
		hash := md5.Sum(inputData)
		hashStr := hex.EncodeToString(hash[:])[:8]
		outDir = fmt.Sprintf("dtbo_decompiled_%s_%s", timestamp, hashStr)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	fmt.Printf("\n开始反编译...\n")
	if err := decompileDtbFiles(tmpDir, outDir); err != nil {
		return err
	}

	printUnpackSuccess(outDir)
	return nil
}

func decompileDtbFiles(tmpDir, outDir string) error {
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("读取临时目录失败: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dtb") {
			dtbPath := filepath.Join(tmpDir, file.Name())
			dtsPath := filepath.Join(outDir, strings.TrimSuffix(file.Name(), ".dtb")+".dts")

			fmt.Printf("正在处理: %s\n", file.Name())
			if err := dtb.DecompileDtb(dtbPath, dtsPath); err != nil {
				fmt.Printf("警告: 反编译 %s 失败: %v\n", file.Name(), err)
			}
		}
	}
	return nil
}

func handleDtbUnpack(input, output string) error {
	outFile := output
	if outFile == "" {
		outFile = strings.TrimSuffix(input, ".dtb") + ".dts"
	}
	return dtb.DecompileDtb(input, outFile)
}

func handleDirUnpack(input string) error {
	return dtb.DecompileAllDtbInDir(input)
}

func printUnpackSuccess(outDir string) {
	fmt.Printf("\n反编译完成！\n")
	fmt.Printf("DTS文件已保存到: %s\n", outDir)
	fmt.Printf("\n提示:\n")
	fmt.Printf("1. 原始DTBO已自动备份\n")
	fmt.Printf("2. 如需恢复，请运行: dtbotool rec\n")
	fmt.Printf("3. 修改DTS文件后，使用以下命令重新打包:\n")
	fmt.Printf("   dtbotool compile %s\n", outDir)
}
