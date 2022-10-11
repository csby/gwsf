package gmodel

type NotifySubscription struct {
	ID   int  `json:"id" note:"消息标识"`
	Data bool `json:"data" note:"是否订阅: true-订阅; false-取消订阅"`
}
