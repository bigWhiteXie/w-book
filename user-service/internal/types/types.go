// Code generated by goctl. DO NOT EDIT.
package types

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInfo struct {
	Token string `json:"token"`
}

type SignReq struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserInfoReq struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
