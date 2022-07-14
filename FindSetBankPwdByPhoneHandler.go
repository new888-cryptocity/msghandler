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

//找回銀行密碼
type FindSetBankPwdByPhoneHandler struct {
}

func (rh *FindSetBankPwdByPhoneHandler) Init() {}
func (rh *FindSetBankPwdByPhoneHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_FindSetBankPwdByPhoneID)
}

func (rh *FindSetBankPwdByPhoneHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.FindSetBankPwdByPhone{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.FindSetBankPwdByPhone)

	vcode := mm.GetVCode()
	passwd := mm.GetPassword()
	ip := rpcinfo.GetIPAddress()
	userid := rpcinfo.GetUserID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharInput("chvCode", vcode, 32)
	ps.AddVarcharInput("chvNewPwd", passwd, 32)
	ps.AddVarcharInput("chvIPAddress", ip, 32)
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrPs_User_FindSetBankPwdByPhone", ps)
	ret, err := d.ExecProc(proc)

	var code int32 = 0
	retmessage := ""
	if err != nil {
		ctx.Error("執行存儲過程PrPs_User_FindSetBankPwdByPhone錯誤, %v", err)
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
	retmsg.SClassID = int32(netproto.HallMsgClassID_FindSetBankPwdByPhoneRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
