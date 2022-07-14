package msghandler

import (
	"regexp"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

type BindBankCardHandler struct {
}

func (rh *BindBankCardHandler) Init() {}
func (rh *BindBankCardHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_BindBankCardID)
}

func (rh *BindBankCardHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BankCardInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BankCardInfo)
	bankcardnumber := mm.GetBankCardNumber()
	bankcardname := mm.GetBankCardName()
	bankname := mm.GetBankName()
	NeedVcode := mm.GetNeedVcode()
	VCode := mm.GetVcode()

	userid := rpcinfo.GetUserID()
	cer := rpcinfo.GetCer()

	var code int32 = 0
	retmessage := ""
	if tret, err := regexp.Match(".*[^\\d]+.*", []byte(bankcardnumber)); tret && err == nil {
		ctx.Error("參數包含非法字符, %v", bankcardnumber)
		code = 0
		retmessage = "賬號包含非法字符"
	} else if tret, err := regexp.Match(".*[!-+<>(){}=!;:&'\"]+.*", []byte(bankcardname)); tret && err == nil {
		ctx.Error("參數包含非法字符, %v", bankcardname)
		code = 0
		retmessage = "名字包含非法字符"
	} else if tret, err := regexp.Match(".*[!-+<>(){}=!;:&'\"]+.*", []byte(bankname)); tret && err == nil {
		ctx.Error("參數包含非法字符, %v", bankname)
		code = 0
		retmessage = "銀行名包含非法字符"
	} else {
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharInput("chvCer", cer, 32)
		ps.AddVarcharInput("chvRealName", bankcardname, 64)
		ps.AddVarcharInput("chvBankAccount", bankcardnumber, 64)
		ps.AddVarcharInput("chvBankName", bankname, 64)

		ps.AddIntInput("intNeedVcode", NeedVcode)
		ps.AddVarcharInput("chvVcode", VCode, 10)

		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_BindBankAccount", ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			ctx.Error("執行存儲過程PrPs_BindBankAccount錯誤, %v", err)
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
	retmsg.SClassID = int32(netproto.HallMsgClassID_BindBankCardRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
