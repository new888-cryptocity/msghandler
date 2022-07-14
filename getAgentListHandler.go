package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

//获取用户邮件列表
type GetAgentListHandler struct {
}

func (rh *GetAgentListHandler) Init() {}
func (rh *GetAgentListHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_GetAgentListID)
}

func (rh *GetAgentListHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.RequestAgentList{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.RequestAgentList)
	areaid := mm.GetAreaID()

	if rpcinfo == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intAreaID", areaid)
	proc := db.NewProcedure("PrPs_GetPayAgent", ps)
	dbret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_GetAgentListRetID), nil)
	retdata := new(netproto.AgentList)

	if err != nil {
		strMsg := fmt.Sprintf("执行存储过程PrPs_GetPayAgent错误, %v", err)
		ctx.Error(strMsg)
	} else {
		if dbret.GetRetTableCount() > 0 {
			tbData := dbret[0]
			for i := 0; i < len(tbData.Rows); i++ {
				agentDetail := new(netproto.AgentDeatil)

				agentDetail.Name = proto.String(tbData.GetValueByColName(i, "Name").(string))
				agentDetail.WXNo = proto.String(tbData.GetValueByColName(i, "WXNo").(string))
				agentDetail.QQ = proto.String(tbData.GetValueByColName(i, "QQ").(string))

				retdata.AgentList = append(retdata.AgentList, agentDetail)
			}
		}
	}

	retmsg.MsgData = retdata

	return []*msg.Message{retmsg}
}
