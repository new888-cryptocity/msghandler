package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"

	proto "github.com/golang/protobuf/proto"
)

type UserSingleControlHandler struct {
}

func (h *UserSingleControlHandler) Init() {}
func (h *UserSingleControlHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_GameRoom) && sclsid == int32(netproto.GameRoomClassID_GetUserSingleControl)
}

func (h *UserSingleControlHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UserSingleControlReq{}
		return mm
	})

	if rmm == nil {
		return nil
	}
	mm := rmm.(*netproto.UserSingleControlReq)

	serverid := mm.GetServerID()
	userid := mm.GetUserID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("sintServerID", serverid)
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrGs_UserSingleControl", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("執行存儲過程 PrGs_UserSingleControl 出錯%v", err)
		return nil
	}

	if ret.GetReturnValue() != 1 {
		mlog.Error("執行存儲過程 PrGs_UserSingleControl 出錯, errcode:%v", ret.GetReturnValue())
		return nil
	}
	if ret.GetRetTableCount() > 0 && len(ret[0].Rows) > 0 {
		cntl := new(netproto.UserSingleControlRes)
		cntl.UserID = proto.Int32(userid)
		cntl.GameWin = proto.Int64(ret[0].GetSingleValueInt64("GameWin"))
		cntl.ControlMoney = proto.Int64(ret[0].GetSingleValueInt64("ControlMoney"))
		cntl.ControlLevel = proto.Int32(ret[0].GetSingleValueInt32("ControlLevel"))

		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
		retmsg.SClassID = int32(netproto.GameRoomClassID_GetUserSingleControl)
		retmsg.MsgData = cntl
		return []*msg.Message{retmsg}
	}
	return nil
}
