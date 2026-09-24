// 定时任务调度器单元测试：invokeTarget 解析（类型推断）与注册表。
package scheduler

import (
	"testing"
)

func TestParseTarget(t *testing.T) {
	cases := []struct {
		target string
		bean   string
		method string
		params []any
	}{
		{"ryTask.ryNoParams", "ryTask", "ryNoParams", nil},
		{"ryTask.ryParams('ry')", "ryTask", "ryParams", []any{"ry"}},
		{"ryTask.ryParams(\"x,y\")", "ryTask", "ryParams", []any{"x,y"}},
		{"ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)", "ryTask", "ryMultipleParams",
			[]any{"ry", true, int64(2000), 316.50, int64(100)}},
	}
	for _, tc := range cases {
		bean, method, params, err := ParseTarget(tc.target)
		if err != nil {
			t.Errorf("%s 解析失败: %v", tc.target, err)
			continue
		}
		if bean != tc.bean || method != tc.method {
			t.Errorf("%s: bean/method = %s.%s, want %s.%s", tc.target, bean, method, tc.bean, tc.method)
		}
		if len(params) != len(tc.params) {
			t.Errorf("%s: 参数个数 = %d, want %d", tc.target, len(params), len(tc.params))
			continue
		}
		for i := range params {
			if params[i] != tc.params[i] {
				t.Errorf("%s: params[%d] = %v(%T), want %v(%T)", tc.target, i, params[i], params[i], tc.params[i], tc.params[i])
			}
		}
	}
	// 非法形态
	if _, _, _, err := ParseTarget("illegal"); err == nil {
		t.Error("非法目标应报错")
	}
}

func TestRegistry(t *testing.T) {
	// 预置任务已在 init 注册
	if !registered("ryTask.ryNoParams") {
		t.Error("ryTask.ryNoParams 应已注册")
	}
	fn, ok := lookup("ryTask", "ryParams")
	if !ok || fn == nil {
		t.Error("ryTask.ryParams 应可查到")
	}
	if _, ok := lookup("ryTask", "notExists"); ok {
		t.Error("未注册目标不应查到")
	}
}

func registered(name string) bool {
	for _, n := range Registered() {
		if n == name {
			return true
		}
	}
	return false
}
