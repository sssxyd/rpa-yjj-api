package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// 获取应用程序名称，并去除后缀（如 .exe）
func get_app_name() string {
	execPath, err := os.Executable() // 获取程序绝对路径
	if err != nil {
		fmt.Println("Error:", err)
		return "unknown"
	}

	appName := filepath.Base(execPath) // 提取文件名（包含后缀）

	// 去除后缀（跨平台处理）
	if runtime.GOOS == "windows" && strings.HasSuffix(appName, ".exe") {
		appName = strings.TrimSuffix(appName, ".exe")
	} else {
		ext := filepath.Ext(appName) // 提取扩展名
		if ext != "" {
			appName = strings.TrimSuffix(appName, ext) // 去除扩展名
		}
	}

	return appName
}

func get_app_root_dir() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return ""
	}
	exeDir := filepath.Dir(exePath)

	// 判断是否在临时目录中运行（典型的 go run 行为）
	if strings.Contains(exePath, os.TempDir()) {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			fmt.Println("Failed to get caller information")
			return ""
		}
		return filepath.Dir(filename)
	} else {
		// 默认返回可执行文件所在目录
		return exeDir
	}
}

// 获取 AppData 目录（如果目录不存在，会自动创建）
func get_app_data_dir() (string, error) {
	var appDataDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: 获取 %AppData% 或 %LocalAppData%
		appDataDir = os.Getenv("AppData") // 常见路径: C:\Users\<User>\AppData\Roaming
		if appDataDir == "" {
			appDataDir = os.Getenv("LocalAppData") // 常见路径: C:\Users\<User>\AppData\Local
		}

	case "darwin":
		// macOS: ~/Library/Application Support
		home := os.Getenv("HOME")
		if home != "" {
			appDataDir = filepath.Join(home, "Library", "Application Support")
		}

	case "linux":
		// Linux: ~/.local/share
		home := os.Getenv("HOME")
		if home != "" {
			appDataDir = filepath.Join(home, ".local", "share")
		}
	}

	// 如果目录路径为空，返回错误
	if appDataDir == "" {
		return "", fmt.Errorf("failed to determine AppData directory for OS: %s", runtime.GOOS)
	}

	appDataDir = filepath.Join(appDataDir, get_app_name())
	// 检查目录是否存在，如果不存在则创建
	if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(appDataDir, os.ModePerm) // 递归创建目录
		if err != nil {
			return "", fmt.Errorf("failed to create AppData directory: %v", err)
		}
	}

	return appDataDir, nil
}
