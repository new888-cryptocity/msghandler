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
type AddPerGameLogHandler struct {
}

func (rh *AddPerGameLogHandler) Init() {}
func (rh *AddPerGameLogHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_DealLogID)
}

func (rh *AddPerGameLogHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DealLog{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.DealLog)
	serverid := mm.GetServerID()
	groupid := mm.GetLianyunID()
	gamedata := mm.GetGameData()
	usergameret := mm.GetUserGameRet()
	begintime := mm.GetGameBeginTime()
	endtime := mm.GetGameEndTime()
	taxamount := mm.GetTaxAmount()
	tableNo := mm.GetTableNo()

	d := GetDealLogDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intServerID", serverid)
	ps.AddVarcharInput("chvGameData", gamedata, -1)
	ps.AddVarcharInput("chvUserGameRet", usergameret, -1)
	ps.AddVarcharInput("dtmGameBeginTime", begintime, 32)
	ps.AddVarcharInput("dtmGameEndTime", endtime, 32)
	ps.AddBigIntInput("lngTaxAmount", taxamount)
	ps.AddIntInput("intGroupID", groupid)
	ps.AddIntInput("tableNo", tableNo)
	proc := db.NewProcedure("PrGs_AddPerGameLog", ps)
	if serverid/100 == 31 { //休閒鬥地主
		proc = db.NewProcedure("PrXXGs_AddPerGameLog", ps)
	}
	_, err := d.ExecProc(proc)

	var retcode int32 = 1
	for i := 0; err != nil && i < 100; i++ {
		ctx.Error("執行存儲過程PrGs_AddPerGameLog錯誤, %v", err)
		_, err = d.ExecProc(proc) //重試一次
		if err == nil {
			ctx.Error("第%v次重試執行存儲過程PrGs_AddPerGameLog成功,serverid：%v，lianyunid：%v，usergameret：%v，endtime：%v", i+1, serverid, groupid, usergameret, endtime)
			break
		}
	}

	if err != nil {
		ctx.Error("PrGs_AddPerGameLog失敗, serverid：%v，lianyunid：%v，endtime：%v", serverid, groupid, endtime)
		ctx.Error("/*【budan2_start需補單】*/ USE PerGameLog; %s \nGO\n/*【budan2_end需補單】*/", proc.GetSqlString())
		retcode = 0
	}

	return []*msg.Message{GetDBRetMsg(retcode)}

	return nil
}
