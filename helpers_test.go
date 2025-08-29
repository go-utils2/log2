package log2

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewExist 测试Exist结构体的创建
func TestNewExist(t *testing.T) {
	t.Run("正常创建", func(t *testing.T) {
		exist := NewExist(10)
		require.NotNil(t, exist)
		require.NotNil(t, exist.lock)
		require.NotNil(t, exist.data)
		require.Equal(t, 0, len(exist.data))
	})

	t.Run("零容量创建", func(t *testing.T) {
		exist := NewExist(0)
		require.NotNil(t, exist)
		require.NotNil(t, exist.data)
	})
}

// TestExist_Exist 测试Exist方法
func TestExist_Exist(t *testing.T) {
	exist := NewExist(5)

	t.Run("不存在的键", func(t *testing.T) {
		require.False(t, exist.Exist("nonexistent"))
	})

	t.Run("存在的键", func(t *testing.T) {
		exist.Set("test-key")
		require.True(t, exist.Exist("test-key"))
	})

	t.Run("空字符串键", func(t *testing.T) {
		exist.Set("")
		require.True(t, exist.Exist(""))
	})
}

// TestExist_Set 测试Set方法
func TestExist_Set(t *testing.T) {
	exist := NewExist(5)

	t.Run("设置新键", func(t *testing.T) {
		exist.Set("new-key")
		require.True(t, exist.Exist("new-key"))
	})

	t.Run("重复设置同一键", func(t *testing.T) {
		exist.Set("duplicate-key")
		exist.Set("duplicate-key")
		require.True(t, exist.Exist("duplicate-key"))
	})

	t.Run("设置多个键", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		for _, key := range keys {
			exist.Set(key)
		}
		for _, key := range keys {
			require.True(t, exist.Exist(key))
		}
	})
}

// TestExist_Copy 测试Copy方法
func TestExist_Copy(t *testing.T) {
	t.Run("复制空的Exist", func(t *testing.T) {
		original := NewExist(5)
		copy := original.Copy()
		
		require.NotNil(t, copy)
		require.NotSame(t, original, copy)
		require.NotSame(t, &original.data, &copy.data)
		require.Equal(t, len(original.data), len(copy.data))
	})

	t.Run("复制有数据的Exist", func(t *testing.T) {
		original := NewExist(5)
		keys := []string{"key1", "key2", "key3"}
		
		// 在原始对象中设置键
		for _, key := range keys {
			original.Set(key)
		}
		
		// 复制对象
		copy := original.Copy()
		
		// 验证复制的对象
		require.NotSame(t, original, copy)
		require.NotSame(t, &original.data, &copy.data)
		require.Equal(t, len(original.data), len(copy.data))
		
		// 验证所有键都被复制
		for _, key := range keys {
			require.True(t, copy.Exist(key))
		}
	})

	t.Run("复制后独立性测试", func(t *testing.T) {
		original := NewExist(5)
		original.Set("original-key")
		
		copy := original.Copy()
		
		// 在复制对象中添加新键
		copy.Set("copy-key")
		
		// 在原始对象中添加新键
		original.Set("new-original-key")
		
		// 验证独立性
		require.True(t, original.Exist("original-key"))
		require.True(t, original.Exist("new-original-key"))
		require.False(t, original.Exist("copy-key"))
		
		require.True(t, copy.Exist("original-key"))
		require.True(t, copy.Exist("copy-key"))
		require.False(t, copy.Exist("new-original-key"))
	})
}

// TestExist_Concurrency 测试并发安全性
func TestExist_Concurrency(t *testing.T) {
	exist := NewExist(100)
	var wg sync.WaitGroup
	goroutineCount := 10
	operationsPerGoroutine := 100

	t.Run("并发Set和Exist", func(t *testing.T) {
		// 启动多个goroutine进行并发操作
		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					key := fmt.Sprintf("key-%d-%d", id, j)
					exist.Set(key)
					require.True(t, exist.Exist(key))
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("并发Copy", func(t *testing.T) {
		// 先设置一些数据
		for i := 0; i < 50; i++ {
			exist.Set(fmt.Sprintf("base-key-%d", i))
		}

		// 并发复制
		copies := make([]*Exist, goroutineCount)
		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				copies[id] = exist.Copy()
			}(i)
		}
		wg.Wait()

		// 验证所有复制都成功
		for i, copy := range copies {
			require.NotNil(t, copy, "Copy %d should not be nil", i)
			require.True(t, copy.Exist("base-key-0"))
		}
	})
}

// TestExist_EdgeCases 测试边界情况
func TestExist_EdgeCases(t *testing.T) {
	t.Run("特殊字符键", func(t *testing.T) {
		exist := NewExist(5)
		specialKeys := []string{
			"key with spaces",
			"key\nwith\nnewlines",
			"key\twith\ttabs",
			"键中文",
			"🔑emoji",
			"very-long-key-" + string(make([]byte, 1000)),
		}

		for _, key := range specialKeys {
			exist.Set(key)
			require.True(t, exist.Exist(key), "Key should exist: %s", key)
		}
	})

	t.Run("大量数据", func(t *testing.T) {
		exist := NewExist(10000)
		count := 5000

		// 设置大量数据
		for i := 0; i < count; i++ {
			exist.Set(fmt.Sprintf("bulk-key-%d", i))
		}

		// 验证所有数据
		for i := 0; i < count; i++ {
			require.True(t, exist.Exist(fmt.Sprintf("bulk-key-%d", i)))
		}

		// 复制大量数据
		copy := exist.Copy()
		for i := 0; i < count; i++ {
			require.True(t, copy.Exist(fmt.Sprintf("bulk-key-%d", i)))
		}
	})
}