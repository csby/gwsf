package galipay

type LoginPage struct {
	Url    string `json:"url" note:"登录页面URL"`
	State  string `json:"state" note:"状态信息"`
	QRCode string `json:"qrCode" note:"登录页面URL二维码"`
}
