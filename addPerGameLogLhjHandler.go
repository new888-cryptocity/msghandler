package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

type AddPerGameLogLhjHandler struct {
}

func (ah *AddPerGameLogLhjHandler) Init() {}
func (ah *AddPerGameLogLhjHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_AddGameLog)
}

func (ah *AddPerGameLogLhjHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GameLogDeal{}
		return mm
	})
	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}
	mm := rmm.(*netproto.GameLogDeal)
	userID := mm.GetUserID()
	gameID := mm.GetGameID()
	groupID := mm.GetGroupID()
	beginTime := mm.GetBeginTime()
	endTime := mm.GetEndTime()
	logPk := mm.GetLogPk()
	gameData := mm.GetGameData()

	d := GetDealLogDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userID)
	ps.AddIntInput("intGameID", gameID)
	ps.AddIntInput("intGroupID", groupID)
	ps.AddVarcharInput("dtmBeginTime", beginTime, 32)
	ps.AddVarcharInput("dtmEndTime", endTime, 32)
	ps.AddVarcharInput("chvLogPk", logPk, 64)
	ps.AddVarcharInput("chvGameData", gameData, -1)
	proc := db.NewProcedure("PrGs_AddPerGameLogLHJ", ps)
	_, err := d.ExecProc(proc)
	if err == nil {
		ctx.Info("執行存儲過程PrGs_AddPerGameLogLHJ成功,userID：%v，gameID：%v，groupID：%v，logPk：%v", userID, gameID, groupID, logPk)
	}
	var retCode int32 = 1
	for i := 0; err != nil && i < 10; i++ {
		ctx.Error("執行存儲過程PrGs_AddPerGameLogNew錯誤, %v", err)
		_, err = d.ExecProc(proc) //重試一次
		if err == nil {
			ctx.Error("第%v次重試執行存儲過程PrGs_AddPerGameLogLHJ成功,userID：%v，gameID：%v，groupID：%v，logPk：%v", i+1, userID, gameID, groupID, logPk)
			break
		}
	}
	if err != nil {
		ctx.Error("第%v次重試執行存儲過程PrGs_AddPerGameLogLHJ失敗,userID：%v，gameID：%v，groupID：%v，logPk：%v", userID, gameID, groupID, logPk)
		ctx.Error("/*【budan2_start需補單】*/ USE PerGameLog; %s \nGO\n/*【budan2_end需補單】*/", proc.GetSqlString())
		retCode = 0
	}
	return []*msg.Message{GetDBRetMsg(retCode)}
}
