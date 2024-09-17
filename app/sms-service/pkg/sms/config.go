package sms

type SmsConf struct {
	TC Tencent `json`
}

type Tencent struct {
	SecretKey string `json`
	SecretId  string `json`
	Endpoint  string `json`
	AppId     string `json`
	SignName  string `json`
	Region    string `json`
}
