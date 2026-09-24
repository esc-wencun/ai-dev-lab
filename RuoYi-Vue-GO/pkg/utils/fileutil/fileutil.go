// Package fileutil 文件上传下载（对位 Java FileUploadUtils + FileUtils + MimeTypeUtils + RuoYiConfig 路径约定）。
//
// 路径约定（对位 RuoYiConfig）：
//   - profile 根目录（Java 默认 D:/ruoyi/uploadPath，Go 配置化，默认 ./uploads）
//   - 上传 {profile}/upload/{yyyy/MM/dd}/、头像 {profile}/avatar/{yyyy/MM/dd}/、下载 {profile}/download/
//   - URL 前缀 /profile（对位 Constants.RESOURCE_PREFIX，由静态文件服务挂载）
package fileutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 允许的扩展名白名单（对位 MimeTypeUtils）。
var (
	ImageExtension = []string{"bmp", "gif", "jpg", "jpeg", "png"}
	// DefaultExtension 对位 DEFAULT_ALLOWED_EXTENSION（无 exe 等）
	DefaultExtension = []string{
		"bmp", "gif", "jpg", "jpeg", "png",
		"doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt",
		"rar", "zip", "gz", "bz2",
		"mp4", "avi", "rmvb",
		"pdf",
	}
)

// 常量对位。
const (
	MaxSize        = 50 * 1024 * 1024 // DEFAULT_MAX_SIZE 50MB
	FileNameLength = 100              // DEFAULT_FILE_NAME_LENGTH
	ResourcePrefix = "/profile"       // 对位 Constants.RESOURCE_PREFIX
)

// ProfileRoot 根目录（main 装配时注入；默认 ./uploads）。
var ProfileRoot = "./uploads"

// ProfilePath profile 根目录绝对/相对路径。
func ProfilePath() string { return ProfileRoot }

// UploadPath {profile}/upload。
func UploadPath() string { return ProfileRoot + "/upload" }

// AvatarPath {profile}/avatar。
func AvatarPath() string { return ProfileRoot + "/avatar" }

// DownloadPath {profile}/download/。
func DownloadPath() string { return ProfileRoot + "/download/" }

// Allowed 判断扩展名（小写比较）在白名单内。
func Allowed(ext string, whitelist []string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	for _, w := range whitelist {
		if ext == w {
			return true
		}
	}
	return false
}

// Upload 保存上传文件并返回 /profile 相对 URL（对位 FileUploadUtils.upload）：
//   - 文件名长度 ≤100、大小 ≤50MB、扩展名在白名单内；
//   - 落盘 {baseDir}/{yyyy/MM/dd}/{原基名}_{seq}.{ext}（对位 extractFilename）；
//   - 返回 "/profile/upload/2026/09/24/xx_ab12.docx" 形态。
func Upload(baseDir string, fileName string, size int64, src io.Reader, whitelist []string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("上传文件名为空")
	}
	if len(fileName) > FileNameLength {
		return "", fmt.Errorf("上传的文件名最长%d个字符", FileNameLength)
	}
	if size > MaxSize {
		return "", fmt.Errorf("上传的文件大小超出限制的文件大小！允许的文件最大大小是：%dMB！", MaxSize/1024/1024)
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	if !Allowed(ext, whitelist) {
		return "", fmt.Errorf("文件格式不支持上传%s类型文件", ext)
	}
	// 原基名（去路径穿越成分）
	base := filepath.Base(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	base = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, base)
	// {yyyy/MM/dd}/{基名}_{seq}{ext}
	datePath := time.Now().Format("2006/01/02")
	seq := newSeq()
	rel := filepath.Join(datePath, fmt.Sprintf("%s_%s%s", base, seq, ext))
	abs := filepath.Join(baseDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	dst, err := os.Create(abs)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	// URL = /profile + {baseDir 相对 ProfileRoot 的部分}/{rel}（对位 getPathFileName）
	relToRoot, err := filepath.Rel(ProfileRoot, baseDir)
	if err != nil {
		relToRoot = "."
	}
	urlPath := "/" + filepath.ToSlash(filepath.Join(relToRoot, rel))
	return ResourcePrefix + urlPath, nil
}

// CheckAllowDownload 下载校验（对位 checkAllowDownload：拒绝 .. 穿越与白名单外）。
func CheckAllowDownload(name string) bool {
	if name == "" || strings.Contains(name, "..") {
		return false
	}
	return Allowed(filepath.Ext(name), DefaultExtension)
}

// StripPrefix 移除 /profile 请求前缀（对位 FileUtils.stripPrefix：substringAfter）。
func StripPrefix(resource string) string {
	idx := strings.Index(resource, ResourcePrefix)
	if idx < 0 {
		return resource
	}
	return resource[idx+len(ResourcePrefix):]
}

// ResolveLocal 解析 /profile URL 为本地绝对路径（含穿越防护）。
func ResolveLocal(resource string) (string, error) {
	rel := StripPrefix(resource)
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("路径非法")
	}
	return filepath.Join(ProfileRoot, filepath.FromSlash(rel)), nil
}

// newSeq 4 位 hex 序列（对位 Seq.getId 上传序列）。
func newSeq() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
