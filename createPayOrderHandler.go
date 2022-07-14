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

//修改密碼
type CreatePayOrderHandler struct {
}

func (rh *CreatePayOrderHandler) Init() {}
func (rh *CreatePayOrderHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_CreatePayOrderID)
}

func (rh *CreatePayOrderHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.CreatePayOrder{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.CreatePayOrder)
	paytypeid := mm.GetPayTypeID()
	amount := mm.GetAmount()
	userid := rpcinfo.GetUserID()
	ip := rpcinfo.GetIPAddress()

	ctx.Info("客戶端請求創建內購訂單: %v %v %v %v", paytypeid, userid, amount, ip)
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intPayTypeID", paytypeid)
	ps.AddIntInput("intUserID", userid)
	ps.AddBigIntInput("lngAmount", amount)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddIntOutput("intOrderID", 0)
	ps.AddVarcharOutput("chvErrMsg", "")
	proc := db.NewProcedure("PrPs_Pay_CreateOrder", ps)
	ret, err := d.ExecProc(proc)

	var code int32 = 0
	var orderid int32 = 0
	retmessage := ""

	if err != nil {
		ctx.Error("執行存儲過程PrPs_Pay_CreateOrder錯誤, %v", err)
		code = 0
		retmessage = "服務器錯誤"
	} else {
		if ret.GetReturnValue() == 1 {
			code = 1
			orderid = int32(ret.GetOutputParamValue("intOrderID").(int64))
		}
		retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
	}

	retData := &netproto.CreatePayOrderRet{}
	retData.Message = proto.String(retmessage)
	retData.Code = proto.Int32(code)
	retData.OrderID = proto.Int32(orderid)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_CreatePayOrderRetID)
	retmsg.MsgData = retData

	return []*msg.Message{retmsg}
}
