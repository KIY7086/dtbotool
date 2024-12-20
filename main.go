package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kiy7086/dtbotool/cmd/compile"
	"github.com/kiy7086/dtbotool/cmd/recovery"
	"github.com/kiy7086/dtbotool/cmd/unpack"
)

const VERSION = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		interactiveMode()
		return
	}

	unpackCmd := flag.NewFlagSet("unpack", flag.ExitOnError)
	rawOutput := unpackCmd.Bool("raw", false, "提取为原始dtb文件")
	output := unpackCmd.String("o", "", "指定输出文件/目录")

	compileCmd := flag.NewFlagSet("compile", flag.ExitOnError)
	compileOutput := compileCmd.String("o", "", "指定输出文件/目录")

	recCmd := flag.NewFlagSet("rec", flag.ExitOnError)
	purgeBackups := recCmd.Bool("purge", false, "删除所有备份文件")
	listBackups := recCmd.Bool("list", false, "列出所有备份文件")

	switch os.Args[1] {
	case "unpack":
		unpackCmd.Parse(os.Args[2:])
		if unpackCmd.NArg() == 0 {
			printUsage()
			return
		}
		if err := unpack.HandleUnpack(unpackCmd.Arg(0), *output, *rawOutput); err != nil {
			fmt.Printf("错误: %v\n", err)
		}

	case "compile":
		compileCmd.Parse(os.Args[2:])
		if compileCmd.NArg() == 0 {
			printUsage()
			return
		}
		if err := compile.HandleCompile(compileCmd.Arg(0), *compileOutput); err != nil {
			fmt.Printf("错误: %v\n", err)
		}

	case "-v", "--version":
		fmt.Printf("DTBO工具 v%s\n", VERSION)

	case "-h", "--help":
		printUsage()

	case "rec":
		recCmd.Parse(os.Args[2:])
		handleRecoveryCommand(*purgeBackups, *listBackups)

	default:
		fmt.Printf("未知命令: %s\n", os.Args[1])
		printUsage()
	}
}

func printUsage() {
	fmt.Printf(`DTBO工具 v%s
用法: 
    dtbotool                              # 进入交互式模式
    dtbotool unpack <输入文件> [输出文件/目录]
    dtbotool compile <输入文件/目录> [输出文件]
    dtbotool rec [选项]                    # 备份管理

示例:
    dtbotool                              # 启动交互式界面
    dtbotool unpack dtbo.img              # 将DTBO镜像转换为DTS文件
    dtbotool unpack dtbo.img --raw        # 将DTBO镜像提取为DTB文件
    dtbotool unpack device.dtb            # 将DTB转换为DTS
    dtbotool unpack dtb_dir/              # 批量转换目录中的DTB文件
    dtbotool compile device.dts           # 将DTS编译为DTB
    dtbotool compile dts_dir/             # 批量编译目录中的DTS文件
    dtbotool compile dtb_dir/ dtbo.img    # 将多个DTB打包为DTBO镜像
    dtbotool rec                          # 恢复最近的备份
    dtbotool rec --list                   # 列出所有备份
    dtbotool rec --purge                  # 清理所有备份

选项:
    --raw    提取为DTB文件而不是转换为DTS
    -o       指定输出文件/目录(可选)
    -v       显示版本信息
    -h       显示帮助信息
    --list   列出所有备份文件
    --purge  删除所有备份文件
`, VERSION)
}

func handleRecoveryCommand(purge, list bool) {
	var err error
	switch {
	case list:
		err = recovery.ListBackups()
	case purge:
		fmt.Println("警告: ��操作将删除所有备份文件!")
		fmt.Print("请输入 'DELETE' 确认删除: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "DELETE" {
			fmt.Println("操作已取消")
			return
		}
		fmt.Print("再次确认，这个操作不可恢复! 输入 'YES' 继续: ")
		fmt.Scanln(&confirm)
		if confirm != "YES" {
			fmt.Println("操作已取消")
			return
		}
		err = recovery.PurgeBackups()
		if err == nil {
			fmt.Println("所有备份文件已删除")
		}
	default:
		err = recovery.RestoreBackup()
	}
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}
}

func interactiveMode() {
	fmt.Println("欢迎使用DTBO工具!")
	for {
		fmt.Printf("\n请选择操作:\n")
		fmt.Println("1. 解包DTBO/DTB")
		fmt.Println("2. 编译DTS/打包DTB")
		fmt.Println("3. 恢复备份")
		fmt.Println("4. 查看备份")
		fmt.Println("5. 退出")
		fmt.Print("\n请输入选项 [1-5]: ")

		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			fmt.Print("\n请输入要解包的文件路径: ")
			var input string
			fmt.Scanln(&input)
			fmt.Print("是否提取为原始DTB文件? [y/N]: ")
			var raw string
			fmt.Scanln(&raw)
			if err := unpack.HandleUnpack(input, "", strings.ToLower(raw) == "y"); err != nil {
				fmt.Printf("错误: %v\n", err)
			}

		case "2":
			fmt.Print("\n请输入要编译的文件/目录路径: ")
			var input string
			fmt.Scanln(&input)
			if err := compile.HandleCompile(input, ""); err != nil {
				fmt.Printf("错误: %v\n", err)
			}

		case "3":
			handleRecoveryCommand(false, false)

		case "4":
			handleRecoveryCommand(false, true)

		case "5":
			fmt.Println("\n再见!")
			return

		default:
			fmt.Println("\n无效的选项，请重试")
		}
	}
}
