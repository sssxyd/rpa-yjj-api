package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/sssxyd/go-lts-core"
)

func init() {
	root_dir := get_app_root_dir()
	log_path := filepath.Join(root_dir, "logs", "app.log")
	storage_path := filepath.Join(root_dir, "data", "storage.db")

	lts.Initialize(&lts.Options{
		LogConfig:     lts.LogConfig{MaxAgeDay: 0, StdOut: true, FilePath: log_path},
		StorageConfig: lts.StorageConfig{FilePath: storage_path},
		DBConfigs:     []lts.DBConfig{},
	})
}

func dispose() {
	lts.Dispose()
}

func handleShutdown() {
	// 创建一个 channel 来接收操作系统信号
	signalChan := make(chan os.Signal, 1)

	// 捕获 SIGINT (Ctrl+C) 和 SIGTERM (systemctl stop) 信号
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-signalChan
	log.Printf("Received signal: %s. Shutting down...", sig)

	dispose()

	os.Exit(0)
}

func main() {
	go handleShutdown()

	defer dispose()

	root_path := get_app_root_dir()
	fmt.Println("Root path:", root_path)

	_, page, cleanup := StartWithPlaywright()

	defer cleanup()

	// 使用页面访问百度（可选）
	_, err := page.Goto("https://www.baidu.com", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		log.Fatalf("无法访问百度: %v", err)
	}
	fmt.Println("已成功访问百度网站")
	time.Sleep(15 * time.Second)

	fmt.Println("--------------------")

}
