package galipay

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/boombuler/barcode/qr"
	"github.com/csby/gwsf/gtype"
	"image/png"
	"net/url"
)

const (
	urlAuth = "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm"
)

func GetLoginPage(appId, redirectUri, state string) (*LoginPage, error) {
	if len(appId) < 1 {
		return nil, fmt.Errorf("appId is empty")
	}
	if len(redirectUri) < 1 {
		return nil, fmt.Errorf("redirectUri is empty")
	}
	if len(state) < 1 {
		state = gtype.NewGuid()
	}

	info := &LoginPage{}
	info.State = state
	info.Url = fmt.Sprintf("%s?appid=%s&redirect_uri=%s&response_type=%s&&scope=%s&state=%s",
		urlAuth,
		appId,
		url.QueryEscape(redirectUri),
		"code",
		"snsapi_userinfo",
		state)

	code, err := qr.Encode(info.Url, qr.H, qr.Auto)
	if err == nil {
		var buf bytes.Buffer
		err = png.Encode(&buf, code)
		if err == nil {
			qrCode := base64.StdEncoding.EncodeToString(buf.Bytes())
			info.QRCode = fmt.Sprintf("data:image/png;base64,%s", qrCode)
		}
	}

	return info, err
}
