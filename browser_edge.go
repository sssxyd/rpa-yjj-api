package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/playwright-community/playwright-go"
)

// isPortAvailable 检查指定端口是否可用
func isPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// allowPortThroughFirewall 添加防火墙规则以允许指定端口
func allowPortThroughFirewall(port string) error {
	// netsh 命令示例：允许 TCP 9222 端口的传入连接
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name=Allow Edge Debug Port "+port,
		"dir=in",
		"action=allow",
		"protocol=TCP",
		"localport="+port)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("无法添加防火墙规则: %v", err)
	}
	fmt.Println("已为端口", port, "添加防火墙规则")
	return nil
}

func GetValidPort() (string, error) {
	count := 0
	startPort := 9222
	endPort := 9322
	for i := startPort; i < endPort; i++ {
		// 随机选择一个端口
		port := i
		portStr := fmt.Sprintf("%d", port)
		// 检查端口是否可用
		if isPortAvailable(portStr) {
			return portStr, nil
		}
		count++
		fmt.Printf("端口 %s 被占用，尝试次数 %d/%d\n", portStr, i+1, count)
	}
	return "", fmt.Errorf("在 %d 次尝试后未找到可用端口", count)
}

// StartWithPlaywright 启动 Edge 并返回浏览器和页面实例
func StartWithPlaywright() (browser playwright.Browser, page playwright.Page, cleanup func()) {
	// 1. 找到系统中 Edge 的可执行文件路径
	edgePath := "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"

	// 2. 获取一个可用端口
	debugPort, err := GetValidPort()
	if err != nil {
		log.Fatalf("无法获取可用端口: %v", err)
	}

	// 3. 添加防火墙规则以允许该端口
	// err = allowPortThroughFirewall(debugPort)
	// if err != nil {
	// 	log.Fatalf("防火墙规则添加失败: %v", err)
	// }

	// 4. 使用系统命令启动 Edge，带调试端口
	//--remote-debugging-port={port} --remote-allow-origins=http://127.0.0.1:{port}
	fmt.Printf("exec command %s --new-window about:blank --remote-debugging-port=%s\n", edgePath, debugPort)
	cmd := exec.Command(edgePath, "--new-window", "about:blank", "--remote-debugging-port="+debugPort, "--remote-allow-origins=http://127.0.0.1:"+debugPort)
	err = cmd.Start()
	if err != nil {
		log.Fatalf("无法启动 Edge 浏览器: %v", err)
	}
	fmt.Println("Edge 浏览器已通过系统命令启动，调试端口:", debugPort)

	// 5. 等待片刻，确保浏览器启动
	time.Sleep(2 * time.Second)

	// 6. 初始化 Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("无法启动 Playwright: %v", err)
	}

	// 7. 连接到已启动的 Edge 实例
	browser, err = pw.Chromium.ConnectOverCDP("http://127.0.0.1:" + debugPort)
	if err != nil {
		log.Fatalf("无法连接到 Edge 实例: %v", err)
	}

	// 8. 获取浏览器上下文并关闭所有默认页面
	contexts := browser.Contexts()
	if len(contexts) == 0 {
		log.Fatalf("未找到任何浏览器上下文")
	}
	browserContext := contexts[0]
	pages := browserContext.Pages()
	for _, p := range pages {
		err = p.Close()
		if err != nil {
			log.Printf("关闭页面失败: %v", err)
		}
	}

	// 9. 创建一个新的空白页面
	page, err = browserContext.NewPage()
	if err != nil {
		log.Fatalf("无法创建新页面: %v", err)
	}
	_, err = page.Goto("about:blank", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		log.Fatalf("无法导航到空白页: %v", err)
	}

	// 10. 设置 User-Agent（可选）
	err = page.SetExtraHTTPHeaders(map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0",
	})
	if err != nil {
		log.Fatalf("无法设置 HTTP 头: %v", err)
	}

	// 11. 返回清理函数
	cleanup = func() {
		browser.Close()
		pw.Stop()
	}

	return browser, page, cleanup
}
