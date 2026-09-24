package types

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestJSONTagsAreCamelCase lint 卡点（对位 Jackson 全局驼峰序列化的 Go 约束）：
// 全仓库 struct 字段凡带 json tag，tag 名必须显式驼峰（首字母小写、无下划线），
// 禁止 snake_case，禁止依赖默认字段名输出（`json:",omitempty"` 这类空名同样不允许）。
// 评审纪律（vo 必须显式写 tag）由 AGENTS.md 保证，本测试卡死"写了 tag 但不是驼峰"的口子。
func TestJSONTagsAreCamelCase(t *testing.T) {
	repoRoot := "../.."
	var violations []string

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			// 跳过隐藏目录与 vendor
			if strings.HasPrefix(name, ".") || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("解析 %s 失败: %w", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			st, ok := n.(*ast.StructType)
			if !ok || st.Fields == nil {
				return true
			}
			for _, f := range st.Fields.List {
				if f.Tag == nil {
					continue
				}
				tag, err := strconvUnquote(f.Tag.Value)
				if err != nil {
					continue
				}
				jsonName, hasJSON := jsonTagName(tag)
				if !hasJSON {
					continue
				}
				line := fset.Position(f.Pos()).Line
				if !isCamelCase(jsonName) {
					violations = append(violations,
						fmt.Sprintf("%s:%d json tag 非驼峰/依赖默认字段名: %q", path, line, jsonName))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("扫描仓库失败: %v", err)
	}

	for _, v := range violations {
		t.Error(v)
	}
}

// jsonTagName 提取 `json:"xxx"` 中的 xxx；无 json tag 返回 false
func jsonTagName(tag string) (string, bool) {
	for _, part := range strings.Split(tag, " ") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "json:") {
			val := strings.Trim(strings.TrimPrefix(part, "json:"), `"'`)
			return strings.Split(val, ",")[0], true
		}
	}
	return "", false
}

var camelCasePattern = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)

// isCamelCase "-"（忽略字段）合法；空串（依赖默认字段名）不合法
func isCamelCase(name string) bool {
	if name == "-" {
		return true
	}
	return camelCasePattern.MatchString(name)
}

func strconvUnquote(quoted string) (string, error) {
	return unquoteSimple(quoted), nil
}

// unquoteSimple 去掉 ast.Tag.Value 的反引号（tag 常量已是字面量，无需完整 Unquote）
func unquoteSimple(quoted string) string {
	return strings.Trim(quoted, "`")
}
