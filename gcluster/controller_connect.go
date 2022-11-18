package gcluster

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"strconv"
)

func (s *Controller) NodeConnect(ctx gtype.Context, ps gtype.Params) {
	instanceIndex := ctx.Query("index")
	if len(instanceIndex) < 1 {
		ctx.Error(gtype.ErrInput, fmt.Errorf("instance index is empty"))
		return
	}
	instanceValue, err := strconv.Atoi(instanceIndex)
	if err != nil {
		ctx.Error(gtype.ErrInput, fmt.Errorf("instance index invalid: %s", err.Error()))
		return
	}
	if instanceValue < 1 || instanceValue > 9 {
		ctx.Error(gtype.ErrInput, fmt.Errorf("instance index %d out of range [1, 9]", instanceValue))
		return
	}
	index := uint64(instanceValue)
	if index == s.cfg.Cluster.Index {
		ctx.Error(gtype.ErrInput, fmt.Errorf("instance index %d is same as itself", index))
		return
	}
	version := ctx.Query("version")

	conn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		s.LogError("cluster socket connect fail:", err)
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	defer conn.Close()

	instance := s.getInstance(index)
	if instance == nil {
		ctx.Error(gtype.ErrInput, fmt.Errorf("instance index %d is invalid", index))
		return
	}

	connection := instance.In
	if connection == nil {
		ctx.Error(gtype.ErrInternal, "the connection is nil")
		return
	}

	ctx.FireAfterInput()

	connection.OnConnect(conn, version)
	if connection.Connected() {
		ctx.Error(gtype.ErrInternal, fmt.Errorf("instance index %d has connected", index))
		return
	}
}

func (s *Controller) NodeConnectDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "加入集群")
	function := catalog.AddFunction(method, uri, "连接")
	function.SetNote("接收或发送节点的同步信息，该接口保持阻塞至连接关闭")
	function.AddInputQuery(true, "index", "实例Index，有效值：1～9", "")
	function.AddInputQuery(false, "version", "实例版本号", "")
	function.SetInputExample(&gtype.SocketMessage{ID: 0})
	function.SetInputFormat(gtype.ArgsFmtJson)
	function.SetOutputExample(&gtype.SocketMessage{ID: 0})
	function.SetOutputFormat(gtype.ArgsFmtJson)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
