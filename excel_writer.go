package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

// IndexToExcelColumn 将从 0 开始的索引转换为 Excel 列名（如 0 -> A, 1 -> B, 25 -> Z, 26 -> AA）
func indexToExcelColumn(index int) string {
	if index < 0 {
		return "" // 无效索引返回空字符串
	}

	result := ""
	for {
		// 计算当前位的字符（A=0, B=1, ..., Z=25）
		char := 'A' + rune(index%26)
		result = string(char) + result

		// 更新 index，除以 26 并减 1（因为是 0-based 到 1-based 的转换）
		index = index/26 - 1

		// 如果 index 小于 0，说明已经处理完所有位
		if index < 0 {
			break
		}
	}
	return result
}

// estimateWidth 估算字符串的宽度
func estimateWidth(value string) float64 {
	// 使用 rune 处理多字节字符（如中文）
	sentence_length := float64(0)
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			sentence_length += 2
		} else {
			sentence_length += 1
		}
	}
	return sentence_length + 2 // 留出一些空白
}

type SimpleExcelTableWriter struct {
	file      *excelize.File
	sheet     string
	rowIndex  int
	rowStyle  int
	colWidths []float64 // 记录每列的最大宽度
}

func NewSimpleExcelTableWriter(headers []string) (*SimpleExcelTableWriter, error) {
	f := excelize.NewFile()
	rowIndex := 0
	if len(headers) > 0 {
		f.SetSheetRow("Sheet1", "A1", &headers)
		// 添加简单的样式：表头加粗并设置背景颜色
		headStyle, err := f.NewStyle(&excelize.Style{
			// 设置单元格格式为字符串（防止数字被自动转换为其他格式）
			NumFmt: 49, // 49 是内置的 "@" 格式，表示纯文本/字符串

			// 设置对齐方式：垂直居中
			Alignment: &excelize.Alignment{
				Vertical:   "center",
				Horizontal: "center",
			},

			// 设置字体：加粗
			Font: &excelize.Font{Bold: true},

			// 设置填充：背景颜色
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0EBF5"}, Pattern: 1},

			// 设置边框：实线
			Border: []excelize.Border{
				{Type: "left", Style: 1, Color: "000000"},   // 左边框，实线，黑色
				{Type: "top", Style: 1, Color: "000000"},    // 上边框，实线，黑色
				{Type: "right", Style: 1, Color: "000000"},  // 右边框，实线，黑色
				{Type: "bottom", Style: 1, Color: "000000"}, // 下边框，实线，黑色
			},
		})
		if err != nil {
			return nil, err
		}
		f.SetCellStyle("Sheet1", "A1", fmt.Sprintf("%s1", indexToExcelColumn(len(headers)-1)), headStyle)
		rowIndex = 1
	}

	// 创建数据行的样式：字符串格式，垂直居中，实线边框
	dataStyle, err := f.NewStyle(&excelize.Style{
		// 设置单元格格式为字符串（防止数字被自动转换为其他格式）
		NumFmt: 49, // 49 是内置的 "@" 格式，表示纯文本/字符串

		// 设置对齐方式：垂直居中
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "left",
		},

		// 设置边框：实线
		Border: []excelize.Border{
			{Type: "left", Style: 1, Color: "000000"},   // 左边框，实线，黑色
			{Type: "top", Style: 1, Color: "000000"},    // 上边框，实线，黑色
			{Type: "right", Style: 1, Color: "000000"},  // 右边框，实线，黑色
			{Type: "bottom", Style: 1, Color: "000000"}, // 下边框，实线，黑色
		},
	})

	if err != nil {
		return nil, err
	}

	// 初始化列宽数组，基于表头内容长度
	colWidths := make([]float64, len(headers))
	for i, header := range headers {
		colWidths[i] = max(estimateWidth(header), 8) // 最小宽度 8
		f.SetColWidth("Sheet1", indexToExcelColumn(i), indexToExcelColumn(i), colWidths[i])
	}

	return &SimpleExcelTableWriter{file: f, sheet: "Sheet1", rowIndex: rowIndex, rowStyle: dataStyle, colWidths: colWidths}, nil
}

func (w *SimpleExcelTableWriter) WriteRow(row []string) error {
	w.rowIndex++
	if len(row) == 0 {
		return nil // 空行不处理
	}
	for colIndex, cellValue := range row {
		cell := fmt.Sprintf("%s%d", indexToExcelColumn(colIndex), w.rowIndex)
		err := w.file.SetCellValue(w.sheet, cell, cellValue)
		if err != nil {
			return err
		}
		err = w.file.SetCellStyle(w.sheet, cell, cell, w.rowStyle)
		if err != nil {
			return err
		}
	}

	// 更新列宽
	for colIndex, cellValue := range row {
		if colIndex >= len(w.colWidths) {
			// 如果列超出当前记录，扩展 colWidths
			w.colWidths = append(w.colWidths, 8) // 默认最小宽度
		}

		sublines := strings.Split(cellValue, "\n")
		newWidth := w.colWidths[colIndex]
		for _, subline := range sublines {
			sublineWidth := estimateWidth(subline)
			if sublineWidth > newWidth {
				newWidth = sublineWidth
			}
		}

		if newWidth > w.colWidths[colIndex] {
			w.colWidths[colIndex] = newWidth
			w.file.SetColWidth(w.sheet, indexToExcelColumn(colIndex), indexToExcelColumn(colIndex), newWidth)
		}
	}

	// 设置行高（可选，基于内容行数估算）
	lines := 1 // 默认单行
	for _, cellValue := range row {
		lineCount := strings.Count(cellValue, "\n") + 1
		if lineCount > lines {
			lines = lineCount
		}
	}
	if lines > 1 {
		w.file.SetRowHeight(w.sheet, w.rowIndex, float64(lines)*15) // 每行约 15 单位高度
	}

	return nil
}

func (w *SimpleExcelTableWriter) SaveAs(filename string) error {
	return w.file.SaveAs(filename)
}

func (w *SimpleExcelTableWriter) Close() {
	w.file = nil
}
