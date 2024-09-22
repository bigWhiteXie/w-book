package sms

type SmsConf struct {
	TC     Tencent `json:",omitempty"`
	Memory Memeory `json:",omitempty"`
}

type SmsCommonConfig struct {
	Name   string `json:","`
	Weight int    `json:","`
}

type Memeory struct {
	SmsCommonConfig
}
type Tencent struct {
	SmsCommonConfig

	SecretKey string `json:","`
	SecretId  string `json:","`
	Endpoint  string `json:",optional"`
	AppId     string `json:",omitempty"`
	SignName  string `json:",omitempty"`
	Region    string `json:",optional,omitempty"`
}
