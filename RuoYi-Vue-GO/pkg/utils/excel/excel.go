// Package excel Excel 导入导出工具（对位 Java ExcelUtil + @Excel 注解，基于 excelize）。
//
// 业务模块用法（对位 @Excel 注解逐字段声明列定义）：
//
//	cols := []excel.Column{
//	    {Title: "用户序号", Field: "userId", Width: 5},
//	    {Title: "用户名称", Field: "userName"},
//	    {Title: "性别", Field: "sex", Converter: map[string]string{"0": "男", "1": "女", "2": "未知"}},
//	}
//	excel.Export(c, "用户数据", cols, rows)            // POST /xxx/export 流式下载
//	list := excel.Parse(data, cols)                   // 导入解析 → []map[string]string
//
// 响应头对齐 Java FileUtils.setAttachmentResponseHeader（前端 request.js download
// 与 plugins/download.js 按 blob + Content-Disposition/download-filename 处理）。
package excel

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Column 导出列定义（对位 @Excel 注解的导出必需属性；sort 由切片顺序承载）。
type Column struct {
	Title     string            // 表头标题（对位 name()）
	Field     string            // 数据字段名（rows 元素 map 的 key，驼峰对齐前端）
	Converter map[string]string // 值转换映射（对位 readConverterExp()/dictType()，导出正查、导入反查）
	Width     float64           // 列宽（对位 width() 默认 16）
}

// sheetNameDefault 默认工作表名。
const sheetNameDefault = "Sheet1"

// Export 写出 xlsx 下载响应（对位 exportExcel(response, list, sheetName)）。
// w 传 gin.Context（实现了 http.ResponseWriter）；rows 元素为 map[string]any
// （service 层从 vo/do 转好），值经 fmt.Sprint 兜底格式化。
func Export(w http.ResponseWriter, sheetName string, cols []Column, rows []map[string]any) {
	data, err := buildXlsx(sheetName, cols, rows)
	if err != nil {
		respErr(w, "导出Excel失败："+err.Error())
		return
	}
	writeAttachment(w, data, sheetName+".xlsx")
}

// ExportTemplate 导入模板下载（对位 importTemplateExcel：表头 + 可选示例行）。
func ExportTemplate(w http.ResponseWriter, sheetName string, cols []Column, exampleRows []map[string]any) {
	Export(w, sheetName, cols, exampleRows)
}

// Parse 解析上传的 xlsx（对位 importExcel：按表头标题映射列，返回字段名→单元格值）。
//   - 表头行：首个含已知标题的行（对齐 Python 版容忍标题前置行）；
//   - 空行跳过；Converter 反查（"男"→"0"），查不到保留原值（用户直接填编码）；
//   - 全部值以 string 承载，业务侧按需自行转换（对位 Java 导入先得 string 再转类型）。
func Parse(data []byte, cols []Column) ([]map[string]string, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("文件不含工作表")
	}
	it, err := f.Rows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}
	defer it.Close()

	titleToCol := map[string]Column{}
	for _, col := range cols {
		titleToCol[col.Title] = col
	}

	var idxCols []indexedColumn
	var result []map[string]string
	for it.Next() {
		row, err := it.Columns()
		if err != nil {
			return nil, fmt.Errorf("读取行失败: %w", err)
		}
		if len(idxCols) == 0 {
			// 未定位表头：本行含任一已知标题则视为表头行
			for i, cell := range row {
				if _, ok := titleToCol[strings.TrimSpace(cell)]; ok {
					idxCols = append(idxCols, indexedColumn{i, titleToCol[strings.TrimSpace(cell)]})
				}
			}
			continue
		}
		if emptyRow(row) {
			continue
		}
		item := map[string]string{}
		for _, ic := range idxCols {
			v := ""
			if ic.idx < len(row) {
				v = strings.TrimSpace(row[ic.idx])
			}
			if ic.col.Converter != nil && v != "" {
				if code, ok := reverseLookup(ic.col.Converter, v); ok {
					v = code
				}
			}
			item[ic.col.Field] = v
		}
		result = append(result, item)
	}
	if len(idxCols) == 0 {
		return nil, fmt.Errorf("未识别到有效的表头行")
	}
	return result, nil
}

// indexedColumn 列下标与定义的配对。
type indexedColumn struct {
	idx int
	col Column
}

// buildXlsx 生成 xlsx 字节（表头样式对位 annotationHeaderStyles：灰底白字加粗居中）。
func buildXlsx(sheetName string, cols []Column, rows []map[string]any) ([]byte, error) {
	f := excelize.NewFile()
	if sheetName == "" {
		sheetName = sheetNameDefault
	}
	if _, err := f.NewSheet(sheetName); err != nil {
		return nil, err
	}
	if err := f.DeleteSheet(sheetNameDefault); err != nil {
		return nil, err
	}

	// 表头：GREY_50_PERCENT 背景（IndexedColors 22 = "FF808080"）+ 白字 + 加粗 + 居中
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#808080"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#808080", Style: 1},
			{Type: "right", Color: "#808080", Style: 1},
			{Type: "top", Color: "#808080", Style: 1},
			{Type: "bottom", Color: "#808080", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}
	for i, col := range cols {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		colName, _ := excelize.ColumnNumberToName(i + 1)
		if err := f.SetCellValue(sheetName, cell, col.Title); err != nil {
			return nil, err
		}
		w := col.Width
		if w == 0 {
			w = 16 // 对位 @Excel width() 默认值
		}
		if err := f.SetColWidth(sheetName, colName, colName, w); err != nil {
			return nil, err
		}
		if err := f.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
	}
	// 数据行（对位 getTargetValue：Converter 正查，日期等格式化转字符串）
	for r, row := range rows {
		for i, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, r+2)
			if err := f.SetCellValue(sheetName, cell, cellValue(row[col.Field], col)); err != nil {
				return nil, err
			}
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// cellValue 单元格值：Converter 正查 + 非字符串兜底格式化。
func cellValue(v any, col Column) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprint(v)
	if col.Converter != nil {
		if mapped, ok := col.Converter[s]; ok {
			return mapped
		}
	}
	return s
}

// reverseLookup 导入反查：显示值 → 编码（对位 reverseByExp）。
func reverseLookup(converter map[string]string, display string) (string, bool) {
	for code, disp := range converter {
		if disp == display {
			return code, true
		}
	}
	return "", false
}

// emptyRow 全空行判断。
func emptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// respErr 错误信封（对位 ResponseUtil Error：HTTP 200 + body code 500）。
// 本包不依赖 gin 类型，gin.Context 的 JSON 方法签名兼容时直接用，否则手写 body。
func respErr(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	fmt.Fprintf(w, `{"code":500,"msg":%q}`, msg) //nolint:errcheck
}

// writeAttachment 写下载响应头+文件流（对位 FileUtils.setAttachmentResponseHeader：
// percentEncode 空格转 %20、Expose-Headers、download-filename 供前端 plugins/download.js 用）。
func writeAttachment(w http.ResponseWriter, data []byte, filename string) {
	enc := url.PathEscape(filename)
	enc = strings.ReplaceAll(enc, "+", "%20")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+enc+";filename*=utf-8''"+enc)
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition,download-filename")
	w.Header().Set("download-filename", enc)
	w.Write(data) //nolint:errcheck // 客户端断开时写失败，无需处理
}
