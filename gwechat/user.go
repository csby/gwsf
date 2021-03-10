package gwechat

import "encoding/json"

type User struct {
	OpenId     string   `json:"openid" note:"用户的标识，对当前开发者帐号唯一"`
	UnionId    string   `json:"unionid" note:"用户统一标识。针对一个微信开放平台帐号下的应用，同一用户的unionid是唯一的"`
	NickName   string   `json:"nickname" note:"用户昵称"`
	Sex        int      `json:"sex" note:"用户性别，1为男性，2为女性"`
	Country    string   `json:"country" note:"国家，如中国为CN"`
	Province   string   `json:"province" note:"用户个人资料填写的省份"`
	City       string   `json:"city" note:"用户个人资料填写的城市"`
	HeadImgUrl string   `json:"headimgurl" note:"用户头像"`
	Privilege  []string `json:"privilege" note:"用户特权信息, 如微信沃卡用户为（chinaunicom）"`

	ErrCode int    `json:"errcode" note:"错误代码"`
	ErrMsg  string `json:"errmsg" note:"错误消息"`
}

func (s *User) unmarshal(v []byte) error {
	return json.Unmarshal(v, s)
}
