package codeerr

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogCodeError_StackTrace(t *testing.T) {
	// 构造深层调用链
	err := a(context.Background())

	// 类型断言获取带堆栈的错误
	var stackTracer interface {
		StackTrace() errors.StackTrace
	}
	require.True(t, errors.As(err, &stackTracer), "错误对象应包含调用栈")

	// 验证调用栈
	stack := stackTracer.StackTrace()
	assert.GreaterOrEqual(t, len(stack), 3, "调用栈应至少包含3层")

	// 检查调用栈内容
	checkStackFrame(t, stack[0], "c") // 最深层调用
	checkStackFrame(t, stack[1], "b")
	checkStackFrame(t, stack[2], "a")

	// 验证过滤了工具方法本身
	for _, frame := range stack {
		file := getFilePath(frame)
		assert.False(t, strings.Contains(file, "code.go"),
			"调用栈不应包含工具方法所在文件")
	}
}

// 辅助函数：检查堆栈帧
func checkStackFrame(t *testing.T, frame errors.Frame, expectFunc string) {
	pc := uintptr(frame) - 1
	fn := runtime.FuncForPC(pc)
	require.NotNil(t, fn, "应能获取函数信息")

	funcName := filepath.Base(fn.Name())
	assert.Contains(t, funcName, expectFunc,
		"堆栈帧应包含预期函数名")
}

// 获取堆栈帧文件路径
func getFilePath(frame errors.Frame) string {
	pc := uintptr(frame) - 1
	fn := runtime.FuncForPC(pc)
	file, _ := fn.FileLine(pc)
	return file
}

/******************** 测试用调用链 ********************/
func a(ctx context.Context) error {
	return b(ctx)
}

func b(ctx context.Context) error {
	return c(ctx)
}

func c(ctx context.Context) error {
	return LogCodeError(ctx, "TEST_001",
		"测试错误: %s", "参数")
}
