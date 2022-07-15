package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//獲取用戶郵件列錶
type GetMailDetailHandler struct {
}

func (rh *GetMailDetailHandler) Init() {}
func (rh *GetMailDetailHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_GetMailDetailID)
}

func (rh *GetMailDetailHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.RequestMailDetail{}
		return mm
	})

	if rpc == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}
	mm := rmm.(*netproto.RequestMailDetail)

	userid := rpc.GetUserID()
	cer := rpc.GetCer()
	id := mm.GetID()
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intMsgID", id)
	ps.AddVarcharInput("chvCer", cer, 32)
	proc := db.NewProcedure("PrPs_Mail_GetMailByID", ps)
	dbret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_GetMailDetailRetID), nil)
	retdata := new(netproto.MailDetail)
	if err != nil {
		strMsg := fmt.Sprintf("執行存儲過程PrPs_Mail_GetMailByID錯誤, %v", err)
		ctx.Error(strMsg)
		return []*msg.Message{GetDBRetMsg(0)}
	} else {
		if dbret.GetRetTableCount() > 0 && len(dbret[0].Rows) > 0 {
			retdata.ID = proto.Int32(int32(dbret[0].GetValueByColName(0, "ID").(int64)))
			retdata.Content = proto.String(dbret[0].GetValueByColName(0, "Content").(string))
			retdata.Title = proto.String(dbret[0].GetValueByColName(0, "title").(string))
			retdata.IsRead = proto.Bool(dbret[0].GetValueByColName(0, "readtype").(int64) == 1)
			retdata.SendTime = proto.String(dbret[0].GetValueByColName(0, "SendTime").(string))
		}
	}

	retmsg.MsgData = retdata

	return []*msg.Message{retmsg}
}
