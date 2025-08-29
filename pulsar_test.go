package log2

import (
	"errors"
	"fmt"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestNewPulsarLogger 测试NewPulsarLogger函数
func TestNewPulsarLogger(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)

	pulsarLog := NewPulsarLogger(targetLogger)

	require.NotNil(t, pulsarLog)
	require.Equal(t, targetLogger, pulsarLog.Logger)
}

// TestPulsarLogger_SubLogger 测试pulsarLogger的SubLogger方法
func TestPulsarLogger_SubLogger(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("基本SubLogger测试", func(t *testing.T) {
		fields := log.Fields{
			"service": "pulsar-service",
			"version": "1.0.0",
		}
		
		subLogger := pulsarLog.SubLogger(fields)
		require.NotNil(t, subLogger)
		require.NotEqual(t, pulsarLog, subLogger)
		
		// 测试子logger是否包含字段
		subLogger.Info("测试消息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message) // Info方法使用空消息
		
		// 检查字段是否存在
		foundService := false
		foundVersion := false
		for _, field := range logs[0].Context {
			if field.Key == "service" {
				foundService = true
				require.Equal(t, "pulsar-service", field.String)
			}
			if field.Key == "version" {
				foundVersion = true
				require.Equal(t, "1.0.0", field.String)
			}
		}
		require.True(t, foundService, "应该包含service字段")
		require.True(t, foundVersion, "应该包含version字段")
	})

	t.Run("空Fields测试", func(t *testing.T) {
		recorded.TakeAll()
		fields := log.Fields{}
		
		subLogger := pulsarLog.SubLogger(fields)
		require.NotNil(t, subLogger)
		
		subLogger.Info("空字段测试")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message)
	})
}

// TestPulsarLogger_WithFields 测试pulsarLogger的WithFields方法
func TestPulsarLogger_WithFields(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("基本WithFields测试", func(t *testing.T) {
		fields := log.Fields{
			"user_id": 123,
			"action":  "login",
		}
		
		entry := pulsarLog.WithFields(fields)
		require.NotNil(t, entry)
		
		// WithFields应该返回log.Entry接口
		entry.Info("用户登录")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message)
		
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
}

// TestPulsarLogger_WithField 测试pulsarLogger的WithField方法
func TestPulsarLogger_WithField(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("基本WithField测试", func(t *testing.T) {
		entry := pulsarLog.WithField("request_id", "req-123")
		require.NotNil(t, entry)
		
		entry.Info("处理请求")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message)
		
		// 检查字段是否存在
		foundRequestID := false
		for _, field := range logs[0].Context {
			if field.Key == "request_id" {
				foundRequestID = true
				require.Equal(t, "req-123", field.String)
			}
		}
		require.True(t, foundRequestID, "应该包含request_id字段")
	})

	t.Run("数字字段测试", func(t *testing.T) {
		recorded.TakeAll()
		entry := pulsarLog.WithField("count", 42)
		require.NotNil(t, entry)
		
		entry.Info("计数信息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		
		// 检查数字字段
		foundCount := false
		for _, field := range logs[0].Context {
			if field.Key == "count" {
				foundCount = true
				require.Equal(t, int64(42), field.Integer)
			}
		}
		require.True(t, foundCount, "应该包含count字段")
	})
}

// TestPulsarLogger_WithError 测试pulsarLogger的WithError方法
func TestPulsarLogger_WithError(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("基本WithError测试", func(t *testing.T) {
		testErr := errors.New("测试错误")
		entry := pulsarLog.WithError(testErr)
		require.NotNil(t, entry)
		
		entry.Error("发生错误")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
		require.Equal(t, "", logs[0].Message)
		
		// 检查错误字段 - Error方法会添加额外的"参数"字段，所以总共有2个字段
		require.Len(t, logs[0].Context, 2, "应该有error和参数两个字段")
		
		// 查找错误字段
		foundError := false
		for _, field := range logs[0].Context {
			if field.Key == "error" {
				foundError = true
				require.Equal(t, zapcore.ErrorType, field.Type)
				require.Equal(t, "测试错误", field.Interface.(error).Error())
			}
		}
		require.True(t, foundError, "应该包含error字段")
	})

	t.Run("nil错误测试", func(t *testing.T) {
		recorded.TakeAll()
		entry := pulsarLog.WithError(nil)
		require.NotNil(t, entry)
		
		entry.Info("无错误")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
	})
}

// TestPulsarLogger_LogMethods 测试pulsarLogger的各种日志方法
func TestPulsarLogger_LogMethods(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("Debug方法测试", func(t *testing.T) {
		pulsarLog.Debug("debug消息", 123)
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.DebugLevel, logs[0].Level)
		require.Equal(t, "", logs[0].Message)
		
		// 检查参数字段
		foundParams := false
		for _, field := range logs[0].Context {
			if field.Key == "参数" {
				foundParams = true
			}
		}
		require.True(t, foundParams, "应该包含参数字段")
	})

	t.Run("Info方法测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Info("info消息", "额外参数")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)
		require.Equal(t, "", logs[0].Message)
	})

	t.Run("Warn方法测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Warn("warn消息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.WarnLevel, logs[0].Level)
		require.Equal(t, "", logs[0].Message)
	})

	t.Run("Error方法测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Error("error消息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
		require.Equal(t, "", logs[0].Message)
	})
}

// TestPulsarLogger_FormatMethods 测试pulsarLogger的格式化方法
func TestPulsarLogger_FormatMethods(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("Debugf方法测试", func(t *testing.T) {
		pulsarLog.Debugf("调试信息: %s, 数字: %d", "测试", 42)
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.DebugLevel, logs[0].Level)
		require.Equal(t, "调试信息: 测试, 数字: 42", logs[0].Message)
	})

	t.Run("Infof方法测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Infof("用户 %s 执行了 %s 操作", "admin", "登录")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)
		require.Equal(t, "用户 admin 执行了 登录 操作", logs[0].Message)
	})

	t.Run("Warnf方法测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Warnf("警告: %s", "内存使用率过高")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.WarnLevel, logs[0].Level)
		require.Equal(t, "警告: 内存使用率过高", logs[0].Message)
	})

	t.Run("Errorf方法测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Errorf("错误代码: %d, 描述: %s", 500, "内部服务器错误")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
		require.Equal(t, "错误代码: 500, 描述: 内部服务器错误", logs[0].Message)
	})

	t.Run("无参数格式化测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Infof("简单消息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "简单消息", logs[0].Message)
	})
}

// TestPulsarLogger_Integration 集成测试
func TestPulsarLogger_Integration(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	zapLogger := zap.New(core)
	baseLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(baseLogger)

	t.Run("完整工作流测试", func(t *testing.T) {
		// 创建带字段和错误的logger
		testErr := errors.New("连接失败")
		loggerWithError := pulsarLog.WithError(testErr).WithFields(log.Fields{
			"service": "pulsar-service",
			"version": "1.0.0",
		}).WithField("request_id", "req-456")
		
		// 记录不同级别的日志
		loggerWithError.Debug("调试信息")
		loggerWithError.Infof("处理请求: %s", "req-456")
		loggerWithError.Warnf("警告: %s", "连接不稳定")
		loggerWithError.Error("处理失败")
		
		logs := recorded.All()
		require.Len(t, logs, 4)
		
		// 验证第一条日志（Debug）
		require.Equal(t, zapcore.DebugLevel, logs[0].Level)
		
		// 验证第二条日志（Infof）
		require.Equal(t, zapcore.InfoLevel, logs[1].Level)
		require.Equal(t, "处理请求: req-456", logs[1].Message)
		
		// 验证第三条日志（Warnf）
		require.Equal(t, zapcore.WarnLevel, logs[2].Level)
		require.Equal(t, "警告: 连接不稳定", logs[2].Message)
		
		// 验证第四条日志（Error）
		require.Equal(t, zapcore.ErrorLevel, logs[3].Level)
		
		// 验证所有日志都包含基础字段
		for _, log := range logs {
			foundService := false
			foundVersion := false
			foundRequestID := false
			foundError := false
			
			for _, field := range log.Context {
				switch field.Key {
				case "service":
					foundService = true
				case "version":
					foundVersion = true
				case "request_id":
					foundRequestID = true
				case "error":
					foundError = true
				}
			}
			
			require.True(t, foundService, "应该包含service字段")
			require.True(t, foundVersion, "应该包含version字段")
			require.True(t, foundRequestID, "应该包含request_id字段")
			require.True(t, foundError, "应该包含error字段")
		}
	})
}

// TestPulsarLogger_EdgeCases 边界情况测试
func TestPulsarLogger_EdgeCases(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	targetLogger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
	pulsarLog := NewPulsarLogger(targetLogger)

	t.Run("nil值测试", func(t *testing.T) {
		pulsarLog.Info(nil, "测试", nil)
		
		logs := recorded.All()
		require.Len(t, logs, 1)
	})

	t.Run("大量字段测试", func(t *testing.T) {
		recorded.TakeAll()
		fields := make(log.Fields)
		for i := 0; i < 50; i++ {
			fields[fmt.Sprintf("field_%d", i)] = i
		}
		
		loggerWithFields := pulsarLog.WithFields(fields)
		loggerWithFields.Info("大量字段测试")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		// Info方法会添加一个额外的字段，所以总数是51
		require.Len(t, logs[0].Context, 51)
	})

	t.Run("特殊字符测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Infof("特殊字符: %s %s %s", "\n", "\t", "🚀")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Contains(t, logs[0].Message, "特殊字符:")
	})

	t.Run("空字符串测试", func(t *testing.T) {
		recorded.TakeAll()
		pulsarLog.Info("")
		pulsarLog.Infof("")
		
		logs := recorded.All()
		require.Len(t, logs, 2)
	})

	t.Run("链式调用测试", func(t *testing.T) {
		recorded.TakeAll()
		// 分步骤创建logger，因为WithError只在pulsarLogger上可用
		loggerWithError := pulsarLog.WithError(errors.New("验证失败"))
		loggerWithFields := loggerWithError.WithField("step", 1).WithField("process", "validation")
		loggerWithFields.Error("处理步骤失败")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
		
		// 验证所有字段都存在
		foundStep := false
		foundProcess := false
		foundError := false
		
		for _, field := range logs[0].Context {
			switch field.Key {
			case "step":
				foundStep = true
			case "process":
				foundProcess = true
			case "error":
				foundError = true
			}
		}
		
		require.True(t, foundStep, "应该包含step字段")
		require.True(t, foundProcess, "应该包含process字段")
		require.True(t, foundError, "应该包含error字段")
	})
}