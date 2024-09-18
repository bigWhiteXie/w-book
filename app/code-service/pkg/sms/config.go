package sms

type SmsConf struct {
	TC Tencent `json:",omitempty"`
}

type Tencent struct {
	SecretKey string `json:",omitempty"`
	SecretId  string `json:",omitempty"`
	Endpoint  string `json:",omitempty"`
	AppId     string `json:",omitempty"`
	SignName  string `json:",omitempty"`
	Region    string `json:",omitempty"`
}
