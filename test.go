package main

import (
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

func Baidu() {
	// 1. 初始化 Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("无法启动 Playwright: %v", err)
	}
	defer pw.Stop()

	// 2. 启动本地安装的 Microsoft Edge 浏览器
	// 使用 ExecutablePath 指定 Edge 的可执行文件路径（Windows 下通常是默认路径）
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Channel:  playwright.String("msedge"),   // 指定使用本地 Edge
		Headless: playwright.Bool(false),        // 显示浏览器窗口
		SlowMo:   playwright.Float(100),         // 慢动作模式，模拟人类操作（可选）
		Args:     []string{"--start-maximized"}, // 启动时最大化窗口
	})
	if err != nil {
		log.Fatalf("无法启动 Edge 浏览器: %v", err)
	}
	defer browser.Close()

	// 3. 创建新页面
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("无法创建新页面: %v", err)
	}

	// 4. 设置额外的 HTTP 头，消除程序化操作特征
	err = page.SetExtraHTTPHeaders(map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0", // 模拟 Edge 的 UA
	})
	if err != nil {
		log.Fatalf("无法设置 HTTP 头: %v", err)
	}

	// 5. 访问百度网站
	_, err = page.Goto("https://www.baidu.com", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		log.Fatalf("无法访问百度: %v", err)
	}
	fmt.Println("已成功访问百度网站")

	// 6. 可选：模拟人类行为（如滚动页面）
	_, err = page.Evaluate("window.scrollTo(0, document.body.scrollHeight)")
	if err != nil {
		log.Fatalf("滚动页面失败: %v", err)
	}

	// 7. 等待一段时间以便观察
	time.Sleep(15 * time.Second)

	// 8. 关闭页面
	page.Close()

	// 9. 关闭浏览器
	browser.Close()
}
