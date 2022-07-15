package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	"github.com/golang/protobuf/proto"
)

//寫遊戲記錄詳情
type AddBetDetailLogHandler struct {
}

func (rh *AddBetDetailLogHandler) Init() {}
func (rh *AddBetDetailLogHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_GS_BetDetailLog)
}

func (rh *AddBetDetailLogHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GS_BetDetailLog{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.GS_BetDetailLog)
	userId := mm.GetUserId()
	groupid := mm.GetGroupID()
	roundId := mm.GetRoundId()
	gameId := mm.GetGameId()
	betAmount := mm.GetBetAmount()
	WinAmount := mm.GetWinAmount()
	result := mm.GetResult()
	tableNo := mm.GetTableNo()
	betDetail := mm.GetBetDetail()
	resultDetail := mm.GetResultDetail()
	balanceBefore := mm.GetBalanceBefore()
	balanceAfter := mm.GetBalanceAfter()

	d := GetDealLogDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("userId", userId)
	ps.AddVarcharInput("roundId", roundId, 100)
	ps.AddIntInput("gameId", gameId)
	ps.AddBigIntInput("betAmount", betAmount)
	ps.AddBigIntInput("winAmount", WinAmount)
	ps.AddVarcharInput("result", result, 255)
	ps.AddIntInput("groupId", groupid)
	ps.AddIntInput("tableNo", tableNo)
	ps.AddVarcharInput("betDetail", betDetail, 255)
	ps.AddVarcharInput("resultDetail", resultDetail, 255)
	ps.AddBigIntInput("balanceBefore", balanceBefore)
	ps.AddBigIntInput("balanceAfter", balanceAfter)

	proc := db.NewProcedure("PrGs_AddBetDetailLog", ps)
	_, err := d.ExecProc(proc)

	if err != nil {
		ctx.Error("執行存儲過程PrGs_AddBetDetailLog錯誤, %v", err)
	}
	return nil
}
