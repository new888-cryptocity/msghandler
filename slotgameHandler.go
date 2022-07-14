package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"

	"github.com/golang/protobuf/proto"
)

type SlotGameHandler struct {
}

func (rh *SlotGameHandler) Init() {}
func (rh *SlotGameHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_GameRoom) {
		return false
	}

	switch sclsid {
	case int32(netproto.GameRoomClassID_SlotGetGameProgressID),
		int32(netproto.GameRoomClassID_SlotSaveGameProgressID),
		int32(netproto.GameRoomClassID_SlotGetJackpotID),
		int32(netproto.GameRoomClassID_SlotUpdateJackpotID),
		int32(netproto.GameRoomClassID_SlotGetJackpotGroupID),
		int32(netproto.GameRoomClassID_SlotUpdateJackpotGroupID):

		return true
	}
	return false
}

func (rh *SlotGameHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.GameRoomClassID_SlotGetGameProgressID):
		return rh.OnSlotGetGameProgress(ctx, clt, m)
	case int32(netproto.GameRoomClassID_SlotSaveGameProgressID):
		return rh.OnSlotSaveGameProgress(ctx, clt, m)
	case int32(netproto.GameRoomClassID_SlotGetJackpotID):
		return rh.OnSlotGetJackpot(ctx, clt, m)
	case int32(netproto.GameRoomClassID_SlotUpdateJackpotID):
		return rh.OnSlotUpdateJackpot(ctx, clt, m)
	case int32(netproto.GameRoomClassID_SlotGetJackpotGroupID):
		return rh.OnSlotGetJackpotGroup(ctx, clt, m)
	case int32(netproto.GameRoomClassID_SlotUpdateJackpotGroupID):
		return rh.OnSlotUpdateJackpotGroup(ctx, clt, m)
	}
	return nil
}

func (rh *SlotGameHandler) OnSlotGetGameProgress(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SlotGetGameProgress{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_SlotGetGameProgressRetID)
	data := &netproto.SlotGetGameProgressRet{}
	retmsg.MsgData = data
	data.GameID = proto.Int32(0)
	data.UserID = proto.Int32(0)
	data.Version = proto.Int32(0)
	data.GameData = proto.String("")
	data.Money = proto.Int64(0)
	if rmm == nil {
		return []*msg.Message{retmsg}
	}

	mm := rmm.(*netproto.SlotGetGameProgress)
	d := GetDatabase()
	ps := db.NewSqlParameters()
	userid := mm.GetUserID()
	gameid := mm.GetGameID()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intGameID", gameid)

	data.UserID = proto.Int32(userid)
	data.GameID = proto.Int32(gameid)

	proc := db.NewProcedure("PrGs_Slot_GetGameProgress", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrGs_Slot_GetGameProgress error userid[%d] gameid[%d]", userid, gameid)
		return []*msg.Message{retmsg}
	}
	retval := ret.GetReturnValue()
	if retval > 0 && ret[0] != nil {
		info := ret[0]
		for i := 0; i < len(info.Rows); i++ {
			data.Version = proto.Int32(int32(info.GetValueByColName(i, "Version").(int64)))
			data.GameData = proto.String(info.GetValueByColName(i, "GameData").(string))
			data.Money = proto.Int64(int64(info.GetValueByColName(i, "Money").(int64)))
		}
	}
	return []*msg.Message{retmsg}
}

func (rh *SlotGameHandler) OnSlotSaveGameProgress(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SlotSaveGameProgress{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.SlotSaveGameProgress)
	d := GetDatabase()
	ps := db.NewSqlParameters()
	userid := mm.GetUserID()
	gameid := mm.GetGameID()
	gamestation := mm.GetGameStation()
	version := mm.GetVersion()
	gamedata := mm.GetGameData()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intGameID", gameid)
	ps.AddIntInput("intGameStation", gamestation)
	ps.AddIntInput("intVersion", version)
	ps.AddVarcharInput("chvGameData", gamedata, -1)
	proc := db.NewProcedure("PrGs_Slot_SaveGameProgress", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrGs_Slot_SaveGameProgress error userid[%d] gameid[%d]", userid, gameid)
		return []*msg.Message{GetDBRetMsg(0)}
	}
	retval := ret.GetReturnValue()
	if retval <= 0 {
		mlog.Error("PrGs_Slot_SaveGameProgress error userid[%d] gameid[%d]", userid, gameid)
		return []*msg.Message{GetDBRetMsg(0)}
	}
	return []*msg.Message{GetDBRetMsg(1)}
}

func (rh *SlotGameHandler) OnSlotGetJackpot(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SlotGetJackpot{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_SlotGetJackpotRetID)
	data := &netproto.SlotGetJackpotRet{}
	retmsg.MsgData = data
	data.GameID = proto.Int32(0)

	if rmm == nil {
		return []*msg.Message{retmsg}
	}

	mm := rmm.(*netproto.SlotGetJackpot)
	d := GetDatabase()
	gameid := mm.GetGameID()
	poolcnt := int(mm.GetPoolCount())
	poolgroupid := int32(0)
	data.GameID = proto.Int32(gameid)
	for i := 0; i < poolcnt; i += 2 {
		ps := db.NewSqlParameters()
		ps.AddIntInput("intGameID", gameid)
		ps.AddIntInput("intPoolGroupID", poolgroupid)
		proc := db.NewProcedure("PrGs_Slot_GetJackpotPool", ps)
		ret, err := d.ExecProc(proc)
		if err != nil {
			mlog.Error("PrGs_Slot_GetJackpotPool error gameid[%d] poolgroupid[%d]", gameid, poolgroupid)
			return []*msg.Message{retmsg}
		}
		retval := ret.GetReturnValue()
		if retval > 0 && ret[0] != nil {
			info := ret[0]
			for j := 0; j < len(info.Rows); j++ {
				jackpot0 := info.GetValueByColName(j, "Jackpot0").(int64)
				jackpot1 := info.GetValueByColName(j, "Jackpot1").(int64)
				data.Jackpots = append(data.Jackpots, jackpot0)
				if i+1 < poolcnt {
					data.Jackpots = append(data.Jackpots, jackpot1)
				}
			}
		}
		poolgroupid++
	}
	return []*msg.Message{retmsg}
}

func (rh *SlotGameHandler) OnSlotUpdateJackpot(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SlotUpdateJackpot{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_SlotGetJackpotRetID)
	data := &netproto.SlotGetJackpotRet{}
	retmsg.MsgData = data
	data.GameID = proto.Int32(0)

	if rmm == nil {
		return []*msg.Message{retmsg}
	}
	mm := rmm.(*netproto.SlotUpdateJackpot)
	d := GetDatabase()
	gameid := mm.GetGameID()
	data.GameID = proto.Int32(gameid)

	poolgroupid := int32(0)
	jackpots := mm.GetChangeJackpots()
	poolcnt := len(jackpots)
	for i := 0; i < poolcnt; i += 2 {
		jackpot0 := jackpots[i]
		jackpot1 := int64(0)
		if i+1 < poolcnt {
			jackpot1 = jackpots[i+1]
		}

		ps := db.NewSqlParameters()
		ps.AddIntInput("intGameID", gameid)
		ps.AddIntInput("intPoolGroupID", poolgroupid)
		ps.AddBigIntInput("lngJackpot0", jackpot0)
		ps.AddBigIntInput("lngJackpot1", jackpot1)

		proc := db.NewProcedure("PrGs_Slot_UpdateJackpotPool", ps)
		ret, err := d.ExecProc(proc)
		if err != nil {
			mlog.Error("PrGs_Slot_UpdateJackpotPool error gameid[%d] poolgroupid[%d] jackpot0[%d] jackpot1[%d]", gameid, poolgroupid, jackpot0, jackpot1)
			return []*msg.Message{retmsg}
		}

		retval := ret.GetReturnValue()
		if retval > 0 && ret[0] != nil {
			info := ret[0]
			for j := 0; j < len(info.Rows); j++ {
				jackpot0 := info.GetValueByColName(j, "Jackpot0").(int64)
				jackpot1 := info.GetValueByColName(j, "Jackpot1").(int64)
				data.Jackpots = append(data.Jackpots, jackpot0)
				if i+1 < poolcnt {
					data.Jackpots = append(data.Jackpots, jackpot1)
				}
				// mlog.Info("[PrGs_Slot_UpdateJackpotPool] Jackpot0[%d] Jackpot1[%d]", jackpot0, jackpot1)
			}
		}

		poolgroupid++
	}

	return []*msg.Message{retmsg}
}

func (rh *SlotGameHandler) OnSlotGetJackpotGroup(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SlotGetJackpotGroup{}
		return mm
	})
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_SlotGetJackpotGroupRetID)
	data := &netproto.SlotGetJackpotGroupRet{}
	retmsg.MsgData = data
	data.GameID = proto.Int32(0)
	data.GroupID = proto.Int32(0)
	if rmm == nil {
		return []*msg.Message{retmsg}
	}

	mm := rmm.(*netproto.SlotGetJackpotGroup)
	d := GetDatabase()
	gameid := mm.GetGameID()
	poolcnt := int(mm.GetPoolCount())
	poolgroupid := int32(0)
	groupid := mm.GetGroupID()
	data.GroupID = proto.Int32(groupid)
	data.GameID = proto.Int32(gameid)
	for i := 0; i < poolcnt; i += 2 {
		ps := db.NewSqlParameters()
		ps.AddIntInput("intGameID", gameid)
		ps.AddIntInput("intPoolGroupID", poolgroupid)
		ps.AddIntInput("intGroupID", groupid)
		proc := db.NewProcedure("PrGs_Slot_GetJackpotPoolGroup", ps)
		ret, err := d.ExecProc(proc)
		if err != nil {
			mlog.Error("PrGs_Slot_GetJackpotPool error gameid[%d] poolgroupid[%d]", gameid, poolgroupid)
			return []*msg.Message{retmsg}
		}
		retval := ret.GetReturnValue()
		if retval > 0 && ret[0] != nil {
			info := ret[0]
			for j := 0; j < len(info.Rows); j++ {
				jackpot0 := info.GetValueByColName(j, "Jackpot0").(int64)
				jackpot1 := info.GetValueByColName(j, "Jackpot1").(int64)
				data.Jackpots = append(data.Jackpots, jackpot0)
				if i+1 < poolcnt {
					data.Jackpots = append(data.Jackpots, jackpot1)
				}
			}
		}
		poolgroupid++
	}
	return []*msg.Message{retmsg}
}

func (rh *SlotGameHandler) OnSlotUpdateJackpotGroup(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.SlotUpdateJackpotGroup{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_SlotGetJackpotGroupRetID)
	data := &netproto.SlotGetJackpotGroupRet{}
	retmsg.MsgData = data
	data.GameID = proto.Int32(0)
	data.GroupID = proto.Int32(0)
	if rmm == nil {
		return []*msg.Message{retmsg}
	}
	mm := rmm.(*netproto.SlotUpdateJackpotGroup)
	d := GetDatabase()
	gameid := mm.GetGameID()
	groupid := mm.GetGroupID()

	data.GameID = proto.Int32(gameid)
	data.GroupID = proto.Int32(groupid)
	poolgroupid := int32(0)
	jackpots := mm.GetChangeJackpots()
	poolcnt := len(jackpots)
	for i := 0; i < poolcnt; i += 2 {
		jackpot0 := jackpots[i]
		jackpot1 := int64(0)
		if i+1 < poolcnt {
			jackpot1 = jackpots[i+1]
		}

		ps := db.NewSqlParameters()
		ps.AddIntInput("intGameID", gameid)
		ps.AddIntInput("intPoolGroupID", poolgroupid)
		ps.AddBigIntInput("lngJackpot0", jackpot0)
		ps.AddBigIntInput("lngJackpot1", jackpot1)
		ps.AddIntInput("intGroupID", groupid)

		proc := db.NewProcedure("PrGs_Slot_UpdateJackpotPoolGroup", ps)
		ret, err := d.ExecProc(proc)
		if err != nil {
			mlog.Error("PrGs_Slot_UpdateJackpotPool error gameid[%d] poolgroupid[%d] jackpot0[%d] jackpot1[%d]", gameid, poolgroupid, jackpot0, jackpot1)
			return []*msg.Message{retmsg}
		}

		retval := ret.GetReturnValue()
		if retval > 0 && ret[0] != nil {
			info := ret[0]
			for j := 0; j < len(info.Rows); j++ {
				jackpot0 := info.GetValueByColName(j, "Jackpot0").(int64)
				jackpot1 := info.GetValueByColName(j, "Jackpot1").(int64)
				data.Jackpots = append(data.Jackpots, jackpot0)
				if i+1 < poolcnt {
					data.Jackpots = append(data.Jackpots, jackpot1)
				}
				// mlog.Info("[PrGs_Slot_UpdateJackpotPool] Jackpot0[%d] Jackpot1[%d]", jackpot0, jackpot1)
			}
		}

		poolgroupid++
	}

	return []*msg.Message{retmsg}
}
