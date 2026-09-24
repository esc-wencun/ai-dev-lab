package excel

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// testColumns 测试列定义（含值转换、默认列宽、自定义列宽）。
var testColumns = []Column{
	{Title: "用户序号", Field: "userId", Width: 5},
	{Title: "用户名称", Field: "userName"},
	{Title: "性别", Field: "sex", Converter: map[string]string{"0": "男", "1": "女", "2": "未知"}},
}

// testRows 测试数据（值类型覆盖 string/int/int64/nil）。
func testRows() []map[string]any {
	return []map[string]any{
		{"userId": int64(1), "userName": "admin", "sex": "0"},
		{"userId": int(2), "userName": "ry", "sex": "1"},
		{"userId": int64(3), "userName": "游客", "sex": nil},
	}
}

// TestExportHeadersAndResponse 验证响应头三元组与 xlsx 可解析回读。
func TestExportHeadersAndResponse(t *testing.T) {
	w := httptest.NewRecorder()
	Export(w, "用户数据", testColumns, testRows())

	// 响应头对位 FileUtils.setAttachmentResponseHeader（前端按 download-filename 取名）
	h := w.Header()
	if ct := h.Get("Content-Type"); !strings.Contains(ct, "spreadsheetml") {
		t.Errorf("Content-Type = %q", ct)
	}
	if cd := h.Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment;") || !strings.Contains(cd, "filename*=utf-8''") {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if expose := h.Get("Access-Control-Expose-Headers"); !strings.Contains(expose, "download-filename") {
		t.Errorf("Access-Control-Expose-Headers = %q", expose)
	}
	if fn := h.Get("download-filename"); fn != "%E7%94%A8%E6%88%B7%E6%95%B0%E6%8D%AE.xlsx" {
		t.Errorf("download-filename = %q", fn)
	}
	if w.Code != 200 {
		t.Errorf("HTTP status = %d", w.Code)
	}

	// 读回验证：表头 + 数据 + Converter 正查
	rows, err := readAll(w.Body.Bytes())
	if err != nil {
		t.Fatalf("读回失败: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("行数 = %d, want 4（表头+3数据）", len(rows))
	}
	wantHeader := []string{"用户序号", "用户名称", "性别"}
	for i, want := range wantHeader {
		if rows[0][i] != want {
			t.Errorf("表头[%d] = %q, want %q", i, rows[0][i], want)
		}
	}
	if rows[1][0] != "1" || rows[1][1] != "admin" || rows[1][2] != "男" {
		t.Errorf("数据行1 = %v", rows[1])
	}
	if rows[2][2] != "女" {
		t.Errorf("数据行2 Converter 未生效: %v", rows[2])
	}
	if rows[3][2] != "" {
		t.Errorf("nil 值应为空串: %v", rows[3])
	}
}

// TestConverterUnknownValue Converter 未收录的值原样输出（对位 Java converterExp 查不到不换）。
func TestConverterUnknownValue(t *testing.T) {
	w := httptest.NewRecorder()
	Export(w, "数据", testColumns, []map[string]any{{"userId": int64(9), "userName": "x", "sex": "9"}})
	rows, err := readAll(w.Body.Bytes())
	if err != nil {
		t.Fatalf("读回失败: %v", err)
	}
	if rows[1][2] != "9" {
		t.Errorf("未收录值应原样输出, got %q", rows[1][2])
	}
}

// TestParse 导入解析：表头定位、Converter 反查、空行跳过、未知标题列忽略。
func TestParse(t *testing.T) {
	w := httptest.NewRecorder()
	Export(w, "用户数据", testColumns, []map[string]any{
		{"userId": int64(1), "userName": "admin", "sex": "0"},
		{"userId": int64(2), "userName": "ry", "sex": "女"},
	})
	rows, err := Parse(w.Body.Bytes(), testColumns)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("解析行数 = %d, want 2", len(rows))
	}
	if rows[0]["userId"] != "1" || rows[0]["userName"] != "admin" {
		t.Errorf("行1 = %v", rows[0])
	}
	if rows[0]["sex"] != "0" {
		t.Errorf("导出显示值'男'应反查为编码'0', got %q", rows[0]["sex"])
	}
	// 行2 的 sex 本来就是"女"（显示值），反查回编码
	if rows[1]["sex"] != "1" {
		t.Errorf("显示值'女'应反查为'1', got %q", rows[1]["sex"])
	}
}

// TestParseDirectCodes 用户直接填编码时保留原值（对齐 Python 版行为）。
func TestParseDirectCodes(t *testing.T) {
	// 导出 sex=编码"2"（Converter 查不到→原样输出编码），Parse 反查也查不到→保留编码
	rows, err := Parse(exportSexRaw("2"), testColumns)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if rows[0]["sex"] != "2" {
		t.Errorf("编码原值应保留, got %q", rows[0]["sex"])
	}
}

// exportSexRaw 导出 sex=编码值 的 xlsx（Converter 不改写时显示编码）。
func exportSexRaw(code string) []byte {
	w := httptest.NewRecorder()
	Export(w, "数据", []Column{{Title: "用户序号", Field: "userId"}, {Title: "性别", Field: "sex", Converter: map[string]string{"0": "男"}}},
		[]map[string]any{{"userId": int64(5), "sex": code}})
	// 导出时 Converter 正查 "2"→查不到→原样输出 "2"
	return w.Body.Bytes()
}

// readAll 解析 xlsx 为字符串矩阵（读回验证辅助；短行按 cols 长度补空串）。
func readAll(data []byte) ([][]string, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	it, err := f.Rows(f.GetSheetList()[0])
	if err != nil {
		return nil, err
	}
	defer it.Close()
	var out [][]string
	for it.Next() {
		row, err := it.Columns()
		if err != nil {
			return nil, err
		}
		for len(row) < len(testColumns) {
			row = append(row, "")
		}
		out = append(out, row)
	}
	return out, nil
}

// TestParseHeaderNotFound 无表头报错。
func TestParseHeaderNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	Export(w, "数据", []Column{{Title: "甲", Field: "a"}}, []map[string]any{{"a": "1"}})
	if _, err := Parse(w.Body.Bytes(), testColumns); err == nil {
		t.Error("无匹配表头应报错")
	}
}

// TestExportTemplate 模板下载：只有表头（+示例行）。
func TestExportTemplate(t *testing.T) {
	w := httptest.NewRecorder()
	ExportTemplate(w, "用户数据", testColumns, []map[string]any{{"userId": int64(1), "userName": "示例", "sex": "0"}})
	rows, err := readAll(w.Body.Bytes())
	if err != nil {
		t.Fatalf("读回失败: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("模板行数 = %d, want 2（表头+1示例）", len(rows))
	}
}

// TestRespJSONError 导出失败场景回 JSON 错误信封（构造非法 sheetName 不易触发，
// 这里直接验证 respErr 输出格式约定）。
func TestRespJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	respErr(w, "boom")
	if w.Code != 200 {
		t.Errorf("HTTP 应 200, got %d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("非 JSON: %s", w.Body.String())
	}
	if body["code"].(float64) != 500 {
		t.Errorf("code = %v, want 500", body["code"])
	}
}
