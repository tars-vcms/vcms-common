package errs

import (
	"context"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/pkg/errors"
	"io"
	"runtime"
	"strconv"
	"strings"
)

type Error struct {
	Code int32
	Msg  string
	Desc string

	st []uintptr // 调用栈
}

// Success 成功提示字符串
const (
	Success = "success"
)

var (
	traceable bool   = false
	content   string = "VCMS-Server"
)

const (
	RetOK = 0

	RetServerContextErr = 101

	RetUnknown = 999
)

// Error 实现error接口，返回error描述
func (e *Error) Error() string {
	if e == nil {
		return Success
	}
	return fmt.Sprintf("code:%d, msg:%s", e.Code, e.Msg)
}

// Format 实现fmt.Formatter接口
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.Error())
			for _, pc := range e.st {
				f := errors.Frame(pc)
				str := fmt.Sprintf("\n%+v", f)
				if !isOutput(str) {
					continue
				}
				_, _ = io.WriteString(s, str)
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	default:
		// unknown format
		_, _ = fmt.Fprintf(s, "%%!%c(errs.Error=%s)", verb, e.Error())
	}
}

func callers() []uintptr {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	st := pcs[0:n]
	return st
}

func isOutput(str string) bool {
	return strings.Contains(str, content)
}

// SetTraceable 控制error是否带堆栈跟踪
func SetTraceable(x bool) {
	traceable = x
}

// SetTraceableWithContent 控制error是否带堆栈跟踪，打印堆栈信息时，根据content进行过滤。
// 避免输出大量无用信息。可以通过配置content为服务名的方式，过滤掉其他插件的堆栈信息。
func SetTraceableWithContent(c string) {
	traceable = true
	content = c
}

// New 创建一个error，默认为业务错误类型，提高业务开发效率
func New(code int, msg string) error {
	err := &Error{
		Code: int32(code),
		Msg:  msg,
	}
	if traceable {
		err.st = callers()
	}
	return err
}

// Newf 创建一个error，默认为业务错误类型，msg支持格式化字符串
func Newf(code int, format string, params ...interface{}) error {
	msg := fmt.Sprintf(format, params...)
	return New(code, msg)
}

// Code 通过error获取error code
func Code(e error) int {
	if e == nil {
		return 0
	}
	err, ok := e.(*Error)
	if !ok {
		return RetUnknown
	}
	if err == (*Error)(nil) {
		return 0
	}
	return int(err.Code)
}

// Msg 通过error获取error msg
func Msg(e error) string {
	if e == nil {
		return Success
	}
	err, ok := e.(*Error)
	if !ok {
		return e.Error()
	}
	if err == (*Error)(nil) {
		return Success
	}
	return err.Msg
}

// HandleError 将报错信息打包到context中
func HandleError(ctx context.Context, err error) (int32, error) {
	k, ok := current.GetResponseContext(ctx)
	if !ok {
		//无法获取context 直接抛出error
		return RetServerContextErr, err
	}
	//判断ctx中是否已经初始化
	if k == nil {
		k = make(map[string]string)
	}
	ret := 0
	msg := "success"
	//error为空直接返回
	if err == nil {
		return RetOK, nil
	} else {
		ret = Code(err)
		msg = Msg(err)
	}

	k["tars-ret"] = strconv.Itoa(ret)
	k["tars-msg"] = msg
	ok = current.SetResponseContext(ctx, k)
	if ok {
		return int32(ret), nil
	} else {
		return int32(ret), err
	}

}
