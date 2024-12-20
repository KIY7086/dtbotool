package dtb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CompileAllDtsInDir 编译目录中的所有DTS文件
func CompileAllDtsInDir(dtsDir string, outDir ...string) error {
	// 确定输出目录
	dtbDir := strings.TrimSuffix(dtsDir, "_decompiled") + "_compiled"
	if len(outDir) > 0 && outDir[0] != "" {
		dtbDir = outDir[0]
	}

	if err := os.MkdirAll(dtbDir, 0755); err != nil {
		return fmt.Errorf("创建DTB输出目录失败: %v", err)
	}

	files, err := os.ReadDir(dtsDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	// 统计文件数量
	var dtsFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dts") {
			dtsFiles = append(dtsFiles, file.Name())
		}
	}

	if len(dtsFiles) == 0 {
		return fmt.Errorf("在 %s 目录中未找到DTS文件", dtsDir)
	}

	fmt.Printf("找到 %d 个 DTS 文件需要编译\n", len(dtsFiles))

	// 编译所有文件
	var failedFiles []string
	for _, fileName := range dtsFiles {
		dtsPath := filepath.Join(dtsDir, fileName)
		dtbPath := filepath.Join(dtbDir, strings.TrimSuffix(fileName, ".dts")+".dtb")

		fmt.Printf("\n正在处理: %s\n", fileName)

		if err := CompileDts(dtsPath, dtbPath); err != nil {
			fmt.Printf("警告: 编译失败: %v\n", err)
			failedFiles = append(failedFiles, fileName)
			continue
		}
	}

	// 显示编译结果摘要
	if len(failedFiles) > 0 {
		fmt.Printf("\n编译完成，但有 %d 个文件失败:\n", len(failedFiles))
		for _, file := range failedFiles {
			fmt.Printf("- %s\n", file)
		}
		return fmt.Errorf("部分文件编译失败")
	}

	fmt.Printf("\n所有文件编译成功，输出目录: %s\n", dtbDir)
	return nil
}

// CompileDts 将DTS文件编译为DTB文件
func CompileDts(dtsFile, dtbFile string) error {
	fmt.Printf("正在编译 %s...\n", dtsFile)

	// 首先检查输入文件是否存在
	if _, err := os.Stat(dtsFile); os.IsNotExist(err) {
		return fmt.Errorf("DTS文件不存在: %s", dtsFile)
	}

	// 使用dtc编译
	cmd := exec.Command("dtc",
		"-I", "dts", // 输入格式为DTS
		"-O", "dtb", // 输出格��为DTB
		"-o", dtbFile, // 输出文件
		dtsFile) // 输入文件

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("编译DTS失败: %v\n%s", err, output)
	}

	fmt.Printf("已将 %s 编译为 %s\n", dtsFile, dtbFile)
	return nil
}
