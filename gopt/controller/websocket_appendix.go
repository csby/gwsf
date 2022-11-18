package controller

import (
	"github.com/csby/gwsf/gtype"
)

type appendix struct {
	impl gtype.Appendix
}

func (s *appendix) Add(value interface{}, name, note string, example interface{}) {
	if s.impl != nil {
		s.impl.Add(value, name, note, example)
	}
}

func (s *appendix) AddItem(item gtype.AppendixItem) {
	if s.impl != nil {
		s.impl.AddItem(item)
	}
}

type appendixItem struct {
	value   interface{}
	name    string
	note    string
	example interface{}
}

func (s *appendixItem) Value() interface{} {
	return s.value
}

func (s *appendixItem) Name() string {
	return s.name
}

func (s *appendixItem) Note() string {
	return s.note
}

func (s *appendixItem) Example() interface{} {
	return s.example
}

func (s *appendixItem) Set(id int, example interface{}) *appendixItem {
	s.value = id
	s.name, s.note = s.getNameAndNote(id)
	s.example = &gtype.SocketMessage{
		ID:   id,
		Data: example,
	}

	return s
}

func (s *appendixItem) getNameAndNote(id int) (string, string) {
	switch id {
	case gtype.WSClusterNodeStatusChanged:
		return "WSClusterNodeStatusChanged", "集群节点状态改变"

	case gtype.WSHeartbeatConnected:
		return "WSHeartbeatConnected", "心跳检测已连接"
	case gtype.WSHeartbeatDisconnected:
		return "WSHeartbeatDisconnected", "心跳检测断开连接"

	case gtype.WSOptUserLogin:
		return "WSOptUserLogin", "用户登陆"
	case gtype.WSOptUserLogout:
		return "WSOptUserLogout", "用户注销"
	case gtype.WSOptUserOnline:
		return "WSOptUserOnline", "用户上线"
	case gtype.WSOptUserOffline:
		return "WSOptUserOffline", "用户下线"

	case gtype.WSSiteUpload:
		return "WSSiteUpload", "上传并发布应用网站"
	case gtype.WSRootSiteUploadFile:
		return "WSRootSiteUploadFile", "根站点-上传文件"
	case gtype.WSRootSiteDeleteFile:
		return "WSRootSiteDeleteFile", "根站点-删除文件"

	case gtype.WSNodeOnline:
		return "WSNodeOnline", "节点上线"
	case gtype.WSNodeOffline:
		return "WSNodeOffline", "节点下线"
	case gtype.WSNodeForwardTcpStart:
		return "WSNodeForwardTcpStart", "节点转发TCP开始"
	case gtype.WSNodeForwardTcpEnd:
		return "WSNodeForwardTcpEnd", "节点转发TCP结束"
	case gtype.WSNodeForwardUdpRequest:
		return "WSNodeForwardUdpRequest", "节点转发UDP请求"
	case gtype.WSNodeForwardUdpResponse:
		return "WSNodeForwardUdpResponse", "节点转发UDP响应"

	case gtype.WSNodeRegister:
		return "WSNodeRegister", "节点注册"
	case gtype.WSNodeRevoke:
		return "WSNodeRevoke", "节点注销"
	case gtype.WSNodeModify:
		return "WSNodeModify", "节点修改"
	case gtype.WSNodeInstanceOnline:
		return "WSNodeInstanceOnline", "节点实例上线"
	case gtype.WSNodeInstanceOffline:
		return "WSNodeInstanceOffline", "节点实例下线"

	case gtype.WSNodeOnlineStateChanged:
		return "WSNodeOnlineStateChanged", "节点在线状态改变"
	case gtype.WSNodeFwdInputListenSvcState:
		return "WSNodeFwdInputListenSvcState", "节点转发服务状态"
	case gtype.WSNodeFwdInputListenItemState:
		return "WSNodeFwdInputListenItemState", "节点转发项目状态"

	case gtype.WSNetworkThroughput:
		return "WSNetworkThroughput", "网络吞吐量"
	case gtype.WSCpuUsage:
		return "WSCpuUsage", "CPU使用率"
	case gtype.WSMemUsage:
		return "WSMemUsage", "内存使用率"

	case gtype.WSReviseProxyServiceStatus:
		return "WSReviseProxyServiceStatus", "反向代理服务状态信息"
	case gtype.WSReviseProxyConnectionOpen:
		return "WSReviseProxyConnectionOpen", "反向代理连接已打开"
	case gtype.WSReviseProxyConnectionShut:
		return "WSReviseProxyConnectionShut", "反向代理连接已关闭"

	case gtype.WSReviseProxyServerAdd:
		return "WSReviseProxyServerAdd", "反向代理添加服务器"
	case gtype.WSReviseProxyServerDel:
		return "WSReviseProxyServerDel", "反向代理删除服务器"
	case gtype.WSReviseProxyServerMod:
		return "WSReviseProxyServerMod", "反向代理修改服务器"

	case gtype.WSReviseProxyTargetAdd:
		return "WSReviseProxyTargetAdd", "反向代理添加目标地址"
	case gtype.WSReviseProxyTargetDel:
		return "WSReviseProxyTargetDel", "反向代理删除目标地址"
	case gtype.WSReviseProxyTargetMod:
		return "WSReviseProxyTargetMod", "反向代理修改目标地址"

	case gtype.WSReviseProxyTargetStatusChanged:
		return "WSReviseProxyTargetStatusChanged", "反向代理目标地址活动状态改变"

	case gtype.WSSvcStatusChanged:
		return "WSSvcStatusChanged", "服务状态改变"

	case gtype.WSCustomSvcAdded:
		return "WSCustomSvcAdded", "添加自定义服务"
	case gtype.WSCustomSvcUpdated:
		return "WSCustomSvcUpdated", "更新自定义服务"
	case gtype.WSCustomSvcDeleted:
		return "WSCustomSvcDeleted", "删除自定义服务"

	case gtype.WSTomcatAppAdded:
		return "WSTomcatAppAdded", "添加tomcat应用"
	case gtype.WSTomcatAppUpdated:
		return "WSTomcatAppUpdated", "更新tomcat应用"
	case gtype.WSTomcatAppDeleted:
		return "WSTomcatAppDeleted", "删除tomcat应用"

	case gtype.WSTomcatCfgAdded:
		return "WSTomcatCfgAdded", "添加tomcat配置"
	case gtype.WSTomcatCfgUpdated:
		return "WSTomcatCfgUpdated", "更新tomcat配置"
	case gtype.WSTomcatCfgDeleted:
		return "WSTomcatCfgDeleted", "删除tomcat配置"

	case gtype.WSNginxAppAdded:
		return "WSNginxAppAdded", "添加nginx应用"
	case gtype.WSNginxAppUpdated:
		return "WSNginxAppUpdated", "更新nginx应用"
	case gtype.WSNginxAppDeleted:
		return "WSNginxAppDeleted", "删除nginx应用"

	default:
		return "", ""
	}
}
