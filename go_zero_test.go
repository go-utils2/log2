package log2

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestGoZeroLogger 测试go-zero日志适配器
func TestGoZeroLogger(t *testing.T) {
	// 创建基础日志器
	config := NewConfig()
	config.Debug = true
	config.HideConsole = false
	config.Service = "test-gozero"

	baseLogger, err := config.Build()
	if err != nil {
		t.Fatalf("创建基础日志器失败: %v", err)
	}

	// 创建go-zero日志适配器
	gzLogger := NewGoZeroLogger(baseLogger)

	// 测试基本日志方法
	t.Run("基本日志方法", func(t *testing.T) {
		gzLogger.Debug("这是一条debug消息")
		gzLogger.Info("这是一条info消息")
		gzLogger.Error("这是一条error消息")
		gzLogger.Slow("这是一条慢查询消息")
	})

	// 测试格式化日志方法
	t.Run("格式化日志方法", func(t *testing.T) {
		gzLogger.Debugf("这是一条格式化debug消息: %s", "测试")
		gzLogger.Infof("这是一条格式化info消息: %d", 123)
		gzLogger.Errorf("这是一条格式化error消息: %v", map[string]string{"key": "value"})
		gzLogger.Slowf("这是一条格式化慢查询消息: %f", 1.23)
	})

	// 测试结构化日志方法
	t.Run("结构化日志方法", func(t *testing.T) {
		gzLogger.Debugv(map[string]interface{}{"debug": "data"})
		gzLogger.Infov([]string{"info", "array"})
		gzLogger.Errorv(struct{ Name string }{Name: "error struct"})
		gzLogger.Slowv(42)
	})

	// 测试键值对日志方法
	t.Run("键值对日志方法", func(t *testing.T) {
		gzLogger.Debugw("debug消息", "key1", "value1", "key2", 123)
		gzLogger.Infow("info消息", "user", "admin", "action", "login")
		gzLogger.Errorw("error消息", "error", "connection failed", "retry", 3)
		gzLogger.Sloww("slow消息", "query", "SELECT * FROM users", "duration", "2.5s")
	})

	// 测试WithCallerSkip
	t.Run("WithCallerSkip", func(t *testing.T) {
		skippedLogger := gzLogger.WithCallerSkip(1)
		skippedLogger.Info("跳过调用栈的消息")
	})

	// 测试WithContext
	t.Run("WithContext", func(t *testing.T) {
		ctx := context.Background()
		ctxLogger := gzLogger.WithContext(ctx)
		ctxLogger.Info("带上下文的消息")
	})

	// 测试WithDuration
	t.Run("WithDuration", func(t *testing.T) {
		durationLogger := gzLogger.WithDuration(time.Second * 2)
		durationLogger.Info("带持续时间的消息")
	})

	// 测试WithFields
	t.Run("WithFields", func(t *testing.T) {
		fieldsLogger := gzLogger.WithFields(
			zap.String("module", "test"),
			zap.Int("version", 1),
			zap.Bool("enabled", true),
		)
		fieldsLogger.Info("带字段的消息")
	})

	// 测试边界情况
	t.Run("边界情况", func(t *testing.T) {
		// 空消息
		gzLogger.Debug()
		gzLogger.Info()
		gzLogger.Error()
		gzLogger.Slow()

		// 空格式化
		gzLogger.Debugf("")
		gzLogger.Infof("只有格式字符串")

		// 奇数个键值对
		gzLogger.Infow("奇数键值对", "key1", "value1", "key2")

		// 非字符串键
		gzLogger.Infow("非字符串键", 123, "value1", true, "value2")
	})
}

// TestGoZeroLoggerIntegration 测试go-zero日志适配器与现有日志器的集成
func TestGoZeroLoggerIntegration(t *testing.T) {
	// 创建带文件输出的日志器
	config := NewConfig()
	config.Debug = true
	config.HideConsole = false
	config.FilePath = "/tmp/gozero_test"
	config.Service = "integration-test"

	baseLogger, err := config.Build()
	if err != nil {
		t.Fatalf("创建基础日志器失败: %v", err)
	}

	// 创建派生日志器
	derivedLogger := baseLogger.Derive("gozero-module")

	// 创建go-zero适配器
	gzLogger := NewGoZeroLogger(derivedLogger)

	// 测试日志输出
	gzLogger.Infow("集成测试消息",
		"component", "gozero-adapter",
		"test", "integration",
		"timestamp", time.Now().Unix(),
	)

	// 测试链式调用
	gzLogger.WithFields(
		zap.String("chain", "test"),
	).WithDuration(time.Millisecond * 100).Info("链式调用测试")
}

// BenchmarkGoZeroLogger 性能测试
func BenchmarkGoZeroLogger(b *testing.B) {
	config := NewConfig()
	config.Debug = false
	config.HideConsole = true // 隐藏控制台输出以提高性能测试准确性
	config.Service = "benchmark"

	baseLogger, err := config.Build()
	if err != nil {
		b.Fatalf("创建基础日志器失败: %v", err)
	}

	gzLogger := NewGoZeroLogger(baseLogger)

	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gzLogger.Info("benchmark info message")
		}
	})

	b.Run("Infof", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gzLogger.Infof("benchmark info message %d", i)
		}
	})

	b.Run("Infow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gzLogger.Infow("benchmark info message", "iteration", i, "type", "benchmark")
		}
	})

	b.Run("WithFields", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gzLogger.WithFields(
				zap.Int("iteration", i),
				zap.String("type", "benchmark"),
			).Info("benchmark with fields")
		}
	})
}