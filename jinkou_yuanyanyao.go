package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type MedicineInfo struct {
	OrderNo       string // 序号
	AuthCode      string // 批准文号或注册证号
	MedicineName  string // 药品名称
	TorchType     string // 剂型
	Specification string // 规格
	CanBi         string // 参比制剂
	CertDate      string // 批准日期
	Manufacturer  string // 生产厂商
	CertHolder    string // 上市许可证持有人
}

func NewMedicineInfo(lines []string) *MedicineInfo {
	if len(lines) < 9 {
		for i := len(lines); i < 9; i++ {
			lines = append(lines, "")
		}
	}
	return &MedicineInfo{
		OrderNo:       lines[0],
		AuthCode:      lines[1],
		MedicineName:  lines[2],
		TorchType:     lines[3],
		Specification: lines[4],
		CanBi:         lines[5],
		CertDate:      lines[6],
		Manufacturer:  lines[7],
		CertHolder:    lines[8],
	}
}

func GetMedicineInfoHeaders() []string {
	return []string{
		"序号",
		"批准文号或注册证号",
		"药品名称",
		"剂型",
		"规格",
		"参比制剂",
		"批准日期",
		"生产厂商",
		"上市许可证持有人",
	}
}

func (medicine *MedicineInfo) ToRowData() []string {
	return []string{
		medicine.OrderNo,
		medicine.AuthCode,
		medicine.MedicineName,
		medicine.TorchType,
		medicine.Specification,
		medicine.CanBi,
		medicine.CertDate,
		medicine.Manufacturer,
		medicine.CertHolder,
	}
}

func get_page_data(edge *PlaywrightEdge) ([]*MedicineInfo, error) {

	locator := edge.CurrentPage().Locator(".layui-table-body.layui-table-main tr")

	// 获取所有匹配的 tr 元素
	trLocators, err := locator.All()
	if err != nil {
		log.Fatalf("无法获取所有 <tr> 元素: %v", err)
	}

	// 统计行数
	fmt.Printf("找到 %d 行\n", len(trLocators))

	results := make([]*MedicineInfo, 0, len(trLocators))
	// 遍历每个 <tr> 并读取文本内容
	for i, tr := range trLocators {
		tdLocators, err := tr.Locator("td").All()
		if err != nil {
			log.Printf("无法获取第 %d 行的 <td> 元素: %v", i+1, err)
			continue
		}
		lines := make([]string, 0, len(tdLocators))
		for j, td := range tdLocators {
			tdText, err := td.TextContent()
			if err != nil {
				log.Printf("无法获取第 %d 行第 %d 列文本: %v", i+1, j+1, err)
				tdText = ""
			}
			lines = append(lines, tdText)
		}
		medicine := NewMedicineInfo(lines)
		results = append(results, medicine)
	}
	return results, nil
}

func search_medicine(edge *PlaywrightEdge) int {
	// 打开页面
	edge.Visit("https://www.cde.org.cn/hymlj/listpage/9cd8db3b7530c6fa0c86485e563f93c7")

	// 点击按钮：更多查询条件
	locator, err := edge.WaitForSelector("#moreBtn", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()

	// 选择：上市销售状况 --> 上市销售中
	locator, err = edge.WaitForSelector(".layui-row:nth-child(4) .layui-input.layui-unselect", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()
	locator, err = edge.WaitForSelector(".layui-unselect.layui-form-select.layui-form-selected dd:nth-child(3)", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()

	// 选择：收录类别 --> 进口原研药品
	locator, err = edge.WaitForSelector(".layui-row:nth-child(5) .layui-input.layui-unselect", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()
	locator, err = edge.WaitForSelector(".layui-unselect.layui-form-select.layui-form-selected dd:nth-child(2)", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()

	// 点击按钮：查询
	locator, err = edge.WaitForSelector("#searchForm > .layui-row > div > button:first-child", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()
	time.Sleep(1 * time.Second)

	// 点击按钮：收起更多查询条件
	locator, err = edge.WaitForSelector("#moreBtn", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()
	time.Sleep(1 * time.Second)

	// 设置每页显示 90 条
	locator, err = edge.WaitForSelector(".layui-laypage-limits select", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.SelectOption(playwright.SelectOptionValues{
		Values: &[]string{"90"},
	})

	time.Sleep(1 * time.Second)
	locator, err = edge.WaitForSelector(".layui-laypage-last", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	lastpage, err := locator.InnerText()
	if err != nil {
		log.Fatalf("无法获取最后一页页码: %v", err)
	}
	fmt.Printf("最后一页页码: %s\n", lastpage)
	page, err := strconv.Atoi(strings.TrimSpace(lastpage))
	if err != nil {
		log.Fatalf("无法转换页码: %v", err)
	}
	return page
}

func next_page(edge *PlaywrightEdge) {
	locator, err := edge.WaitForSelector(".layui-laypage-next .layui-icon", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()
}

func YuanYanYao(output_path string) {
	// 启动浏览器
	edge, err := NewPlaywrightEdge(0)
	if err != nil {
		log.Fatalf("无法启动 Edge 浏览器: %v", err)
	}
	defer edge.Close()

	// 搜索进口原研药品
	page := search_medicine(edge)
	fmt.Printf("总页数: %d\n", page)
	medicines := make([]*MedicineInfo, 0)
	for i := 1; i <= page; i++ {
		fmt.Printf("正在获取第 %d 页数据\n", i)
		pageList, err := get_page_data(edge)
		if err != nil {
			log.Fatalf("无法获取第 %d 页数据: %v", i, err)
		}
		medicines = append(medicines, pageList...)
		next_page(edge)
	}

	excel, err := NewSimpleExcelTableWriter(GetMedicineInfoHeaders())
	if err != nil {
		log.Fatalf("无法创建 Excel 文件: %v", err)
	}
	defer excel.Close()

	for _, medicine := range medicines {
		err = excel.WriteRow(medicine.ToRowData())
		if err != nil {
			log.Fatalf("无法写入 Excel 行: %v", err)
		}
	}
	err = excel.SaveAs(output_path)
	if err != nil {
		log.Fatalf("无法保存 Excel 文件: %v", err)
	}
}
