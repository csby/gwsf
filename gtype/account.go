package gtype

type AccountInfo struct {
	Account string `json:"account" note:"账号"`
	Name    string `json:"name" note:"姓名"`
	BuiltIn bool   `json:"builtIn" note:"是否内置"`
}

type AccountCreate struct {
	Account  string `json:"account" required:"true" note:"账号"`
	Name     string `json:"name" note:"姓名"`
	Password string `json:"password" note:"密码"`
}

type AccountEdit struct {
	Account string `json:"account" required:"true" note:"账号"`
	Name    string `json:"name" note:"姓名"`
}

type AccountPasswordReset struct {
	Account  string `json:"account" required:"true" note:"账号"`
	Password string `json:"password" note:"密码"`
}

type AccountPasswordChange struct {
	OldPassword string `json:"oldPassword" note:"原密码"`
	NewPassword string `json:"newPassword" note:"新密码"`
}
