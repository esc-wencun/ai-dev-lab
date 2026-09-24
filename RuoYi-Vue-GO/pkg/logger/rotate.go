// Package logger 统一日志框架（对位 Java logback.xml / Python 版三文件滚动成果）。
//
// 三路文件（logs/ 目录，按天滚动、保留 MaxAgeDays 天，默认 60）：
//   - sys-info.log  : INFO 及以上（ERROR 也会进，与 Java/Python 多 appender 行为一致）
//   - sys-error.log : 仅 ERROR 及以上
//   - sys-user.log  : 登录日志命名 logger（对位 Python sys-user）
//
// 纪律：业务代码统一入口本包（logger.Info/Error/User），禁止各包自建 zap logger。
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// dailyRotator 按天滚动的追加写文件（zapcore.WriteSyncer 兼容）：
// 当前文件固定为 <dir>/<name>.log，跨天时先将其重命名为 <name>-<前一天>.log 再建新文件，
// 并清理超过 maxAgeDays 的 <name>-*.log 备份。
type dailyRotator struct {
	mu         sync.Mutex
	dir        string
	name       string
	maxAgeDays int
	nowFn      func() time.Time // 可注入，测试用

	day string
	f   *os.File
}

func newDailyRotator(dir, name string, maxAgeDays int) *dailyRotator {
	if maxAgeDays <= 0 {
		maxAgeDays = 60
	}
	return &dailyRotator{dir: dir, name: name, maxAgeDays: maxAgeDays, nowFn: time.Now}
}

// Write 追加写入；跨天自动滚动
func (w *dailyRotator) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	day := w.nowFn().Format("2006-01-02")
	if w.f == nil || day != w.day {
		if err := w.rotateLocked(day); err != nil {
			return 0, err
		}
	}
	return w.f.Write(p)
}

// Sync 刷盘
func (w *dailyRotator) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	return w.f.Sync()
}

// Close 关闭当前文件
func (w *dailyRotator) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}

func (w *dailyRotator) rotateLocked(day string) error {
	if w.f != nil {
		_ = w.f.Close()
		w.f = nil
		// 旧文件改名留档：<name>.log → <name>-<day>.log
		_ = os.Rename(filepath.Join(w.dir, w.name+".log"),
			filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.name, w.day)))
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(w.dir, w.name+".log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	w.f = f
	w.day = day
	w.cleanupLocked()
	return nil
}

// cleanupLocked 删除超过保留期的 <name>-*.log 备份
func (w *dailyRotator) cleanupLocked() {
	matches, err := filepath.Glob(filepath.Join(w.dir, w.name+"-*.log"))
	if err != nil {
		return
	}
	cutoff := w.nowFn().AddDate(0, 0, -w.maxAgeDays)
	for _, m := range matches {
		base := filepath.Base(m)
		dayStr := base[len(w.name)+1 : len(base)-len(".log")]
		day, err := time.ParseInLocation("2006-01-02", dayStr, time.Local)
		if err != nil {
			continue // 非日期后缀的文件不动
		}
		if day.Before(cutoff) {
			_ = os.Remove(m)
		}
	}
}
