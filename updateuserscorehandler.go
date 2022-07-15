package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	"github.com/golang/protobuf/proto"
)

type UpdateUserScoreHandler struct {
}

func (rh *UpdateUserScoreHandler) Init() {}
func (rh *UpdateUserScoreHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_UpdateUserScoreID)
}

func (rh *UpdateUserScoreHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UpdateUserScore{}
		return mm
	})

	if rpc == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.UpdateUserScore)
	ctx.Info("PrGs_UpdateUserMoney update score [game] %v", mm)
	msgid := rpc.GetQueueID()
	userid := mm.GetUserID()
	score := mm.GetScore()
	serverid := mm.GetServerID()
	taxamount := mm.GetTaxAmount()
	flag := mm.GetFlag()
	param := mm.GetParam()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intMsgID", msgid)
	ps.AddIntInput("intUserID", userid)
	ps.AddBigIntInput("lngUpdateAmount", score)
	ps.AddIntInput("intServerid", serverid)
	ps.AddBigIntInput("lngTaxAmount", taxamount)
	ps.AddIntInput("intFlag", flag)
	ps.AddIntInput("intParam", param)
	proc := db.NewProcedure("PrGs_UpdateUserMoney", ps)
	if serverid/100 == 31 { //休閒鬥地主
		proc = db.NewProcedure("PrXXGs_UpdateUserMoney", ps)
	}
	_, err := d.ExecProc(proc)

	var retcode int32 = 1
	for i := 0; err != nil && i < 100; i++ {
		ctx.Error("執行存儲過程PrGs_UpdateUserMoney錯誤, %v", err)
		_, err = d.ExecProc(proc) //重試一次
		if err == nil {
			ctx.Error("第%v次重試執行存儲過程PrGs_UpdateUserMoney成功,msgid: %d, userid: %d", i+1, msgid, userid)
			break
		}
	}

	if err != nil {
		ctx.Error("PrGs_UpdateUserMoney失敗,msgid: %d, userid: %d, score: %d, serverid: %d, taxamount: %d", msgid, userid, score, serverid, taxamount)
		ctx.Error("/*【budan1_start需補單】*/ USE CenterDB; %s \nGO\n/*【budan1_end需補單】*/", proc.GetSqlString())
		retcode = 0
	}

	return []*msg.Message{GetDBRetMsg(retcode)}
}
