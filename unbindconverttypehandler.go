package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

type UnbindConvertTypeHandler struct {
}

func (rh *UnbindConvertTypeHandler) Init() {}
func (rh *UnbindConvertTypeHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_UnbindConvertTypeID)
}

func (rh *UnbindConvertTypeHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UnbindConvertType{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.UnbindConvertType)
	cvttype := mm.GetCvttype()
	tel := mm.GetTel()
	vcode := mm.GetVCode()

	userid := rpcinfo.GetUserID()
	cer := rpcinfo.GetCer()
	countryCode := mm.GetCountryCode()
	if countryCode == 0 {
		countryCode = 86
	}

	var code int32 = 0
	retmessage := ""

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intCvtType", cvttype)
	ps.AddVarcharInput("chvCer", cer, 32)
	ps.AddVarcharInput("chvTel", tel, 64)
	ps.AddVarcharInput("chvVcode", vcode, 64)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntInput("intCountryCode", countryCode)

	proc := db.NewProcedure("PrPs_UnBindAliBankAccount", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		ctx.Error("執行存儲過程PrPs_UnBindAliBankAccount錯誤, %v", err)
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
	retmsg.SClassID = int32(netproto.HallMsgClassID_UnbindConvertTypeRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
