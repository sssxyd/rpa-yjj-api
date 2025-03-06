package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/playwright-community/playwright-go"
)

type MedicineData struct {
	RegisterNo               string // 注册证号
	SourceRegisterNo         string // 原注册证号
	RegisterRemark           string // 注册证号备注
	SubPackageAuthCode       string // 分包装批准文号
	CertHolderCN             string // 上市许可证只有人（中文）
	CertHolderEN             string // 上市许可证只有人（英文）
	CertHolderAddressCN      string // 上市许可证持有人地址（中文）
	CertHolderAddressEN      string // 上市许可证持有人地址（英文）
	CompanyNameCN            string // 公司名称（中文）
	CompanyNameEN            string // 公司名称（英文）
	CompanyAddressCN         string // 地址（中文）
	CompanyAddressEN         string // 地址（英文）
	CompanyRegionCN          string // 国家/地区（中文）
	CompanyRegionEN          string // 国家/地区（英文）
	ProductNameCN            string // 产品名称（中文）
	ProductNameEN            string // 产品名称（英文）
	BrandNameCN              string // 商品名称（中文）
	BrandNameEN              string // 商品名称（英文）
	TorchTypeCN              string // 剂型（中文）
	SpecificationCN          string // 规格（中文）
	PackageSpecCN            string // 包装规格（中文）
	ManufacturerCN           string // 生产厂商（中文）
	ManufacturerEN           string // 生产厂商（英文）
	ManufacturerAddressCN    string // 生产厂商地址（中文）
	ManufacturerAddressEN    string // 生产厂商地址（英文）
	ManufacturerRegionCN     string // 厂商国家/地区（中文）
	ManufacturerRegionEN     string // 厂商国家/地区（英文）
	CertStartDate            string // 发证日期
	CertEndDate              string // 有效期截止日
	SubPackageCompanyName    string // 分包装企业名称
	SubPackageCompanyAddress string // 分包装企业地址
	SubPackageCertStartDate  string // 分包装文号批准日期
	SubPackageCertEndDate    string // 分包装文号有效期截止日
	DrugStandardCode         string // 药品本位码
	ProductCategory          string // 产品类别
	DrugStandardCodeRemark   string // 药品本位码备注
}

func NewMedicineData(lines []string) *MedicineData {
	if len(lines) < 36 {
		for i := len(lines); i < 36; i++ {
			lines = append(lines, "")
		}
	}
	return &MedicineData{
		RegisterNo:               lines[0],
		SourceRegisterNo:         lines[1],
		RegisterRemark:           lines[2],
		SubPackageAuthCode:       lines[3],
		CertHolderCN:             lines[4],
		CertHolderEN:             lines[5],
		CertHolderAddressCN:      lines[6],
		CertHolderAddressEN:      lines[7],
		CompanyNameCN:            lines[8],
		CompanyNameEN:            lines[9],
		CompanyAddressCN:         lines[10],
		CompanyAddressEN:         lines[11],
		CompanyRegionCN:          lines[12],
		CompanyRegionEN:          lines[13],
		ProductNameCN:            lines[14],
		ProductNameEN:            lines[15],
		BrandNameCN:              lines[16],
		BrandNameEN:              lines[17],
		TorchTypeCN:              lines[18],
		SpecificationCN:          lines[19],
		PackageSpecCN:            lines[20],
		ManufacturerCN:           lines[21],
		ManufacturerEN:           lines[22],
		ManufacturerAddressCN:    lines[23],
		ManufacturerAddressEN:    lines[24],
		ManufacturerRegionCN:     lines[25],
		ManufacturerRegionEN:     lines[26],
		CertStartDate:            lines[27],
		CertEndDate:              lines[28],
		SubPackageCompanyName:    lines[29],
		SubPackageCompanyAddress: lines[30],
		SubPackageCertStartDate:  lines[31],
		SubPackageCertEndDate:    lines[32],
		DrugStandardCode:         lines[33],
		ProductCategory:          lines[34],
		DrugStandardCodeRemark:   lines[35],
	}
}

func GetMedicineDataHeaders() []string {
	return []string{
		"注册证号",
		"原注册证号",
		"注册证号备注",
		"分包装批准文号",
		"上市许可证只有人（中文）",
		"上市许可证只有人（英文）",
		"上市许可证持有人地址（中文）",
		"上市许可证持有人地址（英文）",
		"公司名称（中文）",
		"公司名称（英文）",
		"地址（中文）",
		"地址（英文）",
		"国家/地区（中文）",
		"国家/地区（英文）",
		"产品名称（中文）",
		"产品名称（英文）",
		"商品名称（中文）",
		"商品名称（英文）",
		"剂型（中文）",
		"规格（中文）",
		"包装规格（中文）",
		"生产厂商（中文）",
		"生产厂商（英文）",
		"生产厂商地址（中文）",
		"生产厂商地址（英文）",
		"厂商国家/地区（中文）",
		"厂商国家/地区（英文）",
		"发证日期",
		"有效期截止日",
		"分包装企业名称",
		"分包装企业地址",
		"分包装文号批准日期",
		"分包装文号有效期截止日",
		"药品本位码",
		"产品类别",
		"药品本位码备注",
	}
}

func (medicine *MedicineData) ToRowData() []string {
	return []string{
		medicine.RegisterNo,
		medicine.SourceRegisterNo,
		medicine.RegisterRemark,
		medicine.SubPackageAuthCode,
		medicine.CertHolderCN,
		medicine.CertHolderEN,
		medicine.CertHolderAddressCN,
		medicine.CertHolderAddressEN,
		medicine.CompanyNameCN,
		medicine.CompanyNameEN,
		medicine.CompanyAddressCN,
		medicine.CompanyAddressEN,
		medicine.CompanyRegionCN,
		medicine.CompanyRegionEN,
		medicine.ProductNameCN,
		medicine.ProductNameEN,
		medicine.BrandNameCN,
		medicine.BrandNameEN,
		medicine.TorchTypeCN,
		medicine.SpecificationCN,
		medicine.PackageSpecCN,
		medicine.ManufacturerCN,
		medicine.ManufacturerEN,
		medicine.ManufacturerAddressCN,
		medicine.ManufacturerAddressEN,
		medicine.ManufacturerRegionCN,
		medicine.ManufacturerRegionEN,
		medicine.CertStartDate,
		medicine.CertEndDate,
		medicine.SubPackageCompanyName,
		medicine.SubPackageCompanyAddress,
		medicine.SubPackageCertStartDate,
		medicine.SubPackageCertEndDate,
		medicine.DrugStandardCode,
		medicine.ProductCategory,
		medicine.DrugStandardCodeRemark,
	}
}

func search_jinkouyao(edge *PlaywrightEdge) (int, error) {
	// 打开进口原研药列表页面
	err := edge.Visit("https://www.nmpa.gov.cn/datasearch/home-index.html#category=yp")
	if err != nil {
		return 0, err
	}

	time.Sleep(10 * time.Second)

	if err = edge.CurrentPage().Keyboard().Press("Escape"); err != nil {
		log.Printf("按键 escape 失败: %v", err)
	} else {
		log.Println("按键 escape 成功")
	}

	// 点击境外生产药品
	locator, err := edge.WaitForSelector("div.el-col.el-col-8 a[title='境外生产药品']", 10000)
	if err != nil {
		return 0, err
	}
	locator.Click()

	// 定位搜索输入框
	locator, err = edge.WaitForSelector("div.search-input.el-input.el-input-group.el-input-group--append input", 3000)
	if err != nil {
		return 0, err
	}
	locator.Fill("药")

	// 回车搜索
	_, err = edge.OpenNewPage("列表页", func() error {
		return locator.Press("Enter")
	}, 3000)
	if err != nil {
		return 0, err
	}
	// 切换到新页面
	edge.SwitchToNextPage()

	time.Sleep(10 * time.Second)

	if err = edge.CurrentPage().Keyboard().Press("Escape"); err != nil {
		log.Printf("按键 escape 失败: %v", err)
	} else {
		log.Println("按键 escape 成功")
	}

	locator, err = edge.WaitForSelector("div.el-pagination > ul.el-pager", 5000)
	if err != nil {
		log.Fatalf("等待分页元素失败: %v", err)
		return 0, err
	}
	last_page_str, err := locator.Locator("li:last-child").InnerText()
	if err != nil {
		log.Fatalf("无法获取最后一页页码: %v", err)
		return 0, err
	}
	page, err := strconv.Atoi(strings.TrimSpace(last_page_str))
	if err != nil {
		log.Fatalf("无法转换页码: %v", err)
	}
	return page, nil
}

func next_jinkouyao_page(edge *PlaywrightEdge) error {
	// 点击下一页
	locator, err := edge.WaitForSelector("div.el-pagination > button:nth-child(3)", 1000)
	if err != nil {
		log.Fatalf("等待元素失败: %v", err)
	}
	return locator.Click()
}

func wait_for_detail_display(edge *PlaywrightEdge) (playwright.Locator, error) {
	tbody, err := edge.WaitForSelector("table > tbody", 10000)
	if err != nil {
		time.Sleep(1 * time.Second)
		edge.CurrentPage().Reload()
		return nil, nil
	}

	for i := range 5 {
		registerNo, err := tbody.Locator("tr:nth-child(1) > td:nth-child(2) > div > div").InnerText()
		if err == nil && registerNo != "" {
			return tbody, nil
		}
		time.Sleep(1 * time.Second)
		log.Printf("...第 %d 次等待详情页注册证号", i+1)
	}
	time.Sleep(10 * time.Second)
	edge.CurrentPage().Reload()

	registerNo, err := tbody.Locator("tr:nth-child(1) > td:nth-child(2) > div > div").InnerText()
	if err == nil && registerNo != "" {
		return tbody, nil
	}
	return nil, nil
}

func collect_jinkouyao_data(edge *PlaywrightEdge) (*MedicineData, error) {
	var tbody playwright.Locator
	for i := range 30 {
		locator, err := wait_for_detail_display(edge)
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
		return NewMedicineData(make([]string, 0, 36)), nil
	}

	items, err := tbody.Locator("tr").All()
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, 36)
	for idx, item := range items {
		if idx > 35 {
			break
		}
		valstr, err := item.Locator("td:nth-child(2)").InnerText()
		if err != nil {
			return nil, err
		}
		lines = append(lines, valstr)
	}
	return NewMedicineData(lines), nil
}

func get_page_jinkouyao(edge *PlaywrightEdge, pageNo int) ([]*MedicineData, error) {
	locator, err := edge.WaitForSelector("table > tbody", 1000)
	if err != nil {
		return nil, err
	}
	trs, err := locator.Locator("tr").All()
	if err != nil {
		return nil, err
	}
	log.Printf("第%d页共 %d 条数据", pageNo, len(trs))
	medicines := make([]*MedicineData, 0, len(trs))
	for idx, tr := range trs {
		registerNo, err := tr.Locator("td:nth-child(2) > div > p").InnerText()
		if err != nil || strings.TrimSpace(registerNo) == "" {
			log.Printf("第 %d 页第 %d 条数据注册证号为空，跳过", pageNo, idx+1)
			continue
		}
		// 点击详情按钮
		btn := tr.Locator("td:nth-child(5) > div > button")
		_, err = edge.OpenNewPage("详情页", func() error {
			return btn.Click()
		}, 10000)
		if err != nil {
			return nil, err
		}
		// 切换到详情页
		edge.SwitchToNextPage()
		log.Printf("正在获取第 %d 页第 %d 条数据", pageNo, idx+1)
		medicine_data, err := collect_jinkouyao_data(edge)
		if err != nil {
			return nil, err
		}
		if medicine_data != nil {
			log.Printf("...采集药品 %s %s", medicine_data.ProductNameCN, medicine_data.RegisterNo)
			medicines = append(medicines, medicine_data)
		}
		// 返回列表页
		edge.SwitchToPreviousPage()
		// 关闭详情页
		edge.ClosePage("详情页")
	}
	return medicines, nil
}

func go_to_page(edge *PlaywrightEdge, pageNo int) error {
	locator, err := edge.WaitForSelector("div.el-input.el-pagination__editor > input", 1000)
	if err != nil {
		log.Fatalf("等待分页元素失败: %v", err)
		return err
	}
	err = locator.Clear()
	if err != nil {
		log.Fatalf("无法清空页码输入框: %v", err)
	}
	err = locator.Fill(fmt.Sprint(pageNo))
	if err != nil {
		log.Fatalf("无法定位页码: %v", err)
		return err
	}
	err = locator.Press("Enter")
	if err != nil {
		log.Fatalf("无法跳转到第 %d 页: %v", pageNo, err)
		return err
	}
	locator, err = edge.WaitForSelector("table > tbody", 5000)
	if err != nil {
		log.Fatalf("等待表格元素失败: %v", err)
		return err
	}
	return locator.Click()
}

func clear_local_storage(edge *PlaywrightEdge) error {
	// ---------- 清除所有存储 ----------
	page := edge.CurrentPage()
	// 1. 清除 localStorage 和 sessionStorage
	if _, err := page.Evaluate("localStorage.clear()"); err != nil {
		log.Fatalf("清空 localStorage 失败: %v", err)
	}
	if _, err := page.Evaluate("sessionStorage.clear()"); err != nil {
		log.Fatalf("清空 sessionStorage 失败: %v", err)
	}

	// 2. 清除 Cookies（针对当前域名）
	if err := edge.context.ClearCookies(); err != nil {
		log.Fatalf("清除 Cookies 失败: %v", err)
	}

	// 3. 清除 IndexedDB（通过 JavaScript）
	if _, err := page.Evaluate(`
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

func CollectImportDrugs(output_path string, start_page int, end_page int) {
	edge, err := NewPlaywrightEdge(0)
	if err != nil {
		log.Fatalf("无法启动 Edge 浏览器: %v", err)
	}
	defer edge.Close()

	pageCount, err := search_jinkouyao(edge)
	if err != nil {
		log.Fatalf("搜索失败: %v", err)
	}
	log.Printf("共 %d 页", pageCount)
	if end_page > pageCount || end_page <= 0 {
		end_page = pageCount
	}

	go_to_page(edge, start_page)

	medicines := make([]*MedicineData, 0)
	for i := start_page; i <= end_page; i++ {
		log.Printf("正在获取第 %d 页数据", i)
		data_list, err := get_page_jinkouyao(edge, i)
		if err != nil {
			log.Fatalf("获取第 %d 页数据失败: %v", i, err)
		}
		medicines = append(medicines, data_list...)
		log.Printf("第 %d 页数据获取完毕", i)
		if i == end_page {
			log.Println("已到达最后一页")
			break
		}
		err = next_jinkouyao_page(edge)
		if err != nil {
			log.Printf("跳转到第 %d 页失败: %v", i+1, err)
			break
		}
	}
	log.Printf("共 %d 条数据", len(medicines))

	excel, err := NewSimpleExcelTableWriter(GetMedicineDataHeaders())
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

	err = clear_local_storage(edge)
	if err != nil {
		log.Fatalf("清除存储失败: %v", err)
	}

}
