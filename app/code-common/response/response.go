package response

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,optional"`
}

func Ok(data interface{}) *Response {
	return &Response{Code: 200, Msg: "ok", Data: data}
}

func Fail(code int, msg string) *Response {
	return &Response{Code: code, Msg: msg}
}
