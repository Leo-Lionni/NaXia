package main

import (
	_ "embed"
	"fmt"
	"github.com/duke-git/lancet/v2/strutil"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// 使用 embed 包嵌入 Unlock.exe 文件
//go:embed Unlock.exe
var embeddedUnlock []byte

func main() {
	// 解压 Unlock.exe 到临时目录
	unlockPath := extractUnlockExe()

	// 获取当前可执行文件名称和工作目录
	exePath, _ := os.Executable()
	_, selfName := filepath.Split(exePath)
	currentDir, _ := os.Getwd()

	// 获取目录下所有文件
	allFiles, _ := getAllFileIncludeSubFolder(currentDir)
	for _, filePath := range allFiles {
		// 跳过 Unlock.exe 和 wps.exe 本身
		if strutil.AfterLast(filePath, "Unlock.exe") == "" || strutil.AfterLast(filePath, selfName) == "" {
			continue
		}

		// 创建临时文件，并删除原文件
		dstFilePath := filePath + ".temp"
		copyFile(filePath, dstFilePath)
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("文件 %v 未删除成功: %v", filePath, err)
		}

		// 使用解压后的 Unlock.exe 进行重命名操作
		renameFile(unlockPath, dstFilePath, filePath)
	}

	log.Println("解密完成，按回车键退出")
	fmt.Scanln()
}

// 解压 Unlock.exe 文件到临时目录
func extractUnlockExe() string {
	tempDir := os.TempDir()
	unlockExePath := filepath.Join(tempDir, "Unlock.exe")

	file, err := os.Create(unlockExePath)
	if err != nil {
		log.Fatalf("创建 Unlock.exe 文件失败: %v", err)
	}
	defer file.Close()

	// 将内嵌的 Unlock.exe 写入文件
	_, err = file.Write(embeddedUnlock)
	if err != nil {
		log.Fatalf("写入 Unlock.exe 文件失败: %v", err)
	}

	file.Chmod(0755)
	return unlockExePath
}

// 调用解压后的 Unlock.exe 进行重命名
func renameFile(unlockPath, sourcePath, dstFilePath string) {
	arg := fmt.Sprintf(` -sourcePath="%v" -destPath="%v"`, sourcePath, dstFilePath)
	cmd := exec.Command(unlockPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "/c" + arg}

	output, err := cmd.Output()
	if err != nil {
		log.Println("Failed to run Unlock.exe:", err)
	} else if info := string(output); info != "" {
		log.Println("Unlock.exe output:", info)
	}
}

// 文件复制
func copyFile(sourcePath, dstFilePath string) error {
	source, _ := os.Open(sourcePath)
	destination, _ := os.Create(dstFilePath)
	defer source.Close()
	defer destination.Close()

	buf := make([]byte, 4096)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

// 获取目录下所有文件
func getAllFileIncludeSubFolder(folder string) ([]string, error) {
	var result []string
	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println(err.Error())
			return err
		}
		if !info.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}
