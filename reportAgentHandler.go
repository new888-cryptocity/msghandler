package msghandler

import (
	"regexp"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//舉報代理
type ReportAgentHandler struct {
}

func (rh *ReportAgentHandler) Init() {}
func (rh *ReportAgentHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_ReportAgentID)
}

func (rh *ReportAgentHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ReportAgent{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.ReportAgent)
	areaid := mm.GetAreaID()
	agentname := mm.GetAgentName()
	memo := mm.GetContent()

	var code int32 = 0
	retmessage := ""
	if tret, err := regexp.Match(".*[!-+<>(){}=!;:&'\"]+.*", []byte(memo)); tret && err == nil {
		ctx.Error("參數包含非法字符, %v", memo)
		code = 0
		retmessage = "參數包含非法字符"
	} else {
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intAreaID", areaid)
		ps.AddVarcharInput("chvAgentName", agentname, 32)
		ps.AddVarcharInput("chvMemo", memo, 512)
		ps.AddVarcharOutput("chvErrMsg", "")
		proc := db.NewProcedure("PrPs_AddReportAgentLog", ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			ctx.Error("執行存儲過程PrPs_AddReportAgentLog錯誤, %v", err)
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
	retmsg.SClassID = int32(netproto.HallMsgClassID_ReportAgentRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
