package gwechat

import "encoding/json"

type Token struct {
	AccessToken  string `json:"access_token" note:"接口调用凭证"`
	ExpiresIn    int64  `json:"expires_in" note:"access_token接口调用凭证超时时间，单位（秒）"`
	RefreshToken string `json:"refresh_token" note:"用户刷新access_token"`
	OpenId       string `json:"openid" note:"授权用户唯一标识"`
	Scope        string `json:"scope" note:"用户授权的作用域，使用逗号（,）分隔"`
	UnionId      string `json:"unionid" note:"当且仅当该网站应用已获得该用户的userinfo授权时有效"`

	ErrCode int    `json:"errcode" note:"错误代码"`
	ErrMsg  string `json:"errmsg" note:"错误消息"`
}

func (s *Token) unmarshal(v []byte) error {
	return json.Unmarshal(v, s)
}
