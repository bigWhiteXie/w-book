package codeerr

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var codes = map[int]Coder{}
var codeMux = &sync.Mutex{}

type Coder interface {
	// HTTP status that should be used for the associated error code.
	HTTPStatus() int

	// External (user) facing error text.
	String() string

	// Code returns the code of the coder
	Code() int
}

type ErrCode struct {
	// C refers to the code of the ErrCode.
	C int

	// HTTP status that should be used for the associated error code.
	HTTP int

	// External (user) facing error text.
	Ext string
}

var unknownCoder = &ErrCode{C: 500, HTTP: 500, Ext: "系统异常"}

// Code returns the integer code of ErrCode.
func (coder ErrCode) Code() int {
	return coder.C
}

// String implements stringer. String returns the external error message,
// if any.
func (coder ErrCode) String() string {
	return coder.Ext
}

// HTTPStatus returns the associated HTTP status code, if any. Otherwise,
// returns 200.
func (coder ErrCode) HTTPStatus() int {
	if coder.HTTP == 0 {
		return http.StatusInternalServerError
	}

	return coder.HTTP
}

// MustRegister register a user define error code.
// It will panic when the same Code already exist.
func MustRegister(code int, httpStatus int, msg string) {
	if code == 0 {
		panic("code '0' is reserved by 'github.com/marmotedu/errors' as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("code: %d already exist", code))
	}

	codes[code] = &ErrCode{C: code, HTTP: httpStatus, Ext: msg}
}

// ParseCoder parse any error into *withCode.
// nil error will return nil direct.
// None withStack error will be parsed as ErrUnknown.
func ParseCoder(err error) Coder {
	if err == nil {
		return nil
	}
	var codeErr *withCode

	if ok := errors.As(err, &codeErr); ok {
		if coder, ok := codes[codeErr.code]; ok {
			return coder
		}
	}

	return unknownCoder
}

// IsCode reports whether any error in codeerr's chain contains the given error code.
func IsCode(err error, code int) bool {
	if v, ok := err.(*withCode); ok {
		if v.code == code {
			return true
		}
	}
	return false
}

func init() {
	codes[unknownCoder.Code()] = unknownCoder
}
