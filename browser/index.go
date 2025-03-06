package browser

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

type TabPage interface {
	ID() string
	Title() string
	URL() string
	Domain() string
	IsClosed() bool
	BringToFront()
	OpenInNewTab(id string, action func() error, timeout time.Duration) TabPage
	WaitSelector(selector string, timeout float64) playwright.Locator
	QuerySelector(selector string) playwright.Locator
	QuerySelectorAll(selector string) []playwright.Locator
	ClearLocalData() error
	Goto(url string) error
	Evaluate(expression string, arg ...any) (any, error)
	Page() playwright.Page
}

type Browser interface {
	TabPages() []TabPage
	NewTabPage(id string, url string) TabPage
	FindTabPage(id string) TabPage
	SwitchToTabPage(id string) error
	Close() error
}

func ParseJson[T any](jsonstr string) (T, error) {
	var obj T
	err := json.Unmarshal([]byte(jsonstr), &obj)
	if err != nil {
		return obj, fmt.Errorf("json unmarshal error: %w", err)
	}
	return obj, nil
}

func Stringify[T any](obj T) (string, error) {
	// JSON序列化
	jsonstr, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %w", err)
	}
	return string(jsonstr), nil
}

func StartEdgeBrowser(edgePath string, startPort, endPort int) (Browser, error) {
	browser, err := newEdgeBrowser(edgePath, startPort, endPort)
	if err != nil {
		return nil, fmt.Errorf("failed to start Edge browser: %w", err)
	}
	return browser, nil
}
