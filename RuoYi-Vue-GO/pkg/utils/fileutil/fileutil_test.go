// fileutil 单元测试：白名单、穿越防护、落盘。
package fileutil

import (
	"strings"
	"testing"
)

func TestAllowed(t *testing.T) {
	if !Allowed("png", ImageExtension) || !Allowed(".PNG", ImageExtension) {
		t.Error("png 应在图片白名单")
	}
	if Allowed("exe", DefaultExtension) {
		t.Error("exe 应被拒绝")
	}
	if Allowed("", DefaultExtension) {
		t.Error("空扩展名应拒绝")
	}
}

func TestCheckAllowDownload(t *testing.T) {
	if CheckAllowDownload("../secret.txt") {
		t.Error(".. 穿越应拒绝")
	}
	if !CheckAllowDownload("report_20260924.pdf") {
		t.Error("pdf 应允许")
	}
	if CheckAllowDownload("virus.exe") {
		t.Error("exe 应拒绝")
	}
	if CheckAllowDownload("") {
		t.Error("空名应拒绝")
	}
}

func TestStripPrefixAndResolve(t *testing.T) {
	if got := StripPrefix("/profile/avatar/2026/09/24/a.png"); got != "/avatar/2026/09/24/a.png" {
		t.Errorf("StripPrefix = %q", got)
	}
	ProfileRoot = t.TempDir()
	local, err := ResolveLocal("/profile/avatar/2026/09/24/a.png")
	if err != nil {
		t.Fatalf("ResolveLocal 失败: %v", err)
	}
	if !strings.HasSuffix(strings.ReplaceAll(local, "\\", "/"), "avatar/2026/09/24/a.png") {
		t.Errorf("本地路径 = %q", local)
	}
	if _, err := ResolveLocal("/profile/../../etc/passwd"); err == nil {
		t.Error("穿越应报错")
	}
}

func TestUpload(t *testing.T) {
	ProfileRoot = t.TempDir()
	dir := UploadPath()
	url, err := Upload(dir, "测试 文档.docx", 1024, strings.NewReader("hello doc"), DefaultExtension)
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	if !strings.HasPrefix(url, "/profile/upload/2026/") {
		t.Errorf("URL 形态 = %q", url)
	}
	if !strings.Contains(url, "测试 文档_") || !strings.HasSuffix(url, ".docx") {
		t.Errorf("文件名规则 = %q", url)
	}
	// exe 拒绝
	if _, err := Upload(dir, "bad.exe", 10, strings.NewReader("x"), DefaultExtension); err == nil {
		t.Error("exe 应拒绝")
	}
	// 超长文件名
	if _, err := Upload(dir, strings.Repeat("a", 120)+".png", 10, strings.NewReader("x"), ImageExtension); err == nil {
		t.Error("超长文件名应拒绝")
	}
}
