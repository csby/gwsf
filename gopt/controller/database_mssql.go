package controller

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"net"
	"sort"
	"strings"
	"time"
)

func (s *Database) GetSqlServerInstances(ctx gtype.Context, ps gtype.Params) {
	argument := &gtype.DatabaseServer{
		Port: "1434",
	}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Host) <= 0 {
		ctx.Error(gtype.ErrInput, "主机(host)为空")
		return
	}

	instances, err := s.getSqlServerInstances(argument.Host, argument.Port)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	results := make(gtype.SqlServerInstanceCollection, 0)
	for name, value := range instances {
		if len(name) < 1 {
			continue
		}

		val, ok := value["tcp"]
		if !ok {
			continue
		}
		result := &gtype.SqlServerInstance{
			Name: name,
			Port: val,
		}

		val, ok = value["Version"]
		if ok {
			result.Version = val
		}

		results = append(results, result)
	}

	sort.Sort(results)
	ctx.Success(results)
}

func (s *Database) GetSqlServerInstancesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "数据库")
	function := catalog.AddFunction(method, uri, "获取SQL Server实例列表")
	function.SetNote("获取SQL Server实例名称及监听端口等信息")
	function.SetInputJsonExample(&gtype.DatabaseServer{
		Host: "127.0.0.1",
		Port: "1434",
	})
	function.SetOutputDataExample([]*gtype.SqlServerInstance{
		{
			Name: "MSSQLSERVER",
			Port: "1433",
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Database) getSqlServerInstances(host, port string) (map[string]map[string]string, error) {
	dialer := &net.Dialer{KeepAlive: 10 * time.Second}
	conn, err := dialer.Dial("udp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(dialer.KeepAlive))

	_, err = conn.Write([]byte{3})
	if err != nil {
		return nil, err
	}

	var resp = make([]byte, 16*1024-1)
	read, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}

	return s.parseSqlServerInstances(resp[:read]), nil
}

func (s *Database) parseSqlServerInstances(msg []byte) map[string]map[string]string {
	results := map[string]map[string]string{}
	if len(msg) > 3 && msg[0] == 5 {
		out_s := string(msg[3:])
		tokens := strings.Split(out_s, ";")
		instdict := map[string]string{}
		got_name := false
		var name string
		for _, token := range tokens {
			if got_name {
				instdict[name] = token
				got_name = false
			} else {
				name = token
				if len(name) == 0 {
					if len(instdict) == 0 {
						break
					}
					results[strings.ToUpper(instdict["InstanceName"])] = instdict
					instdict = map[string]string{}
					continue
				}
				got_name = true
			}
		}
	}
	return results
}
