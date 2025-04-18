package browser

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os/exec"
	"strings"

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

func extractDomainFromUrl(urlStr string) (string, error) {
	if !strings.Contains(urlStr, "://") { // 补全缺失的协议头
		urlStr = "http://" + urlStr
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	hostname := u.Hostname()
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid hostname: %s", hostname)
	}
	return parts[len(parts)-2] + "." + parts[len(parts)-1], nil
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func getValidPort(start, end int) (int, error) {
	count := 0
	for i := start; i <= end; i++ {
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

func connectToExistingEdge(pw *playwright.Playwright, port string) (playwright.Browser, bool) {
	// 尝试连接到默认调试端口
	browser, err := pw.Chromium.ConnectOverCDP("http://127.0.0.1:" + port)
	if err == nil {
		return browser, true
	}
	return nil, false
}

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
	log.Printf("Edge 浏览器已启动，调试端口: %s", port)
	return cmd, nil
}

func block_debug_port_detector(p playwright.Page, port int) error {
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

		performance.getEntriesByType = function(type){
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

	log.Printf(">>>已拦截Performance API和 WebSocket 请求")

	return nil
}

func listen_page_console_log(page playwright.Page) {
	page.On("console", func(message playwright.ConsoleMessage) {
		log.Printf(">>> %s", message.Text())
	})
}
