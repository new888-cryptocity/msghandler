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

//發送驗證碼
type SendPhoneVCodeHandler struct {
}

func (rh *SendPhoneVCodeHandler) Init() {}
func (rh *SendPhoneVCodeHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_SendPhoneVCodeID)
}

func (rh *SendPhoneVCodeHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SendPhoneVCode{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.SendPhoneVCode)
	tel := mm.GetTel()
	codetype := mm.GetCodeType()

	userid := rpcinfo.GetUserID()
	intPlatformIDClinet := mm.GetPlatformID()
	countryCode := mm.GetCountryCode()
	if countryCode == 0 {
		countryCode = 86
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	if tel != "" {
		ps.AddVarcharInput("chvTel", tel, 32)
	}
	ps.AddIntInput("tnyCodeType", codetype)
	ps.AddIntOutput("tnyCountdown", 0)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntInput("intPlatformIDClinet", intPlatformIDClinet)
	ps.AddIntInput("intCountryCode", countryCode)

	proc := db.NewProcedure("PrPs_User_SendSMSCode", ps)
	ret, err := d.ExecProc(proc)

	code := 0
	retmessage := ""
	var countdown int32 = 0
	if err != nil {
		ctx.Error("執行存儲過程PrPs_User_SendSMSCode錯誤, %v", err)
		code = 0
		retmessage = "server error" //"服務器錯誤"
	} else {
		code = int(ret.GetReturnValue())
		retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
		countdown = int32(ret.GetOutputParamValue("tnyCountdown").(int64))
	}

	coderet := &netproto.SendPhoneVCodeRet{}
	coderet.Code = proto.Int32(int32(code))
	coderet.Message = proto.String(retmessage)
	coderet.CountDown = proto.Int32(countdown)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_SendPhoneVCodeRetID)
	retmsg.MsgData = coderet

	return []*msg.Message{retmsg}
}
