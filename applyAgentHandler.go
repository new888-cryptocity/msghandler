package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

//客戶端申請代理的請求處理
type ApplyAgentHandler struct {
}

func (rh *ApplyAgentHandler) Init() {}
func (rh *ApplyAgentHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_ApplyAgentID)
}

func (rh *ApplyAgentHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ApplyAgent{}
		return mm
	})

	if rpc == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.ApplyAgent)
	areaid := mm.GetAreaID()
	name := mm.GetName()
	tel := mm.GetTel()
	qq := mm.GetQQ()
	wxno := mm.GetWXNo()
	memo := mm.GetMemo()
	userid := rpc.GetUserID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intAreaID", areaid)
	ps.AddVarcharInput("chvName", name, 32)
	ps.AddVarcharInput("chvTel", tel, 20)
	ps.AddVarcharInput("chvQQ", qq, 20)
	ps.AddVarcharInput("chvWXNo", wxno, 20)
	ps.AddVarcharInput("chvMemo", memo, 256)
	ps.AddVarcharOutput("chvErrMsg", "")
	proc := db.NewProcedure("PrPs_ApplyAgent", ps)
	ret, err := d.ExecProc(proc)

	var code int32 = 0
	retmessage := ""
	if err != nil {
		ctx.Error("執行存儲過程PrPs_ApplyAgent錯誤, %v", err)
		code = 0
		retmessage = "服務器錯誤"
	} else {
		if ret.GetReturnValue() == 1 {
			code = 1
		}
		retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_ApplyAgentRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
