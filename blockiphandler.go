package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//封鎖Ip
type BlockIPHandler struct {
}

func (rh *BlockIPHandler) Init() {}
func (rh *BlockIPHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_BlockIPID)
}

func (rh *BlockIPHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BlockIP{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BlockIP)
	ip := mm.GetIPAddress()
	sec := mm.GetSec()
	memo := mm.GetMemo()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddIntInput("intSec", sec)
	ps.AddVarcharInput("chvMemo", memo, 64)

	proc := db.NewProcedure("PrSys_BlockIP", ps)
	_, err := d.ExecProc(proc)

	if err != nil {
		ctx.Error("執行存儲過程PrSys_BlockIP錯誤, %v", err)
	}

	return []*msg.Message{GetDBRetMsg(0)}
}
