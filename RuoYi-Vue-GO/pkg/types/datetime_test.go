package types

import (
	"encoding/json"
	"testing"
	"time"
)

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	tt, err := time.ParseInLocation(DateTimeLayout, s, time.Local)
	if err != nil {
		t.Fatalf("测试时间解析失败: %v", err)
	}
	return tt
}

// TestMarshalJSON 固定输出 yyyy-MM-dd HH:mm:ss
func TestMarshalJSON(t *testing.T) {
	d := DateTime(mustParse(t, "2026-09-24 12:30:45"))
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	if string(b) != `"2026-09-24 12:30:45"` {
		t.Errorf("应输出 yyyy-MM-dd HH:mm:ss，实际 %s", b)
	}
}

// TestMarshalZero 零值输出 null（对位 Java null Date）
func TestMarshalZero(t *testing.T) {
	b, err := json.Marshal(DateTime{})
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	if string(b) != "null" {
		t.Errorf("零值应输出 null，实际 %s", b)
	}
}

// TestUnmarshalJSON 完整格式 / 日期格式 / 空串 / null
func TestUnmarshalJSON(t *testing.T) {
	var d DateTime
	if err := json.Unmarshal([]byte(`"2026-09-24 12:30:45"`), &d); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}
	if d.Time() != mustParse(t, "2026-09-24 12:30:45") {
		t.Errorf("完整格式解析不符: %v", d.Time())
	}

	var d2 DateTime
	if err := json.Unmarshal([]byte(`"2026-09-24"`), &d2); err != nil {
		t.Fatalf("日期格式 Unmarshal 失败: %v", err)
	}
	if d2.Time() != mustParse(t, "2026-09-24 00:00:00") {
		t.Errorf("日期格式解析不符: %v", d2.Time())
	}

	var d3 DateTime
	if err := json.Unmarshal([]byte(`""`), &d3); err != nil || !d3.IsZero() {
		t.Errorf("空串应归零: err=%v zero=%v", err, d3.IsZero())
	}

	var d4 DateTime
	if err := json.Unmarshal([]byte(`null`), &d4); err != nil || !d4.IsZero() {
		t.Errorf("null 应归零: err=%v zero=%v", err, d4.IsZero())
	}

	var d5 DateTime
	if err := json.Unmarshal([]byte(`"not-a-date"`), &d5); err == nil {
		t.Error("非法时间应报错")
	}
}

// TestStructRoundTrip struct 内嵌 DateTime 的 JSON 往返（对位 vo 响应/请求）
func TestStructRoundTrip(t *testing.T) {
	type demo struct {
		ID        int64    `json:"userId"`
		Name      string   `json:"userName"`
		CreatedAt DateTime `json:"createTime"`
	}
	in := demo{ID: 9007199254740993, Name: "admin", CreatedAt: DateTime(mustParse(t, "2026-09-24 08:00:00"))}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	// int64 主键直接输出数字（不转 string），时间固定格式
	want := `{"userId":9007199254740993,"userName":"admin","createTime":"2026-09-24 08:00:00"}`
	if string(b) != want {
		t.Errorf("序列化不符:\n want %s\n got  %s", want, b)
	}

	var out demo
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("Unmarshal 失败: %v", err)
	}
	if out != in {
		t.Errorf("往返不一致: %+v != %+v", out, in)
	}
}

// TestInt64PrimaryKey int64 主键 JSON 输出为数字（契约：与 Java 一致，不转 string）
func TestInt64PrimaryKey(t *testing.T) {
	type row struct {
		ID int64 `json:"deptId"`
	}
	b, _ := json.Marshal(row{ID: 100})
	if string(b) != `{"deptId":100}` {
		t.Errorf("int64 主键应输出数字: %s", b)
	}
}

// TestValueAndScan GORM 存取：driver.Valuer / sql.Scanner 往返
func TestValueAndScan(t *testing.T) {
	in := DateTime(mustParse(t, "2026-09-24 08:00:00"))

	v, err := in.Value()
	if err != nil {
		t.Fatalf("Value 失败: %v", err)
	}
	tt, ok := v.(time.Time)
	if !ok || tt != mustParse(t, "2026-09-24 08:00:00") {
		t.Errorf("Value 应还原为 time.Time: %v", v)
	}

	var out DateTime
	if err := out.Scan(tt); err != nil {
		t.Fatalf("Scan 失败: %v", err)
	}
	if out != in {
		t.Errorf("Scan 往返不一致: %v != %v", out, in)
	}
}

// TestValueScanNull 数据库 NULL 语义：零值写 NULL、NULL 读零值
func TestValueScanNull(t *testing.T) {
	var zero DateTime
	v, err := zero.Value()
	if err != nil || v != nil {
		t.Errorf("零值应写 NULL: v=%v err=%v", v, err)
	}

	var out DateTime
	if err := out.Scan(nil); err != nil || !out.IsZero() {
		t.Errorf("NULL 应读为零值: err=%v zero=%v", err, out.IsZero())
	}

	// 字符串值兼容（MySQL 驱动偶发返回 []byte/string）
	var s DateTime
	if err := s.Scan("2026-09-24 08:00:00"); err != nil || s.Time() != mustParse(t, "2026-09-24 08:00:00") {
		t.Errorf("字符串 Scan 失败: %v %v", s, err)
	}
	var b DateTime
	if err := b.Scan([]byte("2026-09-24 08:00:00")); err != nil || b.Time() != mustParse(t, "2026-09-24 08:00:00") {
		t.Errorf("[]byte Scan 失败: %v %v", b, err)
	}
}
