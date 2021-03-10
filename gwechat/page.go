package gwechat

type LoginPage struct {
	Url    string `json:"url" note:"登录页面URL"`
	QRCode string `json:"qrCode" note:"登录页面URL二维码"`
}
