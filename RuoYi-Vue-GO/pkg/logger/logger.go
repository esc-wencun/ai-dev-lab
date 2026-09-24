package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Options 日志初始化参数
type Options struct {
	Dir        string // 日志目录（如 logs/）
	MaxAgeDays int    // 备份保留天数，<=0 时默认 60
}

// Logger 聚合三路命名 logger 的句柄
type Logger struct {
	base     *zap.Logger // sys-info.log + sys-error.log
	user     *zap.Logger // sys-user.log（登录日志）
	rotators []*dailyRotator
	syncers  []*zapcore.BufferedWriteSyncer
}

var std *Logger

// Init 初始化全局日志（进程生命周期内调用一次）；返回的 Logger 需在退出前 Sync()
func Init(opts Options) (*Logger, error) {
	maxAge := opts.MaxAgeDays
	if maxAge <= 0 {
		maxAge = 60
	}

	encCfg := zapcore.EncoderConfig{
		LevelKey:    "level",
		MessageKey:  "msg",
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime:  zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
	}
	encoder := zapcore.NewConsoleEncoder(encCfg)

	build := func(name string, minLevel zapcore.Level) (*zapcore.BufferedWriteSyncer, *dailyRotator) {
		rot := newDailyRotator(opts.Dir, name, maxAge)
		return &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(rot),
			Size:          256,
			FlushInterval: time.Second,
		}, rot
	}

	infoSyncer, infoRot := build("sys-info", zapcore.InfoLevel)
	errorSyncer, errorRot := build("sys-error", zapcore.ErrorLevel)
	userSyncer, userRot := build("sys-user", zapcore.InfoLevel)

	// sys-info.log 收 INFO 及以上（含 ERROR），sys-error.log 仅 ERROR 及以上——
	// 对位 Java/Python 同一 logger 挂两个不同级别 appender 的行为
	infoCore := zapcore.NewCore(encoder, infoSyncer, zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.InfoLevel
	}))
	errorCore := zapcore.NewCore(encoder, errorSyncer, zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.ErrorLevel
	}))
	userCore := zapcore.NewCore(encoder, userSyncer, zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.InfoLevel
	}))

	l := &Logger{
		base:     zap.New(zapcore.NewTee(infoCore, errorCore)),
		user:     zap.New(userCore),
		rotators: []*dailyRotator{infoRot, errorRot, userRot},
		syncers:  []*zapcore.BufferedWriteSyncer{infoSyncer, errorSyncer, userSyncer},
	}
	std = l
	return l, nil
}

// Get 取全局 logger（Init 之前调用会 panic——编程错误，测试期暴露）
func Get() *Logger {
	if std == nil {
		panic("logger: 未初始化，请先调用 logger.Init")
	}
	return std
}

// Info 系统信息日志 → sys-info.log
func Info(msg string, fields ...zap.Field) { Get().base.Info(msg, fields...) }

// Error 系统错误日志 → sys-error.log 与 sys-info.log
func Error(msg string, fields ...zap.Field) { Get().base.Error(msg, fields...) }

// User 登录日志 → sys-user.log（对位 Python sys-user 命名 logger）
func User(msg string, fields ...zap.Field) { Get().user.Info(msg, fields...) }

// Sync 刷盘（进程退出前必须调用，异步缓冲才会落盘）
func (l *Logger) Sync() {
	_ = l.base.Sync()
	_ = l.user.Sync()
}

// Close 刷盘、停止异步缓冲并关闭文件句柄（进程退出或测试收尾时调用）
func (l *Logger) Close() {
	l.Sync()
	for _, s := range l.syncers {
		s.Stop()
	}
	for _, r := range l.rotators {
		_ = r.Close()
	}
}
