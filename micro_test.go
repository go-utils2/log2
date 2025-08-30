package log2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	microlog "go-micro.dev/v5/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestNewMicroLogger 测试NewMicroLogger函数
func TestNewMicroLogger(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)

	log := NewMicroLogger(targetLogger)

	require.NotNil(t, log)
	require.Equal(t, targetLogger, log.Logger)
	require.NotNil(t, log.options)
	require.Equal(t, "zap-micro", log.String())
}

// TestMicroLogger_Init 测试microLogger的Init方法
func TestMicroLogger_Init(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	t.Run("基本Init测试", func(t *testing.T) {
		err := microLog.Init()
		require.NoError(t, err)
	})

	t.Run("带选项的Init测试", func(t *testing.T) {
		option := func(o *microlog.Options) {
			o.CallerSkipCount = 2
		}

		err := microLog.Init(option)
		require.NoError(t, err)
		require.Equal(t, 2, microLog.options.CallerSkipCount)
	})

	t.Run("多个选项的Init测试", func(t *testing.T) {
		option1 := func(o *microlog.Options) {
			o.CallerSkipCount = 3
		}
		option2 := func(o *microlog.Options) {
			o.Level = microlog.DebugLevel
		}

		err := microLog.Init(option1, option2)
		require.NoError(t, err)
		require.Equal(t, 3, microLog.options.CallerSkipCount)
		require.Equal(t, microlog.DebugLevel, microLog.options.Level)
	})
}

// TestMicroLogger_Options 测试microLogger的Options方法
func TestMicroLogger_Options(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	t.Run("默认选项测试", func(t *testing.T) {
		options := microLog.Options()
		require.Equal(t, 0, options.CallerSkipCount)
		require.Equal(t, microlog.Level(0), options.Level)
	})

	t.Run("修改后的选项测试", func(t *testing.T) {
		microLog.options.CallerSkipCount = 5
		microLog.options.Level = microlog.WarnLevel

		options := microLog.Options()
		require.Equal(t, 5, options.CallerSkipCount)
		require.Equal(t, microlog.WarnLevel, options.Level)
	})
}

// TestMicroLogger_Fields 测试microLogger的Fields方法
func TestMicroLogger_Fields(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	t.Run("基本Fields测试", func(t *testing.T) {
		fields := map[string]interface{}{
			"user_id": 123,
			"action":  "login",
		}

		newLogger := microLog.Fields(fields)
		require.NotNil(t, newLogger)
		require.NotEqual(t, microLog, newLogger)

		// 测试新logger是否包含字段
		newLogger.Log(microlog.InfoLevel, "测试消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message) // Log方法使用空消息

		// 检查字段是否存在
		foundUserID := false
		foundAction := false
		for _, field := range logs[0].Context {
			if field.Key == "user_id" {
				foundUserID = true
				require.Equal(t, int64(123), field.Integer)
			}
			if field.Key == "action" {
				foundAction = true
				require.Equal(t, "login", field.String)
			}
		}
		require.True(t, foundUserID, "应该包含user_id字段")
		require.True(t, foundAction, "应该包含action字段")
	})

	t.Run("空Fields测试", func(t *testing.T) {
		recorded.TakeAll()
		fields := map[string]interface{}{}

		newLogger := microLog.Fields(fields)
		require.NotNil(t, newLogger)

		newLogger.Log(microlog.InfoLevel, "空字段测试")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message) // Log方法使用空消息
	})

	t.Run("CallerSkipCount测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.options.CallerSkipCount = 1

		fields := map[string]interface{}{"test": "value"}
		newLogger := microLog.Fields(fields)

		require.NotNil(t, newLogger)
	})
}

// TestMicroLogger_Log 测试microLogger的Log方法
func TestMicroLogger_Log(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	t.Run("InfoLevel测试", func(t *testing.T) {
		microLog.Log(microlog.InfoLevel, "info消息", 123)

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)
		require.Equal(t, "", logs[0].Message)
	})

	t.Run("DebugLevel测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Log(microlog.DebugLevel, "debug消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.DebugLevel, logs[0].Level)
	})

	t.Run("TraceLevel测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Log(microlog.TraceLevel, "trace消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.DebugLevel, logs[0].Level) // TraceLevel映射到DebugLevel
	})

	t.Run("WarnLevel测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Log(microlog.WarnLevel, "warn消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.WarnLevel, logs[0].Level)
	})

	t.Run("ErrorLevel测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Log(microlog.ErrorLevel, "error消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
	})

	// FatalLevel测试会导致程序退出，跳过此测试
	// t.Run("FatalLevel测试", func(t *testing.T) {
	// 	recorded.TakeAll()
	// 	microLog.Log(microlog.FatalLevel, "fatal消息")
	//
	// 	logs := recorded.All()
	// 	require.Len(t, logs, 1)
	// 	require.Equal(t, zapcore.FatalLevel, logs[0].Level)
	// })

	t.Run("未知Level测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Log(microlog.Level(100), "未知级别消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level) // 默认映射到InfoLevel
	})
}

// TestMicroLogger_Logf 测试microLogger的Logf方法
func TestMicroLogger_Logf(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	t.Run("InfoLevel格式化测试", func(t *testing.T) {
		microLog.Logf(microlog.InfoLevel, "用户 %s 执行了 %s 操作", "admin", "登录")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)
		require.Equal(t, "用户 admin 执行了 登录 操作", logs[0].Message)
	})

	t.Run("DebugLevel格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.DebugLevel, "调试信息: %d", 42)

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.DebugLevel, logs[0].Level)
		require.Equal(t, "调试信息: 42", logs[0].Message)
	})

	t.Run("TraceLevel格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.TraceLevel, "跟踪: %v", true)

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.DebugLevel, logs[0].Level) // TraceLevel映射到DebugLevel
		require.Equal(t, "跟踪: true", logs[0].Message)
	})

	t.Run("WarnLevel格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.WarnLevel, "警告: %s", "内存使用率过高")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.WarnLevel, logs[0].Level)
		require.Equal(t, "警告: 内存使用率过高", logs[0].Message)
	})

	t.Run("ErrorLevel格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.ErrorLevel, "错误代码: %d", 500)

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
		require.Equal(t, "错误代码: 500", logs[0].Message)
	})

	// FatalLevel测试会导致程序退出，跳过此测试
	// t.Run("FatalLevel格式化测试", func(t *testing.T) {
	// 	recorded.TakeAll()
	// 	microLog.Logf(microlog.FatalLevel, "致命错误: %s", "系统崩溃")
	//
	// 	logs := recorded.All()
	// 	require.Len(t, logs, 1)
	// 	require.Equal(t, zapcore.FatalLevel, logs[0].Level)
	// 	require.Equal(t, "致命错误: 系统崩溃", logs[0].Message)
	// })

	t.Run("未知Level格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.Level(100), "未知: %s", "测试")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level) // 默认映射到InfoLevel
		require.Equal(t, "未知: 测试", logs[0].Message)
	})

	t.Run("无参数格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.InfoLevel, "简单消息")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "简单消息", logs[0].Message)
	})
}

// TestMicroLogger_String 测试microLogger的String方法
func TestMicroLogger_String(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	require.Equal(t, "zap-micro", microLog.String())
}

// TestMicroLogger_Integration 集成测试
func TestMicroLogger_Integration(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	baseLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(baseLogger)

	t.Run("完整工作流测试", func(t *testing.T) {
		// 初始化选项
		err := microLog.Init(func(o *microlog.Options) {
			o.CallerSkipCount = 1
			o.Level = microlog.DebugLevel
		})
		require.NoError(t, err)

		// 添加字段
		loggerWithFields := microLog.Fields(map[string]interface{}{
			"service": "user-service",
			"version": "1.0.0",
		})

		// 记录不同级别的日志
		loggerWithFields.Log(microlog.InfoLevel, "服务启动")
		loggerWithFields.Logf(microlog.WarnLevel, "警告: %s", "配置文件缺失")
		loggerWithFields.Log(microlog.ErrorLevel, "服务异常")

		logs := recorded.All()
		require.Len(t, logs, 3)

		// 验证第一条日志
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)

		// 验证第二条日志
		require.Equal(t, zapcore.WarnLevel, logs[1].Level)
		require.Equal(t, "警告: 配置文件缺失", logs[1].Message)

		// 验证第三条日志
		require.Equal(t, zapcore.ErrorLevel, logs[2].Level)

		// 验证所有日志都包含字段
		for _, log := range logs {
			foundService := false
			foundVersion := false
			for _, field := range log.Context {
				if field.Key == "service" {
					foundService = true
				}
				if field.Key == "version" {
					foundVersion = true
				}
			}
			require.True(t, foundService, "应该包含service字段")
			require.True(t, foundVersion, "应该包含version字段")
		}
	})
}

// TestMicroLogger_EdgeCases 边界情况测试
func TestMicroLogger_EdgeCases(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	microLog := NewMicroLogger(targetLogger)

	t.Run("nil值测试", func(t *testing.T) {
		microLog.Log(microlog.InfoLevel, nil, "测试", nil)

		logs := recorded.All()
		require.Len(t, logs, 1)
	})

	t.Run("大量字段测试", func(t *testing.T) {
		recorded.TakeAll()
		fields := make(map[string]interface{})
		for i := 0; i < 100; i++ {
			fields[fmt.Sprintf("field_%d", i)] = i
		}

		loggerWithFields := microLog.Fields(fields)
		loggerWithFields.Log(microlog.InfoLevel, "大量字段测试")

		logs := recorded.All()
		require.Len(t, logs, 1)
		// Log方法会添加一个额外的字段，所以总数是101
		require.Len(t, logs[0].Context, 101)
	})

	t.Run("特殊字符测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Logf(microlog.InfoLevel, "特殊字符: %s %s %s", "\n", "\t", "🚀")

		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Contains(t, logs[0].Message, "特殊字符:")
	})

	t.Run("空字符串测试", func(t *testing.T) {
		recorded.TakeAll()
		microLog.Log(microlog.InfoLevel, "")
		microLog.Logf(microlog.InfoLevel, "")

		logs := recorded.All()
		require.Len(t, logs, 2)
	})
}
