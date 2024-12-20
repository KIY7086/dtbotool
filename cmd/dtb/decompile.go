package dtb

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DecompileAllDtbInDir 反编译目录中的所有DTB文件
func DecompileAllDtbInDir(dtbDir string) error {
	// 创建DTS输出目录
	dtsDir := strings.TrimSuffix(dtbDir, "_extracted") + "_decompiled"
	if err := os.MkdirAll(dtsDir, 0755); err != nil {
		return fmt.Errorf("创建DTS输出目录失败: %v", err)
	}

	files, err := os.ReadDir(dtbDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dtb") {
			dtbPath := filepath.Join(dtbDir, file.Name())
			dtsPath := filepath.Join(dtsDir, strings.TrimSuffix(file.Name(), ".dtb")+".dts")
			if err := DecompileDtb(dtbPath, dtsPath); err != nil {
				fmt.Printf("警告: 反编译 %s 失败: %v\n", dtbPath, err)
			}
		}
	}

	fmt.Printf("已将反编译的DTS文件保存到: %s\n", dtsDir)
	return nil
}

// DecompileDtb 将DTB文件反编译为DTS文件
func DecompileDtb(dtbFile, dtsFile string) error {
	// 使用dtc反编译
	cmd := exec.Command("dtc",
		"-I", "dtb", // 输入格式为DTB
		"-O", "dts", // 输出格式为DTS
		"-o", dtsFile, // 输出文件
		"-S", "4", // 缩进级别
		"-R", "8", // 每行最大引用数
		"-b", "0", // 设置引导CPU为0
		"-@",    // 启用符号引用
		dtbFile) // 输入文件

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("反编译DTB失败: %v\n%s", err, output)
	}

	fmt.Printf("已将 %s 反编译为 %s\n", dtbFile, dtsFile)
	return nil
}
