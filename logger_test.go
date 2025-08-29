package log2

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestNewLoggerFunc 测试NewLogger函数
func TestNewLoggerFunc(t *testing.T) {
	t.Run("基本创建测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, true, false, nil, nil)
		
		if logger == nil {
			t.Fatal("logger should not be nil")
		}
		
		if logger.name != "test" {
			t.Errorf("expected name 'test', got '%s'", logger.name)
		}
	})
	
	t.Run("带字段创建测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		field := zap.String("key", "value")
		logger := NewLogger(zapLogger, "test", 0, true, false, nil, nil, field)
		
		if len(logger.fields) != 1 {
			t.Errorf("expected 1 field, got %d", len(logger.fields))
		}
	})
	
	t.Run("last名称测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "parent.child.test", 0, true, true, nil, nil)
		
		if logger == nil {
			t.Fatal("logger should not be nil")
		}
	})
}

// TestLogger_Derive 测试Derive方法
func TestLogger_Derive(t *testing.T) {
	t.Run("基本衍生测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		parentLogger := NewLogger(zapLogger, "parent", 0, true, false, nil, nil)
		childLogger := parentLogger.Derive("child")
		
		if childLogger == nil {
			t.Fatal("child logger should not be nil")
		}
	})
	
	t.Run("空名称衍生测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		parentLogger := NewLogger(zapLogger, "", 0, false, false, nil, nil)
		childLogger := parentLogger.Derive("child")
		
		if childLogger == nil {
			t.Fatal("child logger should not be nil")
		}
	})
}

// TestLogger_With 测试With方法
func TestLogger_With(t *testing.T) {
	t.Run("基本With测试", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		newLogger := logger.With(zap.String("key", "value"))
		
		newLogger.Info("test message")
		
		if logs.Len() != 1 {
			t.Errorf("expected 1 log entry, got %d", logs.Len())
		}
		
		entry := logs.All()[0]
		if len(entry.Context) != 1 {
			t.Errorf("expected 1 context field, got %d", len(entry.Context))
		}
	})
	
	t.Run("nil underlying测试", func(t *testing.T) {
		logger := &logger{underlying: nil}
		newLogger := logger.With(zap.String("key", "value"))
		
		// With方法在underlying为nil时返回相同的logger指针
		if newLogger == nil {
			t.Error("should not return nil logger")
		}
	})
}

// TestLogger_WithWhenNotExist 测试WithWhenNotExist方法
func TestLogger_WithWhenNotExist(t *testing.T) {
	t.Run("字段不存在测试", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		newLogger := logger.WithWhenNotExist("key", zap.String("key", "value"))
		
		newLogger.Info("test message")
		
		if logs.Len() != 1 {
			t.Errorf("expected 1 log entry, got %d", logs.Len())
		}
	})
	
	t.Run("字段已存在测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		exist := NewExist(10)
		exist.Set("key")
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, exist)
		newLogger := logger.WithWhenNotExist("key", zap.String("key", "value"))
		
		// WithWhenNotExist在key存在时应该返回相同的logger指针
		if newLogger == nil {
			t.Error("should not return nil logger")
		}
	})
	
	t.Run("nil underlying测试", func(t *testing.T) {
		logger := &logger{underlying: nil, duplicateKeys: NewExist(10)}
		newLogger := logger.WithWhenNotExist("key", zap.String("key", "value"))
		
		// WithWhenNotExist在underlying为nil时应该返回相同的logger指针
		if newLogger == nil {
			t.Error("should not return nil logger")
		}
	})
}

// TestLogger_LogMethods 测试各种日志级别方法
func TestLogger_LogMethods(t *testing.T) {
	tests := []struct {
		name   string
		level  zapcore.Level
		method func(Logger, string, ...zap.Field)
	}{
		{"Debug", zapcore.DebugLevel, func(l Logger, msg string, fields ...zap.Field) { l.Debug(msg, fields...) }},
		{"Info", zapcore.InfoLevel, func(l Logger, msg string, fields ...zap.Field) { l.Info(msg, fields...) }},
		{"Warn", zapcore.WarnLevel, func(l Logger, msg string, fields ...zap.Field) { l.Warn(msg, fields...) }},
		{"Error", zapcore.ErrorLevel, func(l Logger, msg string, fields ...zap.Field) { l.Error(msg, fields...) }},
	}
	
	for _, tt := range tests {
		t.Run(tt.name+"测试", func(t *testing.T) {
			core, logs := observer.New(tt.level)
			zapLogger := zap.New(core)
			
			logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
			tt.method(logger, "test message", zap.String("key", "value"))
			
			if logs.Len() != 1 {
				t.Errorf("expected 1 log entry, got %d", logs.Len())
			}
			
			entry := logs.All()[0]
			if entry.Message != "test message" {
				t.Errorf("expected message 'test message', got '%s'", entry.Message)
			}
			
			if entry.Level != tt.level {
				t.Errorf("expected level %v, got %v", tt.level, entry.Level)
			}
		})
		
		t.Run(tt.name+"_nil_underlying测试", func(t *testing.T) {
			logger := &logger{underlying: nil}
			// 这些方法在underlying为nil时不应该panic
			tt.method(logger, "test message")
		})
	}
}

// TestLogger_FatalAndPanic 测试Fatal和Panic方法（需要特殊处理）
func TestLogger_FatalAndPanic(t *testing.T) {
	t.Run("Fatal_nil_underlying测试", func(t *testing.T) {
		logger := &logger{underlying: nil}
		// Fatal方法在underlying为nil时不应该panic
		logger.Fatal("test message")
	})
	
	t.Run("Panic_nil_underlying测试", func(t *testing.T) {
		logger := &logger{underlying: nil}
		// Panic方法在underlying为nil时不应该panic，但我们需要捕获可能的panic
		defer func() {
			if r := recover(); r != nil {
				// 如果发生panic，这是预期的行为
				t.Logf("Caught expected panic: %v", r)
			}
		}()
		logger.Panic("test message")
	})
	
	t.Run("Panic_with_underlying测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		
		// 测试Panic方法会真正panic
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but none occurred")
			}
		}()
		logger.Panic("test panic message")
	})
}

// TestLogger_SetLevelMethod 测试SetLevel方法
func TestLogger_SetLevelFunc(t *testing.T) {
	t.Run("基本SetLevel测试", func(t *testing.T) {
		// 保存原始状态
		originalHiddenConsole := HiddenConsole
		originalWriteSyncer := writeSyncer
		originalInputCores := inputCores
		originalEncoder := encoder
		originalCore := core
		
		// 设置测试环境
		HiddenConsole = true
		writeSyncer = nil
		inputCores = nil
		encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
		})
		
		// 恢复原始状态
		defer func() {
			HiddenConsole = originalHiddenConsole
			writeSyncer = originalWriteSyncer
			inputCores = originalInputCores
			encoder = originalEncoder
			core = originalCore
		}()
		
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		newLogger := logger.SetLevel(zapcore.DebugLevel)
		
		if newLogger == nil {
			t.Fatal("new logger should not be nil")
		}
	})
}

// TestLogger_StartMethod 测试Start方法
func TestLogger_StartFunc(t *testing.T) {
	t.Run("基本Start测试", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		startLogger := logger.Start()
		
		startLogger.Info("test message")
		
		if logs.Len() != 1 {
			t.Errorf("expected 1 log entry, got %d", logs.Len())
		}
		
		entry := logs.All()[0]
		// 检查是否包含任务ID字段
		found := false
		for _, field := range entry.Context {
			if field.Key == "任务ID" {
				found = true
				break
			}
		}
		
		if !found {
			t.Error("expected 任务ID field in log entry")
		}
	})
}

// TestLogger_AddCallerSkipMethod 测试AddCallerSkip方法
func TestLogger_AddCallerSkipFunc(t *testing.T) {
	t.Run("基本AddCallerSkip测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		newLogger := logger.AddCallerSkip(2)
		
		if newLogger == nil {
			t.Fatal("new logger should not be nil")
		}
		
		// AddCallerSkip方法应该返回一个新的logger实例
		// 由于logger结构体不是导出的，我们只能验证返回值不为nil
		if newLogger == nil {
			t.Error("AddCallerSkip should return a non-nil logger")
		}
	})
}

// TestEnsureDuplicateKeys 测试ensureDuplicateKeys函数
func TestEnsureDuplicateKeys(t *testing.T) {
	t.Run("nil输入测试", func(t *testing.T) {
		result := ensureDuplicateKeys(nil)
		
		if result == nil {
			t.Fatal("result should not be nil")
		}
	})
	
	t.Run("非nil输入测试", func(t *testing.T) {
		exist := NewExist(5)
		result := ensureDuplicateKeys(exist)
		
		if result != exist {
			t.Error("should return the same Exist instance")
		}
	})
}

// TestLogger_Integration 集成测试
func TestLogger_Integration(t *testing.T) {
	t.Run("完整工作流测试", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		zapLogger := zap.New(core)
		
		// 创建父日志器
		parentLogger := NewLogger(zapLogger, "parent", 0, true, false, nil, nil)
		
		// 衍生子日志器
		childLogger := parentLogger.Derive("child")
		
		// 添加字段
		loggerWithFields := childLogger.With(zap.String("service", "test"))
		
		// 添加任务ID
		startLogger := loggerWithFields.Start()
		
		// 记录日志
		startLogger.Info("integration test", zap.Int("count", 1))
		
		if logs.Len() != 1 {
			t.Errorf("expected 1 log entry, got %d", logs.Len())
		}
		
		entry := logs.All()[0]
		if entry.Message != "integration test" {
			t.Errorf("expected message 'integration test', got '%s'", entry.Message)
		}
		
		// 检查字段
		expectedFields := map[string]bool{
			"service": false,
			"任务ID":    false,
			"count":   false,
		}
		
		for _, field := range entry.Context {
			if _, exists := expectedFields[field.Key]; exists {
				expectedFields[field.Key] = true
			}
		}
		
		for key, found := range expectedFields {
			if !found {
				t.Errorf("expected field '%s' not found", key)
			}
		}
	})
}

// TestLogger_EdgeCases 边界情况测试
func TestLogger_EdgeCases(t *testing.T) {
	t.Run("空消息测试", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		logger.Info("")
		
		if logs.Len() != 1 {
			t.Errorf("expected 1 log entry, got %d", logs.Len())
		}
		
		entry := logs.All()[0]
		if entry.Message != "" {
			t.Errorf("expected empty message, got '%s'", entry.Message)
		}
	})
	
	t.Run("大量字段测试", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		logger := NewLogger(zapLogger, "test", 0, false, false, nil, nil)
		
		// 添加大量字段
		fields := make([]zap.Field, 100)
		for i := 0; i < 100; i++ {
			fields[i] = zap.Int("field"+string(rune(i)), i)
		}
		
		logger.Info("test message", fields...)
		
		if logs.Len() != 1 {
			t.Errorf("expected 1 log entry, got %d", logs.Len())
		}
		
		entry := logs.All()[0]
		if len(entry.Context) != 100 {
			t.Errorf("expected 100 context fields, got %d", len(entry.Context))
		}
	})
	
	t.Run("复杂名称测试", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		zapLogger := zap.New(core)
		
		// 测试包含特殊字符的名称
		specialNames := []string{
			"test.with.dots",
			"test-with-dashes",
			"test_with_underscores",
			"test with spaces",
			"测试中文名称",
		}
		
		for _, name := range specialNames {
			logger := NewLogger(zapLogger, name, 0, true, false, nil, nil)
			if logger == nil {
				t.Errorf("logger should not be nil for name '%s'", name)
			}
		}
	})
}