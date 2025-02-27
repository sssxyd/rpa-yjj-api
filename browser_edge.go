package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// isPortAvailable 检查指定端口是否可用
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func getValidPort() (int, error) {
	count := 0
	startPort := 9222
	endPort := 20000
	for i := startPort; i < endPort; i++ {
		// 随机选择一个端口
		port := i
		// 检查端口是否可用
		if isPortAvailable(port) {
			return port, nil
		}
		count++
		fmt.Printf("端口 %d 被占用，尝试次数 %d/%d\n", port, i+1, count)
	}
	return 0, fmt.Errorf("在 %d 次尝试后未找到可用端口", count)
}

// connectToExistingEdge 检查是否有已运行的 Edge 实例并尝试连接
func connectToExistingEdge(pw *playwright.Playwright, port string) (playwright.Browser, bool) {
	// 尝试连接到默认调试端口
	browser, err := pw.Chromium.ConnectOverCDP("http://127.0.0.1:" + port)
	if err == nil {
		return browser, true
	}
	return nil, false
}

// killEdgeProcesses 关闭所有 Edge 进程
func killEdgeProcesses() error {
	// 检查进程是否存在
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq msedge.exe")
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "msedge.exe") {
		fmt.Println("未找到 msedge.exe 进程")
		return nil
	}

	// 终止进程
	cmd = exec.Command("taskkill", "/F", "/IM", "msedge.exe")
	output, err = cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			log.Printf("命令退出码: %d, 输出: %s", exitCode, string(output))
			switch exitCode {
			case 1:
				return fmt.Errorf("权限不足，请以管理员身份运行")
			case 128:
				return fmt.Errorf("未找到 msedge.exe 进程")
			default:
				return fmt.Errorf("未知错误，退出码: %d, 输出: %s", exitCode, string(output))
			}
		}
		return fmt.Errorf("终止 msedge 失败: %v, 输出: %s", err, string(output))
	}
	fmt.Println("成功终止 msedge 进程")
	return nil
}

// startNewEdge 启动新的 Edge 实例
func startNewEdge(edgePath, port string) (*exec.Cmd, error) {
	cmd := exec.Command(edgePath,
		"--new-window",
		"about:blank",
		"--remote-debugging-port="+port,
		"--remote-allow-origins=http://127.0.0.1:"+port)
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("无法启动 Edge 浏览器: %v", err)
	}
	return cmd, nil
}

func blockPerformanceAPIDetection(p playwright.Page) error {
	_, err := p.Evaluate(`() => {
        // 保存原始方法
        const originalGetEntriesByType = performance.getEntriesByType;
        const originalGetEntries = performance.getEntries;

        // 重写 getEntriesByType 方法
        performance.getEntriesByType = function(type) {
            const entries = originalGetEntriesByType.call(this, type);
            // 过滤掉包含 127.0.0.1 的条目
            return entries.filter(entry => {
                return !entry.name.includes("127.0.0.1") && !entry.name.includes("localhost");
            });
        };

        // 重写 getEntries 方法
        performance.getEntries = function() {
            const entries = originalGetEntries.call(this);
            // 过滤掉包含 127.0.0.1 的条目
            return entries.filter(entry => {
                return !entry.name.includes("127.0.0.1") && !entry.name.includes("localhost");
            });
        };

        console.log(">>> 已拦截 Performance API，隐藏 127.0.0.1 连接");
    }`)
	if err != nil {
		log.Printf("注入 Performance API 拦截脚本失败: %v", err)
		return err
	}
	log.Printf("已在 %s 上拦截 Performance API", p.URL())
	return nil
}

type PlaywrightEdge struct {
	port        int
	pw          *playwright.Playwright
	browser     playwright.Browser
	context     playwright.BrowserContext
	allPages    map[string]playwright.Page
	currentPage playwright.Page
}

func NewPlaywrightEdge(port int) (*PlaywrightEdge, error) {
	// 1. 找到系统中 Edge 的可执行文件路径
	edgePath := "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"

	// 2. 获取一个可用端口
	var debugPort int
	if port > 0 {
		debugPort = port
	} else {
		var err error
		debugPort, err = getValidPort()
		if err != nil {
			return nil, fmt.Errorf("无法获取可用端口: %v", err)
		}
	}

	// 3. 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("无法启动 Playwright: %v", err)
	}

	// 4. 尝试连接到已运行的 Edge 实例
	browser, found := connectToExistingEdge(pw, fmt.Sprintf("%d", debugPort))
	if !found {
		// 如果没有找到已运行的 Edge 实例，则关闭所有 Edge 进程并启动一个新实例
		err = killEdgeProcesses()
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("无法关闭 Edge 进程: %v", err)
		}

		// 启动新的 Edge 实例
		_, err = startNewEdge(edgePath, fmt.Sprintf("%d", debugPort))
		if err != nil {
			pw.Stop()
			return nil, fmt.Errorf("无法启动 Edge 浏览器: %v", err)
		}

		// 等待片刻，确保浏览器启动
		time.Sleep(2 * time.Second)
		log.Printf("Edge 浏览器已通过系统命令启动，调试端口: %d\n", debugPort)

		// 连接到新启动的 Edge 实例
		browser, found = connectToExistingEdge(pw, fmt.Sprintf("%d", debugPort))
		if !found {
			pw.Stop()
			return nil, fmt.Errorf("无法连接到新启动的 Edge 实例")
		}
	} else {
		log.Printf("已找到正在运行的 Edge 实例，调试端口:%d\n", debugPort)
	}

	// 5. 获取浏览器上下文并关闭所有默认页面
	contexts := browser.Contexts()
	if len(contexts) == 0 {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("未找到任何浏览器上下文")
	}
	browserContext := contexts[0]
	pages := browserContext.Pages()
	for _, p := range pages {
		err = p.Close()
		if err != nil {
			log.Printf("关闭页面失败: %v", err)
		}
	}

	// 6. 创建 PlaywrightEdge 实例
	pe := &PlaywrightEdge{
		port:        debugPort,
		pw:          pw,
		browser:     browser,
		context:     browserContext,
		allPages:    make(map[string]playwright.Page),
		currentPage: nil,
	}

	// 7. 创建一个默认页面
	pe.currentPage, err = pe.NewPage("default", "about:blank")
	if err != nil {
		pe.Close()
		return nil, err
	}

	return pe, nil
}

func (pe *PlaywrightEdge) Close() {
	log.Println("正在关闭 Edge 浏览器...")
	for _, p := range pe.allPages {
		err := p.Close()
		if err != nil {
			log.Printf("关闭页面失败: %v", err)
		}
	}
	log.Println("已关闭所有页面")
	pe.currentPage = nil
	pe.context.Close()
	log.Println("已关闭浏览器上下文")
	pe.browser.Close()
	log.Println("已关闭浏览器")
	pe.pw.Stop()
	log.Println("已关闭 Playwright")
}

func (pe *PlaywrightEdge) NewPage(id string, url string) (playwright.Page, error) {
	if url == "" {
		url = "about:blank"
	}
	// 创建一个新的空白页面
	page, err := pe.context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("无法创建新页面: %v", err)
	}

	page.On("console", func(message playwright.ConsoleMessage) {
		log.Printf("console: %s", message.Text())
	})

	pe.currentPage = page
	err = pe.Goto(url)
	if err != nil {
		return nil, err
	}
	pe.allPages[id] = page

	return page, nil
}

func (pe *PlaywrightEdge) GetPage(id string) playwright.Page {
	if page, ok := pe.allPages[id]; ok {
		return page
	}
	return nil
}

func (pe *PlaywrightEdge) CurrentPage() playwright.Page {
	return pe.currentPage
}

func (pe *PlaywrightEdge) ClosePage(id string) error {
	if page, ok := pe.allPages[id]; ok {
		err := page.Close()
		if err != nil {
			return fmt.Errorf("关闭页面失败: %v", err)
		}
		delete(pe.allPages, id)
	}
	return nil
}

func (pe *PlaywrightEdge) Goto(url string) error {
	// 导航
	_, err := pe.currentPage.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return fmt.Errorf("无法访问网站: %v", err)
	}

	// 拦截 Performance API
	err = blockPerformanceAPIDetection(pe.currentPage)
	if err != nil {
		log.Printf("拦截 Performance API 失败: %v", err)
	}

	// 等待页面完全加载
	err = pe.currentPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
	if err != nil {
		return fmt.Errorf("等待页面加载失败: %v", err)
	}
	log.Printf("已成功访问网站: %s", url)

	return nil
}
