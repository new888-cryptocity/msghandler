package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"

	proto "github.com/golang/protobuf/proto"
)

//用戶離開廣場
type UserLeaveHallHandler struct {
}

func (h *UserLeaveHallHandler) Init() {}
func (h *UserLeaveHallHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_UserLogoutID)
}

func (h *UserLeaveHallHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UserLogout{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.UserLogout)

	ctx.Info("user logout hall.%v", mm)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", mm.GetUserID())
	ps.AddIntInput("sintServerID", rpc.GetRouteServerID())
	ps.AddVarcharInput("chvCer", mm.GetCer(), 32)

	proc := db.NewProcedure("PrPs_UserLeave", ps)
	_, err := d.ExecProc(proc)

	var retcode int32 = 1
	if err != nil {
		ctx.Error("執行存儲過程PrPs_UserLeave錯誤, %v", err)
		retcode = 0
	}

	return []*msg.Message{GetDBRetMsg(retcode)}

}
