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

//找回賬號密碼
type FindSetPwdByPhoneHandler struct {
}

func (rh *FindSetPwdByPhoneHandler) Init() {}
func (rh *FindSetPwdByPhoneHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_FindSetPwdByPhoneID)
}

func (rh *FindSetPwdByPhoneHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.FindSetPwdByPhone{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.FindSetPwdByPhone)

	tel := mm.GetTel()
	vcode := mm.GetVCode()
	passwd := mm.GetPassword()
	ip := rpcinfo.GetIPAddress()
	countryCode := mm.GetCountryCode()
	if countryCode == 0 {
		countryCode = 86
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvTel", tel, 32)
	ps.AddVarcharInput("chvCode", vcode, 32)
	ps.AddVarcharInput("chvNewPwd", passwd, 32)
	ps.AddVarcharInput("chvIPAddress", ip, 32)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntInput("intCountryCode", countryCode)

	proc := db.NewProcedure("PrPs_User_FindSetPwdByPhone", ps)
	ret, err := d.ExecProc(proc)

	var code int32 = 0
	retmessage := ""
	if err != nil {
		ctx.Error("執行存儲過程PrPs_User_FindSetPwdByPhone錯誤, %v", err)
		code = 0
		retmessage = "服務器錯誤"
	} else {
		code = int32(ret.GetReturnValue())
		retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_FindSetPwdByPhoneRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
