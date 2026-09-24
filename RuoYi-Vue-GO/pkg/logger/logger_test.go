package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", name, err)
	}
	return string(b)
}

// TestThreeFilesLogging 三路日志落文件断言：
// INFO 只进 sys-info.log；ERROR 同时进 sys-info.log 与 sys-error.log；
// 登录日志只进 sys-user.log。
func TestThreeFilesLogging(t *testing.T) {
	dir := t.TempDir()
	l, err := Init(Options{Dir: dir})
	if err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	defer l.Close()

	Info("hello-info", zap.String("k", "v"))
	Error("hello-error")
	User("admin 登录成功", zap.String("userName", "admin"))
	l.Sync()

	infoLog := readFile(t, dir, "sys-info.log")
	errorLog := readFile(t, dir, "sys-error.log")
	userLog := readFile(t, dir, "sys-user.log")

	if !strings.Contains(infoLog, "hello-info") {
		t.Errorf("sys-info.log 应含 INFO 日志: %q", infoLog)
	}
	if !strings.Contains(infoLog, "hello-error") {
		t.Errorf("ERROR 应同时写入 sys-info.log（对位 Java 多 appender）: %q", infoLog)
	}
	if strings.Contains(errorLog, "hello-info") {
		t.Errorf("sys-error.log 不应含 INFO 日志: %q", errorLog)
	}
	if !strings.Contains(errorLog, "hello-error") {
		t.Errorf("sys-error.log 应含 ERROR 日志: %q", errorLog)
	}
	if !strings.Contains(userLog, "admin 登录成功") {
		t.Errorf("sys-user.log 应含登录日志: %q", userLog)
	}
	if strings.Contains(userLog, "hello-info") || strings.Contains(userLog, "hello-error") {
		t.Errorf("sys-user.log 不应混入系统日志: %q", userLog)
	}
}

// TestDailyRotationAndCleanup 按天滚动与过期清理（nowFn 注入模拟跨天）
func TestDailyRotationAndCleanup(t *testing.T) {
	dir := t.TempDir()
	day1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.Local)
	rot := newDailyRotator(dir, "sys-info", 60)
	rot.nowFn = func() time.Time { return day1 }

	if _, err := rot.Write([]byte("day1-line\n")); err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	// 跨天：旧文件应改名为备份，新 <name>.log 只含新一天内容
	day2 := day1.AddDate(0, 0, 1)
	rot.nowFn = func() time.Time { return day2 }
	if _, err := rot.Write([]byte("day2-line\n")); err != nil {
		t.Fatalf("跨天写入失败: %v", err)
	}
	backup := readFile(t, dir, "sys-info-2026-01-01.log")
	if !strings.Contains(backup, "day1-line") {
		t.Errorf("滚动备份应保留 day1 内容: %q", backup)
	}
	cur := readFile(t, dir, "sys-info.log")
	if strings.Contains(cur, "day1-line") || !strings.Contains(cur, "day2-line") {
		t.Errorf("当前文件应只含 day2 内容: %q", cur)
	}

	// 超过保留期（60 天）的备份应被清理
	day100 := day1.AddDate(0, 0, 100)
	rot.nowFn = func() time.Time { return day100 }
	if _, err := rot.Write([]byte("day100-line\n")); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sys-info-2026-01-01.log")); !os.IsNotExist(err) {
		t.Errorf("超过 60 天的备份应被清理")
	}
	if err := rot.Close(); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}
}

// TestGetPanicsBeforeInit 未初始化就取 logger 属编程错误，应 panic
func TestGetPanicsBeforeInit(t *testing.T) {
	saved := std
	defer func() { std = saved }()

	std = nil
	defer func() {
		if rec := recover(); rec == nil {
			t.Error("未初始化调用 Get 应 panic")
		}
	}()
	Get()
}
