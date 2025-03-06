package browser

import (
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

type EdgeTabPage struct {
	id      string
	url     string
	browser *EdgeBrowser
	page    playwright.Page
	session playwright.CDPSession
	active  bool
}

func newEdgeTabPage(id string, url string, browser *EdgeBrowser, page playwright.Page) *EdgeTabPage {

	tabPage := &EdgeTabPage{
		id:      id,
		url:     url,
		browser: browser,
		page:    page,
	}

	return tabPage
}

func (t *EdgeTabPage) ID() string {
	return t.id
}

func (t *EdgeTabPage) Title() string {
	title, err := t.page.Title()
	if err != nil {
		log.Printf("Failed to get page title: %v", err)
		return ""
	}
	return title
}

func (t *EdgeTabPage) URL() string {
	t.page.BringToFront()
	return t.page.URL()
}

func (t *EdgeTabPage) Domain() string {
	url := t.URL()
	domain, err := extractDomainFromUrl(url)
	if err != nil {
		log.Printf("Failed to extract domain from URL: %s", url)
		return ""
	}
	return domain
}

func (t *EdgeTabPage) IsClosed() bool {
	return t.page.IsClosed()
}

func (t *EdgeTabPage) BringToFront() {
	t.page.BringToFront()
}

func (t *EdgeTabPage) Page() playwright.Page {
	return t.page
}

func (t *EdgeTabPage) OpenInNewTab(id string, action func() error, timeout time.Duration) TabPage {
	// 互斥锁，防止同时打开多个标签页
	t.browser.locker.Lock()
	defer t.browser.locker.Unlock()

	// 创建一个通道来接收新页面
	newPageChan := make(chan playwright.Page, 1)

	// 在 goroutine 中监听新页面
	go func() {
		newPage, err := t.browser.context.WaitForEvent("page")
		if err != nil {
			log.Printf("等待新页面失败: %v", err)
			return
		}
		newPageObj := newPage.(playwright.Page)
		err = newPageObj.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded,
		})
		if err != nil {
			log.Printf("新标签页加载失败: %v", err)
			newPageObj.Close()
			return
		}

		listen_page_console_log(newPageObj)
		block_debug_port_detector(newPageObj, t.browser.port)

		newPageChan <- newPageObj
	}()

	// 触发事件
	err := action()
	if err != nil {
		log.Printf("触发事件失败: %v", err)
		return nil
	}

	// 等待新页面
	select {
	case newPage := <-newPageChan:
		tabPage := t.browser.addTabPage(id, newPage.URL(), newPage)
		log.Printf("成功捕获新标签页, ID: %s", id)
		return tabPage
	case <-time.After(timeout):
		log.Printf("等待新标签页超时: %v", timeout)
		return nil
	}
}

func (t *EdgeTabPage) WaitSelector(selector string, timeout float64) playwright.Locator {
	locator := t.page.Locator(selector)
	if locator == nil {
		log.Printf("无法找到选择器: %s", selector)
		return nil
	}
	count, err := locator.Count()
	if err != nil {
		log.Printf("无法获取选择器数量: %v", err)
		return nil
	}
	if count == 0 {
		log.Printf("选择器未匹配到任何元素: %s", selector)
		return nil
	}
	if count > 1 {
		log.Printf("选择器匹配到多个元素: %s, 取第一条返回", selector)
		locator = locator.First()
	}
	err = locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(timeout),
	})
	if err != nil {
		return nil
	}
	return locator
}

func (t *EdgeTabPage) QuerySelector(selector string) playwright.Locator {
	locator := t.page.Locator(selector)
	if locator == nil {
		log.Printf("无法找到选择器: %s", selector)
		return nil
	}
	count, err := locator.Count()
	if err != nil {
		log.Printf("无法获取选择器数量: %v", err)
		return nil
	}
	if count == 0 {
		log.Printf("选择器未匹配到任何元素: %s", selector)
		return nil
	}
	if count > 1 {
		log.Printf("选择器匹配到多个元素: %s, 取第一条返回", selector)
		locator = locator.First()
	}
	return locator
}

func (t *EdgeTabPage) QuerySelectorAll(selector string) []playwright.Locator {
	locator := t.page.Locator(selector)
	if locator == nil {
		log.Printf("无法找到选择器: %s", selector)
		return []playwright.Locator{}
	}
	count, err := locator.Count()
	if err != nil {
		log.Printf("无法获取选择器数量: %v", err)
		return []playwright.Locator{}
	}
	if count == 0 {
		log.Printf("选择器未匹配到任何元素: %s", selector)
		return []playwright.Locator{}
	}
	if count == 1 {
		return []playwright.Locator{locator}
	}
	items, err := locator.All()
	if err != nil {
		log.Printf("无法获取所有选择器: %v", err)
		return []playwright.Locator{}
	}
	return items
}

func (t *EdgeTabPage) ClearLocalData() error {
	// ---------- 清除所有存储 ----------
	log.Printf("正在清除站点%s的所有本地存储...", t.Domain())
	// 1. 清除 localStorage 和 sessionStorage
	if _, err := t.page.Evaluate("localStorage.clear()"); err != nil {
		log.Fatalf("清空 localStorage 失败: %v", err)
	}
	if _, err := t.page.Evaluate("sessionStorage.clear()"); err != nil {
		log.Fatalf("清空 sessionStorage 失败: %v", err)
	}

	// 2. 清除 Cookies（针对当前域名）
	if err := t.browser.context.ClearCookies(); err != nil {
		log.Fatalf("清除 Cookies 失败: %v", err)
	}

	// 3. 清除 IndexedDB（通过 JavaScript）
	if _, err := t.page.Evaluate(`
		async () => {
			const databases = await window.indexedDB.databases();
			for (const db of databases) {
				if (db.name) {
					window.indexedDB.deleteDatabase(db.name);
				}
			}
		}
	`); err != nil {
		log.Fatalf("清除 IndexedDB 失败: %v", err)
	}
	log.Println("所有存储已清除！")

	return nil
}

func (t *EdgeTabPage) Goto(url string) error {
	// 导航
	_, err := t.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return fmt.Errorf("无法访问网站: %v", err)
	}

	block_debug_port_detector(t.page, t.browser.port)

	// 等待页面完全加载
	err = t.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
	if err != nil {
		return fmt.Errorf("等待页面加载失败: %v", err)
	}
	log.Printf("已成功访问网站: %s", url)

	return nil
}

func (t *EdgeTabPage) Evaluate(expression string, arg ...any) (any, error) {
	return t.page.Evaluate(expression, arg...)
}
