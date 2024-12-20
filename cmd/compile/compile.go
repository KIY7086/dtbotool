package compile

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kiy7086/dtbotool/cmd/dtb"
	"github.com/kiy7086/dtbotool/cmd/dtbo"
)

// HandleCompile 处理编译操作
func HandleCompile(input, output string) error {
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("'%s' 不存在", input)
	}

	fmt.Printf("正在编译 %s ...\n", input)

	if info, err := os.Stat(input); err == nil && info.IsDir() {
		return handleDirCompile(input, output)
	}
	return handleFileCompile(input, output)
}

func handleDirCompile(input, output string) error {
	files, err := os.ReadDir(input)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	var dtsCount, dtbCount int
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".dts") {
			dtsCount++
		} else if strings.HasSuffix(file.Name(), ".dtb") {
			dtbCount++
		}
	}

	if dtsCount > 0 {
		return handleDtsCompile(input, output, dtsCount)
	} else if dtbCount > 0 {
		return handleDtbCompile(input, output)
	}
	return fmt.Errorf("目录中未找到 DTS 或 DTB 文件")
}

func handleDtsCompile(input, output string, dtsCount int) error {
	fmt.Printf("\n找到 %d 个 DTS 文件，请选择操作:\n", dtsCount)
	fmt.Println("1. 编译为DTB文件")
	fmt.Println("2. 编译并打包为DTBO镜像")
	fmt.Print("\n请选择 [1/2]: ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		if err := dtb.CompileAllDtsInDir(input); err != nil {
			return fmt.Errorf("编译失败: %v", err)
		}
		fmt.Printf("已编译 %d 个 DTB 文件\n", dtsCount)
		return nil

	case "2":
		return compileDtsToDtbo(input, output)

	default:
		return fmt.Errorf("操作已取消")
	}
}

func compileDtsToDtbo(input, output string) error {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "dtbo_compile_*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 编译到临时目录
	if err := dtb.CompileAllDtsInDir(input, tmpDir); err != nil {
		return fmt.Errorf("编译失败: %v", err)
	}

	// 生成输出文件名
	if output == "" {
		output = generateDtboFileName(input)
	}

	// 打包为 DTBO
	if err := dtbo.PackDtbo(tmpDir, output); err != nil {
		return fmt.Errorf("打包失败: %v", err)
	}

	if err := verifyDtboImage(output); err != nil {
		os.Remove(output)
		return fmt.Errorf("DTBO验证失败: %v", err)
	}

	fmt.Printf("已生成 DTBO 镜像: %s\n", output)
	return nil
}

func handleDtbCompile(input, output string) error {
	if output == "" {
		output = generateDtboFileName(input)
	}

	if err := dtbo.PackDtbo(input, output); err != nil {
		return fmt.Errorf("打包失败: %v", err)
	}

	if err := verifyDtboImage(output); err != nil {
		os.Remove(output)
		return fmt.Errorf("DTBO验证失败: %v", err)
	}

	fmt.Printf("已生成 DTBO 镜像: %s\n", output)
	return nil
}

func handleFileCompile(input, output string) error {
	if !strings.HasSuffix(input, ".dts") {
		return fmt.Errorf("不支持的文件类型，请使用 .dts 文件")
	}

	if output == "" {
		output = strings.TrimSuffix(input, ".dts") + ".dtb"
	}

	if err := dtb.CompileDts(input, output); err != nil {
		return fmt.Errorf("编译失败: %v", err)
	}

	return nil
}

func generateDtboFileName(input string) string {
	timestamp := time.Now().Format("20060102_150405")
	dirContent, err := os.ReadDir(input)
	if err != nil {
		return fmt.Sprintf("dtbo_%s.img", timestamp)
	}

	var contentStr string
	for _, entry := range dirContent {
		contentStr += entry.Name()
	}
	hash := md5.Sum([]byte(contentStr))
	hashStr := hex.EncodeToString(hash[:])[:8]
	return fmt.Sprintf("dtbo_%s_%s.img", timestamp, hashStr)
}

func verifyDtboImage(dtboFile string) error {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "dtbo_verify_*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 验证文件大小
	info, err := os.Stat(dtboFile)
	if err != nil {
		return fmt.Errorf("无法读取DTBO文件: %v", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("DTBO文件为空")
	}
	if info.Size() > 20*1024*1024 {
		return fmt.Errorf("DTBO文件过大 (%d bytes)", info.Size())
	}

	// 验证DTBO格式
	if err := dtbo.UnpackDtbo(dtboFile, tmpDir); err != nil {
		return fmt.Errorf("DTBO格式无效: %v", err)
	}

	// 验证解出的DTB文件
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("读取临时目录失败: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("DTBO文件中未包含DTB")
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dtb") {
			dtbPath := filepath.Join(tmpDir, file.Name())
			if err := dtb.VerifyDtb(dtbPath); err != nil {
				return fmt.Errorf("DTB文件 %s 验证失败: %v", file.Name(), err)
			}
		}
	}

	return nil
}
