// Package types 基础序列化类型（对位 Java 全局 Jackson 配置的 Go 落地）。
//
// 契约（与 Java/Python 版一致，前端日期显示不乱的前提）：
//   - DateTime 输出固定 `yyyy-MM-dd HH:mm:ss`（对位 Jackson 全局日期格式）；
//   - 零值输出 null（对位 Java Date 为 null 时 Jackson 默认行为）；
//   - int64 主键直接输出数字，不转 string。
//
// do/vo 的时间字段统一用 DateTime，禁止裸 time.Time 直接出现在响应体。
package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateTimeLayout 序列化格式（对位 Java yyyy-MM-dd HH:mm:ss）
const DateTimeLayout = "2006-01-02 15:04:05"

// DateLayout 日期格式（对位 Java yyyy-MM-dd，兼容前端只传日期的场景）
const DateLayout = "2006-01-02"

// DateTime 对位 Java Date/LocalDateTime 在 JSON 中的统一形态
type DateTime time.Time

// Now 当前时间
func Now() DateTime {
	return DateTime(time.Now())
}

// Time 还原为 time.Time
func (t DateTime) Time() time.Time {
	return time.Time(t)
}

// IsZero 零值判断
func (t DateTime) IsZero() bool {
	return time.Time(t).IsZero()
}

// MarshalJSON 输出 `yyyy-MM-dd HH:mm:ss`；零值输出 null
func (t DateTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + tt.Format(DateTimeLayout) + `"`), nil
}

// UnmarshalJSON 接受 `yyyy-MM-dd HH:mm:ss` / `yyyy-MM-dd`；空串与 null 归零
func (t *DateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*t = DateTime(time.Time{})
		return nil
	}
	parsed, err := parseDateTime(s)
	if err != nil {
		return err
	}
	*t = DateTime(parsed)
	return nil
}

// parseDateTime 依次尝试 完整格式 / 日期格式 / RFC3339
func parseDateTime(s string) (time.Time, error) {
	for _, layout := range []string{DateTimeLayout, DateLayout, time.RFC3339} {
		if tt, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return tt, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间 %q（期望 %s 或 %s）", s, DateTimeLayout, DateLayout)
}

// Value 实现 driver.Valuer（GORM 写库：零值写 NULL）
func (t DateTime) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return nil, nil
	}
	return tt, nil
}

// Scan 实现 sql.Scanner（GORM 读库）
func (t *DateTime) Scan(value any) error {
	if value == nil {
		*t = DateTime(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = DateTime(v)
		return nil
	case string:
		parsed, err := parseDateTime(v)
		if err != nil {
			return err
		}
		*t = DateTime(parsed)
		return nil
	case []byte:
		parsed, err := parseDateTime(string(v))
		if err != nil {
			return err
		}
		*t = DateTime(parsed)
		return nil
	default:
		return fmt.Errorf("DateTime.Scan 不支持类型 %T", value)
	}
}
