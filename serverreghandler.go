package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	proto "github.com/golang/protobuf/proto"
)

type ServerRegHandler struct {
}

func (h *ServerRegHandler) Init() {}
func (rh *ServerRegHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_ServerMgm) && sclsid == int32(netproto.ServerMgmClassID_GameServerRegister)
}

func (rh *ServerRegHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ServerRegInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.ServerRegInfo)

	serverid := mm.GetServerID()
	servertype := mm.GetServerType()
	groupid := mm.GetLianyunID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intServerID", serverid)
	ps.AddIntInput("intServerType", servertype)
	ps.AddIntInput("intGroupID", groupid)

	proc := db.NewProcedure("PrGs_ReRegister", ps)
	_, err := d.ExecProc(proc)

	var retcode int32 = 1
	if err != nil {
		ctx.Error("執行存儲過程PrGs_ServerReg錯誤, %v", err)
		retcode = 0
	}

	return []*msg.Message{GetDBRetMsg(retcode)}
}
