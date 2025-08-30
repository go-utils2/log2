package log2

import (
	"context"
	"fmt"
	"time"

	_ "github.com/zeromicro/go-zero/core/logc"
	"go.uber.org/zap"
)

// GoZeroLogger go-zero日志适配器
// 实现go-zero的logx.Logger接口，将go-zero的日志调用适配到当前的Logger接口
type GoZeroLogger struct {
	logger Logger // 底层日志器
	skip   int    // 调用栈跳过层数
}

// NewGoZeroLogger 创建一个新的go-zero日志适配器
// 参数:
//   - logger: 底层日志器实例
//   - skip: 调用栈跳过层数，默认为1
//
// 返回值:
//   - *GoZeroLogger: go-zero日志适配器实例
func NewGoZeroLogger(logger Logger, skip ...int) *GoZeroLogger {
	skipLevel := 1
	if len(skip) > 0 {
		skipLevel = skip[0]
	}

	return &GoZeroLogger{
		logger: logger.AddCallerSkip(skipLevel),
		skip:   skipLevel,
	}
}

// Debug 输出Debug级别日志
func (g *GoZeroLogger) Debug(v ...interface{}) {
	if len(v) > 0 {
		g.logger.Debug(formatMessage(v...))
	}
}

// Debugf 输出格式化的Debug级别日志
func (g *GoZeroLogger) Debugf(format string, v ...interface{}) {
	g.logger.Debug(formatMessagef(format, v...))
}

// Debugv 输出带字段的Debug级别日志
func (g *GoZeroLogger) Debugv(v interface{}) {
	g.logger.Debug("debug", zap.Any("data", v))
}

// Debugw 输出带键值对的Debug级别日志
func (g *GoZeroLogger) Debugw(msg string, keysAndValues ...interface{}) {
	fields := convertKeysAndValues(keysAndValues...)
	g.logger.Debug(msg, fields...)
}

// Error 输出Error级别日志
func (g *GoZeroLogger) Error(v ...interface{}) {
	if len(v) > 0 {
		g.logger.Error(formatMessage(v...))
	}
}

// Errorf 输出格式化的Error级别日志
func (g *GoZeroLogger) Errorf(format string, v ...interface{}) {
	g.logger.Error(formatMessagef(format, v...))
}

// Errorv 输出带字段的Error级别日志
func (g *GoZeroLogger) Errorv(v interface{}) {
	g.logger.Error("error", zap.Any("data", v))
}

// Errorw 输出带键值对的Error级别日志
func (g *GoZeroLogger) Errorw(msg string, keysAndValues ...interface{}) {
	fields := convertKeysAndValues(keysAndValues...)
	g.logger.Error(msg, fields...)
}

// Info 输出Info级别日志
func (g *GoZeroLogger) Info(v ...interface{}) {
	if len(v) > 0 {
		g.logger.Info(formatMessage(v...))
	}
}

// Infof 输出格式化的Info级别日志
func (g *GoZeroLogger) Infof(format string, v ...interface{}) {
	g.logger.Info(formatMessagef(format, v...))
}

// Infov 输出带字段的Info级别日志
func (g *GoZeroLogger) Infov(v interface{}) {
	g.logger.Info("info", zap.Any("data", v))
}

// Infow 输出带键值对的Info级别日志
func (g *GoZeroLogger) Infow(msg string, keysAndValues ...interface{}) {
	fields := convertKeysAndValues(keysAndValues...)
	g.logger.Info(msg, fields...)
}

// Slow 输出慢查询日志（使用Warn级别）
func (g *GoZeroLogger) Slow(v ...interface{}) {
	if len(v) > 0 {
		g.logger.Warn(formatMessage(v...), zap.String("type", "slow"))
	}
}

// Slowf 输出格式化的慢查询日志
func (g *GoZeroLogger) Slowf(format string, v ...interface{}) {
	g.logger.Warn(formatMessagef(format, v...), zap.String("type", "slow"))
}

// Slowv 输出带字段的慢查询日志
func (g *GoZeroLogger) Slowv(v interface{}) {
	g.logger.Warn("slow", zap.Any("data", v), zap.String("type", "slow"))
}

// Sloww 输出带键值对的慢查询日志
func (g *GoZeroLogger) Sloww(msg string, keysAndValues ...interface{}) {
	fields := convertKeysAndValues(keysAndValues...)
	fields = append(fields, zap.String("type", "slow"))
	g.logger.Warn(msg, fields...)
}

// WithCallerSkip 返回一个新的日志器，跳过指定层数的调用栈
func (g *GoZeroLogger) WithCallerSkip(skip int) *GoZeroLogger {
	return &GoZeroLogger{
		logger: g.logger.AddCallerSkip(skip),
		skip:   g.skip + skip,
	}
}

// WithContext 返回一个带上下文的日志器（当前实现忽略context）
func (g *GoZeroLogger) WithContext(ctx context.Context) *GoZeroLogger {
	// 当前实现不处理context，直接返回自身
	// 如果需要支持context相关功能，可以在这里扩展
	return g
}

// WithDuration 返回一个带持续时间字段的日志器
func (g *GoZeroLogger) WithDuration(duration time.Duration) *GoZeroLogger {
	return &GoZeroLogger{
		logger: g.logger.With(zap.Duration("duration", duration)),
		skip:   g.skip,
	}
}

// WithFields 返回一个带字段的日志器
func (g *GoZeroLogger) WithFields(fields ...zap.Field) *GoZeroLogger {
	return &GoZeroLogger{
		logger: g.logger.With(fields...),
		skip:   g.skip,
	}
}

// formatMessage 格式化消息
func formatMessage(v ...interface{}) string {
	if len(v) == 0 {
		return ""
	}
	if len(v) == 1 {
		if str, ok := v[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(v...)
}

// formatMessagef 格式化消息（带格式）
func formatMessagef(format string, v ...interface{}) string {
	if len(v) == 0 {
		return format
	}
	return fmt.Sprintf(format, v...)
}

// convertKeysAndValues 将键值对转换为zap字段
func convertKeysAndValues(keysAndValues ...interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			// 奇数个参数，最后一个作为值，键为"extra"
			fields = append(fields, zap.Any("extra", keysAndValues[i]))
			break
		}

		key, ok := keysAndValues[i].(string)
		if !ok {
			// 键不是字符串，转换为字符串
			key = fmt.Sprintf("%v", keysAndValues[i])
		}

		value := keysAndValues[i+1]
		fields = append(fields, zap.Any(key, value))
	}

	return fields
}
