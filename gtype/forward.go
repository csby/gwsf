package gtype

type Forward struct {
	RequestID         string `json:"requestId" note:"请求ID"`
	ResponseID        string `json:"responseId" note:"响应ID"`
	NodeCertificateID string `json:"nodeCertId" note:"节点证书ID"`
	NodeInstanceID    string `json:"nodeInstId" note:"节点实例ID"`
	TargetAddress     string `json:"targetAddr" note:"目标地址"`
	TargetPort        string `json:"targetPort" note:"目标端口"`
}

func (s *Forward) CopyTo(target *Forward) {
	if target == nil {
		return
	}

	target.RequestID = s.ResponseID
	target.ResponseID = s.ResponseID
	target.NodeCertificateID = s.NodeCertificateID
	target.NodeInstanceID = s.NodeInstanceID
	target.TargetAddress = s.TargetAddress
	target.TargetPort = s.TargetPort
}
