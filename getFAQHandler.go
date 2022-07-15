package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	proto "github.com/golang/protobuf/proto"
)

//獲取FAQ
type GetFAQHandler struct {
}

func (rh *GetFAQHandler) Init() {}
func (rh *GetFAQHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_GetFAQID)
}

func (rh *GetFAQHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GetFAQ{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.GetFAQ)
	platformid := mm.GetPlatformID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intPlatformID", platformid)
	ps.AddVarcharSizeOutput("chvContent", "", 8000)

	proc := db.NewProcedure("PrPs_GetFAQ", ps)
	ret, err := d.ExecProc(proc)

	faqContent := ""
	if err != nil {
		ctx.Error("執行存儲過程PrPs_GetFAQ錯誤, %v", err)
	} else {
		faqContent = ret.GetOutputParamValue("chvContent").(string)
		ctx.Info("返回結果：%v", faqContent)
	}

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_GetFAQRetID), nil)
	retdata := new(netproto.FAQDetail)
	retdata.Content = proto.String(faqContent)
	retmsg.MsgData = retdata

	return []*msg.Message{retmsg}
}
