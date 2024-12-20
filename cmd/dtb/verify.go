package dtb

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// VerifyDtb 验证DTB文件
func VerifyDtb(dtbFile string) error {
	// 1. 检查文件大小
	info, err := os.Stat(dtbFile)
	if err != nil {
		return fmt.Errorf("无法读取DTB文件: %v", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("DTB文件为空")
	}

	// 2. 使用dtc反编译验证
	cmd := exec.Command("dtc",
		"-I", "dtb",
		"-O", "dts",
		"-f", // 强制处理
		"-q", // 安静模式
		dtbFile)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("DTB文件格式无效: %v\n%s", err, output)
	}

	// 3. 检查关键属性
	cmd = exec.Command("dtc",
		"-I", "dtb",
		"-O", "dts",
		"-f",
		dtbFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("无法读取DTB内容: %v", err)
	}

	// 检查必要的属性
	dts := string(output)
	requiredProps := []string{
		"compatible",
		"model",
	}

	var missingProps []string
	for _, prop := range requiredProps {
		if !strings.Contains(dts, prop) {
			missingProps = append(missingProps, prop)
		}
	}

	if len(missingProps) > 0 {
		fmt.Printf("警告: DTB文件缺少以下关键属性:\n")
		for _, prop := range missingProps {
			fmt.Printf("  - %s\n", prop)
		}
	}

	fmt.Printf("DTB文件验证通过 (大小: %d 字节)\n", info.Size())
	return nil
}
