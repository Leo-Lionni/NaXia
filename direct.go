package main

import (
	"fmt"
	"github.com/duke-git/lancet/v2/strutil"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 获取当前执行文件路径和名称
	path, _ := os.Executable()
	_, selfName := filepath.Split(path)
	currentDir, _ := os.Getwd()

	// 获取所有文件（包含子目录）
	allFiles, _ := getAllFileIncludeSubFolder(currentDir)
	for _, filePath := range allFiles {
		// 跳过自身和 Unlock.exe
		if strutil.AfterLast(filePath, "Unlock.exe") == "" {
			continue
		}
		if strutil.AfterLast(filePath, selfName) == "" {
			continue
		}

		// 为每个文件创建临时文件路径
		tempFilePath := filePath + ".temp"
		
		// 复制文件到临时文件
		err := copyFile(filePath, tempFilePath)
		if err != nil {
			log.Printf("复制文件 %v 失败: %v", filePath, err)
			continue
		}

		// 删除原文件
		err = os.Remove(filePath)
		if err != nil {
			log.Printf("删除文件 %v 失败: %v", filePath, err)
			continue
		}

		// 直接重命名临时文件为原文件名
		err = os.Rename(tempFilePath, filePath)
		if err != nil {
			log.Printf("重命名文件 %v 失败: %v", tempFilePath, err)
		} else {
			log.Printf("文件 %v 解锁成功", filePath)
		}
	}

	log.Println("解密完成，按回车键退出")
	fmt.Scanln()
}

// 复制文件
func copyFile(sourcePath, dstFilePath string) (err error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, 1024)
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

// 获取目录下所有文件（包含子目录）
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
