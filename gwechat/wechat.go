package gwechat

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/boombuler/barcode/qr"
	"github.com/csby/gwsf/gclient"
	"github.com/csby/gwsf/gtype"
	"image/png"
	"io"
	"net/url"
	"sort"
	"strings"
)

const (
	urlToken        = "https://api.weixin.qq.com/sns/oauth2/access_token"
	urlTokenRefresh = "https://api.weixin.qq.com/sns/oauth2/refresh_token"
	urlUserInfo     = "https://api.weixin.qq.com/sns/userinfo"
	urlAuth         = "https://open.weixin.qq.com/connect/oauth2/authorize"
)

// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
func GetAccessToken(appId, secret, code string) (*Token, error) {
	uri := fmt.Sprintf("%s?appId=%s&secret=%s&code=%s&grant_type=%s",
		urlToken,
		appId, secret, code, "authorization_code")

	client := gclient.Http{}
	output, _, rc, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	if rc != 200 {
		return nil, fmt.Errorf("http error, code=%d", rc)
	}

	info := &Token{}
	err = info.unmarshal(output)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
func RefreshAccessToken(appId, refreshToken string) (*Token, error) {
	uri := fmt.Sprintf("%s?appId=%s&grant_type=%s&refresh_token=%s",
		urlTokenRefresh,
		appId, "refresh_token", refreshToken)

	client := gclient.Http{}
	output, _, rc, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	if rc != 200 {
		return nil, fmt.Errorf("http error, code=%d", rc)
	}

	info := &Token{}
	err = info.unmarshal(output)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
func GetUserInfo(openId, accessToken string) (*User, error) {
	uri := fmt.Sprintf("%s?access_token=%s&openid=%s",
		urlUserInfo,
		accessToken, openId)

	client := gclient.Http{}
	output, _, rc, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	if rc != 200 {
		return nil, fmt.Errorf("http error, code=%d", rc)
	}

	info := &User{}
	err = info.unmarshal(output)
	if err != nil {
		return nil, err
	}

	return info, nil
}

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

	code, err := qr.Encode(info.Url, qr.M, qr.Auto)
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

func SignUrl(token string, params ...string) string {
	if len(token) < 1 {
		return ""
	}
	count := len(params)
	if count < 1 {
		return ""
	}
	content := make([]string, 0)
	content = append(content, token)
	for index := 0; index < count; index++ {
		content = append(content, params[index])
	}

	sort.Strings(content)
	sha := sha1.New()

	_, err := io.WriteString(sha, strings.Join(content, ""))
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", sha.Sum(nil))
}
