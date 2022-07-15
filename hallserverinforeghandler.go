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

//遊戲服務器啓動
type HallServerInfoRegHandler struct {
}

func (rh *HallServerInfoRegHandler) Init() {}
func (rh *HallServerInfoRegHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_HallServerInfoID)
}

func (rh *HallServerInfoRegHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.HallServerInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.HallServerInfo)

	serverid := mm.GetServerID()
	status := mm.GetStatus()
	playercnt := mm.GetPlayerCnt()
	addr := mm.GetAddr()
	groupid := mm.GetLianyunID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("sintServerID", serverid)
	ps.AddIntInput("sintStatus", status)
	ps.AddIntInput("sintPlayerCnt", playercnt)
	ps.AddVarcharInput("chvAddr", addr, 64)
	ps.AddIntInput("intGroupID", groupid)
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrGs_HallInfoReg", ps)
	dbret, err := d.ExecProc(proc)

	//ctx.Debug("註冊大廳信息到db ServerID=%d, Addr=%s, Status=%d, LianyunID=%d, playercnt=%d",
	//	serverid, addr, status, groupid, playercnt)

	if err != nil {
		message := fmt.Sprintf("執行存儲過程PrGs_HallInfoReg錯誤, %v %v", err, dbret)
		ctx.Error(message)
	}

	return []*msg.Message{}
}
