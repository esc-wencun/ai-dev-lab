// handler 包通用辅助：下载响应头 / 服务器 URL / 时间戳。
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// serverURL 请求侧完整地址（对位 ServerConfig.getUrl：scheme://host:port）。
func serverURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

// writeAttachment 下载响应（对位 FileUtils.setAttachmentResponseHeader + writeBytes）。
func writeAttachment(c *gin.Context, data []byte, filename string) {
	enc := url.PathEscape(filename)
	enc = strings.ReplaceAll(enc, "+", "%20")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+enc+";filename*=utf-8''"+enc)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition,download-filename")
	c.Header("download-filename", enc)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// timeNowMillis 当前毫秒时间戳（下载文件名前缀，对位 System.currentTimeMillis()）。
func timeNowMillis() int64 { return time.Now().UnixMilli() }

// fmtInt int64 转字符串。
func fmtInt(v int64) string { return fmt.Sprintf("%d", v) }

// mustJSON 序列化为 JSON 字节（失败返回 "{}"，调用方都是可控结构）。
func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}
