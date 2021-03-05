package gtype

import "strings"

type ForwardId struct {
	ID string `json:"id" note:"转发ID"`
}

type ForwardRequest struct {
	ForwardId

	NodeInstanceID string `json:"nodeInstId" note:"结点实例ID"`
	TargetAddress  string `json:"targetAddr" note:"目标地址"`
	TargetPort     string `json:"targetPort" note:"目标端口"`
}

func (s *ForwardRequest) CopyTo(target *ForwardRequest) {
	if target == nil {
		return
	}

	target.ID = s.ID
	target.NodeInstanceID = s.NodeInstanceID
	target.TargetAddress = s.TargetAddress
	target.TargetPort = s.TargetPort
}

type ForwardInfo struct {
	ForwardId

	Time       DateTime `json:"time" note:"开始时间"`
	SourceNode *Node    `json:"sourceNode" note:"发起结点"`
	TargetNode *Node    `json:"targetNode" note:"目标结点"`
	TargetHost string   `json:"targetHost" note:"目标主机"`
}

type ForwardInfoArray []*ForwardInfo

func (s ForwardInfoArray) Len() int {
	return len(s)
}

func (s ForwardInfoArray) Less(i, j int) bool {
	if s[i].Time.After(s[j].Time) {
		return true
	}

	return false
}

func (s ForwardInfoArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ForwardInfoFilter struct {
	SourceName string `json:"sourceName" note:"发起结点名称"`
	SourceIP   string `json:"sourceIp" note:"发起结点IP地址"`
	TargetName string `json:"targetName" note:"目标结点名称"`
	TargetIP   string `json:"targetIp" note:"目标结点IP地址"`
	TargetHost string `json:"targetHost" note:"目标主机"`
}

func (s *ForwardInfoFilter) match(link *ForwardInfo) bool {
	if link == nil {
		return false
	}

	if len(s.SourceIP) > 0 {
		if link.SourceNode == nil {
			return false
		}
		if !strings.Contains(link.SourceNode.IP, s.SourceIP) {
			return false
		}
	}

	if len(s.SourceName) > 0 {
		if link.SourceNode == nil {
			return false
		}
		if !strings.Contains(link.SourceNode.Name, s.SourceName) {
			return false
		}
	}

	if len(s.TargetIP) > 0 {
		if link.TargetNode == nil {
			return false
		}
		if !strings.Contains(link.TargetNode.IP, s.TargetIP) {
			return false
		}
	}

	if len(s.TargetName) > 0 {
		if link.TargetNode == nil {
			return false
		}
		if !strings.Contains(link.TargetNode.Name, s.TargetName) {
			return false
		}
	}

	if len(s.TargetHost) > 0 {
		if !strings.Contains(link.TargetHost, s.TargetHost) {
			return false
		}
	}

	return true
}
