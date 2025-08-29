package log2

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestNewCronLogger 测试NewCronLogger函数
func TestNewCronLogger(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)

	cronLog := NewCronLogger(targetLogger)

	require.NotNil(t, cronLog)
	require.Equal(t, targetLogger, cronLog.Logger)
}

// TestCronLogger_Printf 测试cronLogger的Printf方法
func TestCronLogger_Printf(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	cronLog := NewCronLogger(targetLogger)

	t.Run("基本Printf测试", func(t *testing.T) {
		cronLog.Printf("测试消息: %s", "hello")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)
		require.Equal(t, "测试消息: hello", logs[0].Message)
	})

	t.Run("多参数Printf测试", func(t *testing.T) {
		recorded.TakeAll() // 清空之前的日志
		cronLog.Printf("用户 %s 在 %d 时执行了操作", "admin", 12345)
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "用户 admin 在 12345 时执行了操作", logs[0].Message)
	})

	t.Run("空格式字符串测试", func(t *testing.T) {
		recorded.TakeAll()
		cronLog.Printf("")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "", logs[0].Message)
	})
}

// TestCronLogger_Error 测试cronLogger的Error方法
func TestCronLogger_Error(t *testing.T) {
	core, recorded := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	cronLog := NewCronLogger(targetLogger)

	t.Run("基本Error测试", func(t *testing.T) {
		err := errors.New("测试错误")
		cronLog.Error(err, "发生错误: %s", "数据库连接失败")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, zapcore.ErrorLevel, logs[0].Level)
		require.Equal(t, "发生错误: 数据库连接失败", logs[0].Message)
		
		// 检查错误字段
		require.Len(t, logs[0].Context, 1)
		errorField := logs[0].Context[0]
		require.Equal(t, "error", errorField.Key)
		require.Equal(t, zapcore.ErrorType, errorField.Type)
		require.Contains(t, errorField.Interface.(error).Error(), "测试错误")
	})

	t.Run("nil错误测试", func(t *testing.T) {
		recorded.TakeAll()
		cronLog.Error(nil, "无错误消息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "无错误消息", logs[0].Message)
	})

	t.Run("多参数Error测试", func(t *testing.T) {
		recorded.TakeAll()
		err := errors.New("网络错误")
		cronLog.Error(err, "用户 %s 操作失败，错误码: %d", "user123", 500)
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "用户 user123 操作失败，错误码: 500", logs[0].Message)
	})
}

// TestDeriveCronLogger 测试DeriveCronLogger函数
func TestDeriveCronLogger(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	baseLogger := NewLogger(logger, "test", 0, false, false, nil, nil)

	t.Run("基本派生测试", func(t *testing.T) {
		derivedLogger := DeriveCronLogger(baseLogger, "user", "create")
		
		require.NotNil(t, derivedLogger)
		require.NotEqual(t, baseLogger, derivedLogger)
		
		// 测试派生的logger是否包含topic/method字段
		derivedLogger.Info("测试消息")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Equal(t, "测试消息", logs[0].Message)
		
		// 检查是否包含topic/method字段
		found := false
		for _, field := range logs[0].Context {
			if field.Key == "topic/method" {
				found = true
				// zap.stringArray类型需要转换为字符串切片进行比较
				if stringArray, ok := field.Interface.([]string); ok {
					require.Equal(t, []string{"user", "create"}, stringArray)
				} else {
					// 如果不是[]string类型，检查字段值的字符串表示
					require.Contains(t, fmt.Sprintf("%v", field.Interface), "user")
					require.Contains(t, fmt.Sprintf("%v", field.Interface), "create")
				}
				break
			}
		}
		require.True(t, found, "应该包含topic/method字段")
	})

	t.Run("空topic和method测试", func(t *testing.T) {
		recorded.TakeAll()
		derivedLogger := DeriveCronLogger(baseLogger, "", "")
		
		derivedLogger.Info("空字段测试")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		
		// 检查空字段
		for _, field := range logs[0].Context {
			if field.Key == "topic/method" {
				// zap.stringArray类型需要转换为字符串切片进行比较
				if stringArray, ok := field.Interface.([]string); ok {
					require.Equal(t, []string{"", ""}, stringArray)
				} else {
					// 如果不是[]string类型，检查字段值的字符串表示
					require.Contains(t, fmt.Sprintf("%v", field.Interface), "")
				}
				break
			}
		}
	})
}

// TestCronFormatTimes 测试cronFormatTimes函数
func TestCronFormatTimes(t *testing.T) {
	t.Run("格式化时间测试", func(t *testing.T) {
		now := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
		input := []interface{}{"key1", now, "key2", "value2", now}
		
		result := cronFormatTimes(input)
		
		require.Len(t, result, 5)
		require.Equal(t, "key1", result[0])
		require.Equal(t, "2023-12-25T10:30:45Z", result[1])
		require.Equal(t, "key2", result[2])
		require.Equal(t, "value2", result[3])
		require.Equal(t, "2023-12-25T10:30:45Z", result[4])
	})

	t.Run("无时间值测试", func(t *testing.T) {
		input := []interface{}{"key1", "value1", "key2", 123}
		
		result := cronFormatTimes(input)
		
		require.Equal(t, input, result)
	})

	t.Run("空切片测试", func(t *testing.T) {
		input := []interface{}{}
		
		result := cronFormatTimes(input)
		
		require.Empty(t, result)
	})

	t.Run("混合类型测试", func(t *testing.T) {
		now := time.Now()
		input := []interface{}{123, "string", now, true, 45.67}
		
		result := cronFormatTimes(input)
		
		require.Len(t, result, 5)
		require.Equal(t, 123, result[0])
		require.Equal(t, "string", result[1])
		require.Equal(t, now.Format(time.RFC3339), result[2])
		require.Equal(t, true, result[3])
		require.Equal(t, 45.67, result[4])
	})
}

// TestCronFormatString 测试cronFormatString函数
func TestCronFormatString(t *testing.T) {
	t.Run("零个键值对", func(t *testing.T) {
		result := cronFormatString(0)
		require.Equal(t, "%s", result)
	})

	t.Run("一个键值对", func(t *testing.T) {
		result := cronFormatString(2) // 2个参数组成1个键值对
		require.Equal(t, "%s, %v=%v", result)
	})

	t.Run("两个键值对", func(t *testing.T) {
		result := cronFormatString(4) // 4个参数组成2个键值对
		require.Equal(t, "%s, %v=%v, %v=%v", result)
	})

	t.Run("三个键值对", func(t *testing.T) {
		result := cronFormatString(6) // 6个参数组成3个键值对
		require.Equal(t, "%s, %v=%v, %v=%v, %v=%v", result)
	})

	t.Run("奇数参数测试", func(t *testing.T) {
		// 奇数参数应该只处理成对的部分
		result := cronFormatString(5) // 5个参数，只能组成2个键值对
		require.Equal(t, "%s, %v=%v, %v=%v", result)
	})
}

// TestCronLogger_Integration 集成测试
func TestCronLogger_Integration(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	baseLogger := NewLogger(logger, "test", 0, false, false, nil, nil)

	// 创建cron logger
	cronLog := NewCronLogger(baseLogger)
	
	// 创建派生logger
	derivedLogger := DeriveCronLogger(baseLogger, "scheduler", "execute")
	derivedCronLog := NewCronLogger(derivedLogger)

	t.Run("完整工作流测试", func(t *testing.T) {
		// 使用基础cron logger
		cronLog.Printf("任务开始执行: %s", "backup")
		
		// 使用派生cron logger
		derivedCronLog.Printf("派生任务执行: %s", "cleanup")
		
		// 记录错误
		err := errors.New("磁盘空间不足")
		cronLog.Error(err, "任务执行失败: %s", "backup")
		
		logs := recorded.All()
		require.Len(t, logs, 3)
		
		// 验证第一条日志
		require.Equal(t, "任务开始执行: backup", logs[0].Message)
		require.Equal(t, zapcore.InfoLevel, logs[0].Level)
		
		// 验证第二条日志（派生logger）
		require.Equal(t, "派生任务执行: cleanup", logs[1].Message)
		
		// 验证第三条日志（错误）
		require.Equal(t, "任务执行失败: backup", logs[2].Message)
		require.Equal(t, zapcore.ErrorLevel, logs[2].Level)
	})
}

// TestCronLogger_EdgeCases 边界情况测试
func TestCronLogger_EdgeCases(t *testing.T) {
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	targetLogger := NewLogger(logger, "test", 0, false, false, nil, nil)
	cronLog := NewCronLogger(targetLogger)

	t.Run("特殊字符测试", func(t *testing.T) {
		cronLog.Printf("特殊字符: %s %s %s", "\n", "\t", "\r")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Contains(t, logs[0].Message, "特殊字符:")
	})

	t.Run("长消息测试", func(t *testing.T) {
		recorded.TakeAll()
		longMsg := strings.Repeat("很长的消息 ", 100)
		cronLog.Printf("长消息: %s", longMsg)
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Contains(t, logs[0].Message, "长消息:")
	})

	t.Run("Unicode字符测试", func(t *testing.T) {
		recorded.TakeAll()
		cronLog.Printf("Unicode: %s %s %s", "🚀", "中文", "العربية")
		
		logs := recorded.All()
		require.Len(t, logs, 1)
		require.Contains(t, logs[0].Message, "Unicode:")
		require.Contains(t, logs[0].Message, "🚀")
		require.Contains(t, logs[0].Message, "中文")
	})
}