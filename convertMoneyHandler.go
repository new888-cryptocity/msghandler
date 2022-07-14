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

//聯係客服
type ConvertMoneyHandler struct {
}

func (rh *ConvertMoneyHandler) Init() {}
func (rh *ConvertMoneyHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_ConvertMoneyID)
}

func (rh *ConvertMoneyHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ConvertMoney{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.ConvertMoney)
	amount := mm.GetAmount()
	cvttype := mm.GetCvttype()

	userid := rpcinfo.GetUserID()
	cer := rpcinfo.GetCer()
	ip := rpcinfo.GetIPAddress()

	var code int32 = 0
	retmessage := ""
	if amount <= 0 {
		ctx.Error("兌換金額不正確, %v", amount)
		code = 0
		retmessage = "兌換金額不正確"
	} else {
		procName := "PrPs_AddConvertLog"
		if cvttype == 2 { //is convert bank
			procName = "PrPs_AddBankConvertLog"
		}
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharInput("chvCer", cer, 32)
		ps.AddBigIntInput("lngAmount", amount)
		ps.AddVarcharInput("chvIP", ip, 15)
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure(procName, ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			ctx.Error("執行存儲過程%v錯誤, %v", procName, err)
			code = 0
			retmessage = "服務器錯誤"
		} else {
			if ret.GetReturnValue() == 1 {
				code = 1
			}
			retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
		}
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_ConvertMoneyRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
