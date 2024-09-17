package codeerr

import (
	"codexie.com/w-book-user/pkg/common/response"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type withCode struct {
	msg  string
	code int
}

func WithCode(code int, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return errors.Wrap(&withCode{
		msg:  msg,
		code: code,
	}, msg)
}

func (e *withCode) StackTrace() errors.StackTrace {
	// Get full stack trace
	stack := errors.WithStack(e).(interface {
		StackTrace() errors.StackTrace
	}).StackTrace()

	// Remove the top frame (which corresponds to the `WithCode` call)
	return stack[1:] // Skip the first stack frame
}

// HandleErr 日志记录异常并封装返回给用户的异常响应
//
//	@Description:
//	@param ctx
//	@param err
//	@return *response.Response
func HandleErr(ctx context.Context, err error) *response.Response {
	logx.WithContext(ctx).Errorf("%+v", err)
	coder := ParseCoder(err)
	return response.Fail(coder.Code(), coder.String())
}

func (w *withCode) Error() string { return w.msg }
