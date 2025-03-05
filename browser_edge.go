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

func removeByValue(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // 未找到匹配值，返回原切片
}

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

func block_detector(p playwright.Page, port int) error {
	script := fmt.Sprintf(`() => {
		const debugPort = "%d";
		const originalGetEntries = performance.getEntries;
		const originalGetEntriesByType = performance.getEntriesByType;
		const OriginalWebSocket = window.WebSocket;

		// 过滤 Performance 条目
        performance.getEntries = function() {
            return originalGetEntries.call(this).filter(entry => 
                !entry.name.includes("127.0.0.1:" + debugPort) && 
                !entry.name.includes("localhost:" + debugPort)
            );
        };

		performance.getEntriesByType = function(type) {
			return originalGetEntriesByType.call(this, type).filter(entry =>
                !entry.name.includes("127.0.0.1:" + debugPort) && 
                !entry.name.includes("localhost:" + debugPort)
			);
		};

		// 拦截 WebSocket 连接
		window.WebSocket = function(urlArg, protocols) {
		// 统一处理 URL 格式
		const url = urlArg instanceof URL ? urlArg.href : urlArg;
		
		// 检测目标地址
		if (typeof url === 'string' && 
			(url.includes("127.0.0.1:" + debugPort) || 
			url.includes("localhost:" + debugPort))) {
			
			// 创建虚假的 WebSocket 对象
			const fakeWs = new OriginalWebSocket('ws://invalid-host-' + Date.now());
			
			// 立即关闭连接并修改状态
			Object.defineProperty(fakeWs, 'readyState', {
			value: OriginalWebSocket.CLOSED,
			writable: false
			});
			
			// 强制触发错误事件
			const errorEvent = new Event('error');
			errorEvent.initEvent('error', false, false);
			
			// 设置异步触发确保执行顺序
			setTimeout(() => {
			if (typeof fakeWs.onerror === 'function') {
				fakeWs.onerror(errorEvent);
			}
			fakeWs.dispatchEvent(errorEvent);
			
			fakeWs.close();
			}, 0);
			
			return fakeWs;
		}
		
		// 正常连接处理
		return protocols ? 
			new OriginalWebSocket(urlArg, protocols) : 
			new OriginalWebSocket(urlArg);
		};

		// 保持原型链完整
		window.WebSocket.prototype = OriginalWebSocket.prototype;		
	}`, port)
	_, err := p.Evaluate(script)
	if err != nil {
		log.Printf("注入拦截脚本失败: %v", err)
		return err
	}

	// // 网络请求拦截
	// p.Route("**/*:*", func(route playwright.Route) {
	// 	req := route.Request()
	// 	if strings.Contains(req.URL(), fmt.Sprintf(":%d", port)) {
	// 		route.Abort()
	// 	} else {
	// 		route.Continue()
	// 	}
	// })

	log.Printf(">>>已拦截 Performance API 和网络请求")

	return nil
}

type PlaywrightEdge struct {
	port     int
	pw       *playwright.Playwright
	browser  playwright.Browser
	context  playwright.BrowserContext
	allPages map[string]playwright.Page
	tabIds   []string
	index    int
}

func NewPlaywrightEdge(port int) (*PlaywrightEdge, error) {
	// 1. 找到系统中 Edge 的可执行文件路径
	edgePath := "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"

	// 2. 获取一个可用端口，用于调试
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
		port:     debugPort,
		pw:       pw,
		browser:  browser,
		context:  browserContext,
		allPages: make(map[string]playwright.Page),
		tabIds:   []string{},
		index:    0,
	}

	// 7. 创建一个默认页面
	_, err = pe.NewPage("default", "about:blank")
	if err != nil {
		pe.Close()
		return nil, err
	}
	pe.setCurrentPage("default")
	return pe, nil
}

func (pe *PlaywrightEdge) addPage(id string, page playwright.Page) {
	pe.allPages[id] = page
	pe.tabIds = append(pe.tabIds, id)
}

func (pe *PlaywrightEdge) removePage(id string) {
	p := pe.GetPage(id)
	if p == nil {
		return
	}
	err := p.Close()
	if err != nil {
		log.Printf("关闭页面失败: %v", err)
	}
	delete(pe.allPages, id)
	pe.tabIds = removeByValue(pe.tabIds, id)
	if len(pe.tabIds) == 0 {
		pe.NewPage("default", "about:blank")
		pe.index = 0
	} else {
		if pe.index >= len(pe.tabIds) {
			pe.index = len(pe.tabIds) - 1
		}
		pe.allPages[pe.tabIds[pe.index]].BringToFront()
	}
}

func (pe *PlaywrightEdge) setCurrentPage(id string) {
	for i, tabId := range pe.tabIds {
		if tabId == id {
			if pe.index == i {
				return
			}
			pe.index = i
			pe.allPages[id].BringToFront()
			log.Printf("切换到标签页: %s", pe.tabIds[pe.index])
			return
		}
	}
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

	// 监听控制台消息
	page.On("console", func(message playwright.ConsoleMessage) {
		log.Printf("console: %s", message.Text())
	})

	pe.addPage(id, page)

	err = pe.PageVisit(id, url)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (pe *PlaywrightEdge) GetPage(id string) playwright.Page {
	if page, ok := pe.allPages[id]; ok {
		return page
	}
	return nil
}

func (pe *PlaywrightEdge) CurrentPage() playwright.Page {
	return pe.allPages[pe.tabIds[pe.index]]
}

func (pe *PlaywrightEdge) ClosePage(id string) error {
	if id == "default" {
		return fmt.Errorf("默认页面不能关闭")
	}
	pe.removePage(id)
	return nil
}

func (pe *PlaywrightEdge) CloseCurrentPage() error {
	return pe.ClosePage(pe.tabIds[pe.index])
}

func (pe *PlaywrightEdge) PageVisit(id string, url string) error {
	// 导航
	page := pe.GetPage(id)
	if page == nil {
		return fmt.Errorf("未找到 ID 为 %s 的页面", id)
	}
	_, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return fmt.Errorf("无法访问网站: %v", err)
	}

	// // 拦截 Performance API
	// err = blockPerformanceAPIDetection(page)
	// if err != nil {
	// 	log.Printf("拦截 Performance API 失败: %v", err)
	// }
	block_detector(page, pe.port)

	// 等待页面完全加载
	err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
	if err != nil {
		return fmt.Errorf("等待页面加载失败: %v", err)
	}
	log.Printf("已成功访问网站: %s", url)

	return nil
}

func (pe *PlaywrightEdge) Visit(url string) error {
	return pe.PageVisit(pe.tabIds[pe.index], url)
}

func (pe *PlaywrightEdge) OpenNewPage(id string, action func() error, timeout float64) (playwright.Page, error) {
	// 创建一个通道来接收新页面
	newPageChan := make(chan playwright.Page, 1)

	// 在 goroutine 中监听新页面
	go func() {
		newPage, err := pe.context.WaitForEvent("page")
		if err != nil {
			log.Printf("等待新页面失败: %v", err)
			return
		}
		newPageObj := newPage.(playwright.Page)
		err = newPageObj.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded,
		})
		if err != nil {
			log.Printf("新页面加载失败: %v", err)
			return
		}
		block_detector(newPageObj, pe.port)
		newPageChan <- newPageObj
	}()

	// 触发事件
	err := action()
	if err != nil {
		return nil, fmt.Errorf("触发事件失败: %v", err)
	}

	// 等待新页面
	select {
	case newPage := <-newPageChan:
		pe.addPage(id, newPage)
		log.Printf("成功捕获新标签页，ID: %s", id)
		return newPage, nil
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return nil, fmt.Errorf("超时，未捕获到新标签页")
	}
}

func (pe *PlaywrightEdge) SwitchToPage(id string) error {
	if pe.tabIds[pe.index] == id {
		return nil
	}
	pe.setCurrentPage(id)
	return nil
}

func (pe *PlaywrightEdge) SwitchToNextPage() error {
	if len(pe.tabIds) <= 1 || pe.index == len(pe.tabIds)-1 {
		return nil
	}
	pe.index++
	pe.allPages[pe.tabIds[pe.index]].BringToFront()
	log.Printf("切换到标签页: %s", pe.tabIds[pe.index])
	return nil
}

func (pe *PlaywrightEdge) SwitchToPreviousPage() error {
	if len(pe.tabIds) <= 1 || pe.index == 0 {
		return nil
	}
	pe.index--
	pe.allPages[pe.tabIds[pe.index]].BringToFront()
	log.Printf("切换到标签页: %s", pe.tabIds[pe.index])
	return nil
}

func (pe *PlaywrightEdge) WaitForSelector(selector string, timeout float64) (playwright.Locator, error) {
	page := pe.CurrentPage()
	locator := page.Locator(selector)
	err := locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(timeout),
	})
	if err != nil {
		return nil, err
	}
	return locator, nil
}
