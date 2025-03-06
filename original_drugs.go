package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type OriginalDrug struct {
	ActiveIngredients            string // 活性成分
	ActiveIngredientsEN          string // 活性成分（英文）
	DrugName                     string // 药品名称
	DrugNameEN                   string // 药品名称（英文）
	ProductName                  string // 商品名
	ProductNameEN                string // 商品名（英文）
	TorchType                    string // 剂型
	DrugDeliveryRoute            string // 给药途径
	Specification                string // 规格
	ReferenceProduct             string // 参比制剂
	ATCCode                      string // ATC码
	AuthCode                     string // 批准文号/注册证号
	CertDate                     string // 批准日期
	MarketingAuthorizationHolder string // 上市许可持有人
	Manufacturer                 string // 生产厂商
	MarketingSalesStatus         string // 上市销售状态
	Category                     string // 收录类别
	InstructionBook              string // 说明书
	ReviewReport                 string // 审评报告
}

func NewOriginalDrug(lines []string) *OriginalDrug {
	if len(lines) < 9 {
		for i := len(lines); i < 9; i++ {
			lines = append(lines, "")
		}
	}
	return &OriginalDrug{
		ActiveIngredients:            lines[0],
		ActiveIngredientsEN:          lines[1],
		DrugName:                     lines[2],
		DrugNameEN:                   lines[3],
		ProductName:                  lines[4],
		ProductNameEN:                lines[5],
		TorchType:                    lines[6],
		DrugDeliveryRoute:            lines[7],
		Specification:                lines[8],
		ReferenceProduct:             lines[9],
		ATCCode:                      lines[10],
		AuthCode:                     lines[11],
		CertDate:                     lines[12],
		MarketingAuthorizationHolder: lines[13],
		Manufacturer:                 lines[14],
		MarketingSalesStatus:         lines[15],
		Category:                     lines[16],
		InstructionBook:              lines[17],
		ReviewReport:                 lines[18],
	}
}

func GetOriginalDrugHeaders() []string {
	return []string{
		"活性成分",
		"活性成分（英文）",
		"药品名称",
		"药品名称（英文）",
		"商品名",
		"商品名（英文）",
		"剂型",
		"给药途径",
		"规格",
		"参比制剂",
		"ATC码",
		"批准文号/注册证号",
		"批准日期",
		"上市许可持有人",
		"生产厂商",
		"上市销售状态",
		"收录类别",
		"说明书",
		"审评报告",
	}
}

func (medicine *OriginalDrug) ToRowData() []string {
	return []string{
		medicine.ActiveIngredients,
		medicine.ActiveIngredientsEN,
		medicine.DrugName,
		medicine.DrugNameEN,
		medicine.ProductName,
		medicine.ProductNameEN,
		medicine.TorchType,
		medicine.DrugDeliveryRoute,
		medicine.Specification,
		medicine.ReferenceProduct,
		medicine.ATCCode,
		medicine.AuthCode,
		medicine.CertDate,
		medicine.MarketingAuthorizationHolder,
		medicine.Manufacturer,
		medicine.MarketingSalesStatus,
		medicine.Category,
		medicine.InstructionBook,
		medicine.ReviewReport,
	}
}

func od_wait_for_detail_display(edge *PlaywrightEdge) (playwright.Locator, error) {
	tbody, err := edge.WaitForSelector("table > tbody", 10000)
	if err != nil {
		time.Sleep(1 * time.Second)
		edge.CurrentPage().Reload()
		return nil, nil
	}

	for i := range 5 {
		registerNo, err := tbody.Locator("tr:nth-child(12) > td:nth-child(2)").InnerText()
		if err == nil && registerNo != "" {
			return tbody, nil
		}
		time.Sleep(1 * time.Second)
		log.Printf("...第 %d 次等待详情页注册证号", i+1)
	}
	time.Sleep(10 * time.Second)
	edge.CurrentPage().Reload()

	registerNo, err := tbody.Locator("tr:nth-child(12) > td:nth-child(2)").InnerText()
	if err == nil && registerNo != "" {
		return tbody, nil
	}
	return nil, nil
}

func od_get_drug_detail(edge *PlaywrightEdge) (*OriginalDrug, error) {
	var tbody playwright.Locator
	for i := range 30 {
		locator, err := od_wait_for_detail_display(edge)
		if err != nil {
			return nil, err
		}
		if locator != nil {
			tbody = locator
			break
		}
		log.Printf("第 %d 次等待详情页显示", i+1)
	}
	if tbody == nil {
		return NewOriginalDrug(make([]string, 0, 19)), nil
	}

	items, err := tbody.Locator("tr").All()
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, 19)
	for idx, item := range items {
		if idx > 18 {
			break
		}
		valstr, err := item.Locator("td:nth-child(2)").InnerText()
		if err != nil {
			return nil, err
		}
		lines = append(lines, valstr)
	}
	return NewOriginalDrug(lines), nil
}

func od_get_page_data(edge *PlaywrightEdge, pageNo int) ([]*OriginalDrug, error) {

	locator := edge.CurrentPage().Locator(".layui-table-body.layui-table-main tr")

	// 获取所有匹配的 tr 元素
	trs, err := locator.All()
	if err != nil {
		log.Fatalf("无法获取所有 <tr> 元素: %v", err)
	}
	log.Printf("第%d页共 %d 条数据", pageNo, len(trs))

	// 统计行数
	fmt.Printf("找到 %d 行\n", len(trs))
	medicines := make([]*OriginalDrug, 0, len(trs))
	// 遍历每个 <tr> 并读取文本内容
	for i, tr := range trs {
		registerNo, err := tr.Locator("td:nth-child(2) > div").InnerText()
		if err != nil || strings.TrimSpace(registerNo) == "" {
			log.Printf("第 %d 页第 %d 条数据注册证号为空，跳过", pageNo, i+1)
			if err != nil {
				log.Printf("无法获取注册证号: %v", err)
			}
			break
		}
		// 点击详情
		btn := tr.Locator("td:nth-child(3) > div > a")
		_, err = edge.OpenNewPage("详情页", func() error {
			return btn.Click()
		}, 30000)
		if err != nil {
			return nil, err
		}
		// 切换到详情页
		edge.SwitchToNextPage()
		log.Printf("正在获取第 %d 页第 %d 条数据", pageNo, i+1)
		medicine_data, err := od_get_drug_detail(edge)
		if err != nil {
			return nil, err
		}
		if medicine_data != nil {
			log.Printf("...采集药品 %s %s", medicine_data.DrugName, medicine_data.AuthCode)
			medicines = append(medicines, medicine_data)
		}
		// 返回列表页
		edge.SwitchToPreviousPage()
		// 关闭详情页
		edge.ClosePage("详情页")
	}
	return medicines, nil
}

func od_search_medicine(edge *PlaywrightEdge) int {
	// 打开页面
	edge.Visit("https://www.cde.org.cn/hymlj/listpage/9cd8db3b7530c6fa0c86485e563f93c7")

	// 点击按钮：更多查询条件
	locator, err := edge.WaitForSelector("#moreBtn", 10000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()

	// 选择：上市销售状况 --> 上市销售中
	// locator, err = edge.WaitForSelector(".layui-row:nth-child(4) .layui-input.layui-unselect", 1000)
	// if err != nil {
	// 	log.Fatalf("等待元素失败: %v", err)
	// }
	// locator.Click()
	// locator, err = edge.WaitForSelector(".layui-unselect.layui-form-select.layui-form-selected dd:nth-child(3)", 1000)
	// if err != nil {
	// 	log.Fatalf("等待元素失败: %v", err)
	// }
	// locator.Click()

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

	time.Sleep(3 * time.Second)
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

func od_next_page(edge *PlaywrightEdge) {
	locator, err := edge.WaitForSelector(".layui-laypage-next .layui-icon", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	locator.Click()
}

func CollectOriginalDrugs(output_path string, start_page int, end_page int) {
	// 启动浏览器
	edge, err := NewPlaywrightEdge(0)
	if err != nil {
		log.Fatalf("无法启动 Edge 浏览器: %v", err)
	}
	defer edge.Close()

	// 搜索进口原研药品
	total_page := od_search_medicine(edge)
	fmt.Printf("总页数: %d\n", total_page)
	if end_page <= 0 || end_page > total_page {
		end_page = total_page
	}
	medicines := make([]*OriginalDrug, 0)

	for i := 1; i < start_page; i++ {
		od_next_page(edge)
	}
	for i := start_page; i <= end_page; i++ {
		fmt.Printf("正在获取第 %d 页数据\n", i)
		pageList, err := od_get_page_data(edge, i)
		if err != nil {
			log.Fatalf("无法获取第 %d 页数据: %v", i, err)
		}
		medicines = append(medicines, pageList...)
		if i < end_page {
			od_next_page(edge)
		}
	}

	log.Printf("共 %d 条数据", len(medicines))
	edge.ClearLocalData()

	excel, err := NewSimpleExcelTableWriter(GetOriginalDrugHeaders())
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
