package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	"github.com/golang/protobuf/proto"
)

type AddUserTaskGameEvent struct {
}

func (ag *AddUserTaskGameEvent) Init() {}
func (ag *AddUserTaskGameEvent) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_GameRoom) {
		return false
	}
	switch sclsid {
	case int32(netproto.GameRoomClassID_AddUserTaskGameEvent):
		return true
	}
	return false
}

func (ag *AddUserTaskGameEvent) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	switch m.SClassID {
	case int32(netproto.GameRoomClassID_AddUserTaskGameEvent):
		return ag.AddUserGameEvent(m)
	}
	return nil
}

func (ag *AddUserTaskGameEvent) AddUserGameEvent(m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.AddUserTaskGameEventReq{}
		return mm
	})
	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.AddUserTaskGameEventReq)
	userID := mm.GetUserID()
	gameID := mm.GetGameID()
	gameType := mm.GetGameType()
	eventID := mm.GetEventID()
	addValue := mm.GetAddValue()

	mlog.Warn("AddUserGameEvent userID:%d, gameID:%d, gameType:%d, eventID:%d, addValue:%d",
		userID, gameID, gameType, eventID, addValue)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userID)
	ps.AddIntInput("intGameID", gameID)
	ps.AddIntInput("intGameType", gameType)
	ps.AddIntInput("intGameBehavior", eventID)
	ps.AddIntInput("intAddValue", addValue)
	proc := db.NewProcedure("PrGs_AddUserTaskProgress", ps)
	_, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("執行存儲過程PrGs_AddUserTaskProgress失敗,err:%v", err)
	}
	return []*msg.Message{GetDBRetMsg(0)}
}
