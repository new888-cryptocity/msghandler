package msghandler

import (
	"666.com/gameserver/dbserver/src/dal/utility"
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

const POINTSRACE_ACTIVITY_ID int32 = 1

// 積分賽
type PointsRaceHandler struct {
}

func (ph *PointsRaceHandler) Init() {

}

func (ph *PointsRaceHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_Hall) {
		return false
	}
	switch netproto.HallMsgClassID(sclsid) {
	case netproto.HallMsgClassID_DZPKHALL_PointsRaceActivityConfig,
		netproto.HallMsgClassID_DZPKHALL_PointsRaceRank:
		return true
	}

	return false
}

func (ph *PointsRaceHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[PointsRaceHandler][OnMessage] %d %d", m.BClassID, m.SClassID)
	switch netproto.HallMsgClassID(m.SClassID) {
	case netproto.HallMsgClassID_DZPKHALL_PointsRaceActivityConfig:
		return ph.loadActivityConfig(ctx, clt, m)
	case netproto.HallMsgClassID_DZPKHALL_PointsRaceRank:
		return ph.loadRank(ctx, clt, m)
	}

	return nil
}

// 讀取活動配置
func (ph *PointsRaceHandler) loadActivityConfig(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, ok := m.MsgData.(*rpcmsg.RPCMessage)
	if !ok {
		mlog.Error("[PointsRaceHandler][loadActivityConfig] Error! RPC Parser Failed!")
		return nil
	}
	d := GetDatabase()
	sp := "DZPKHALL_PrPs_PointsRace_LoadActivityConfig"
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", *rpc.RPCInfo.UserID)
	ps.AddIntInput("intID", POINTSRACE_ACTIVITY_ID)
	ps.AddIntInput("intGroupID", 0)
	proc := db.NewProcedure(sp, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("[PointsRaceHandler][loadActivityConfig] Error! " + err.Error())
		return nil
	}

	data := utility.ParserPointsRaceActivityConfig(ret)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_PointsRaceActivityConfigRet)
	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

// 取得排行榜
func (ph *PointsRaceHandler) loadRank(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DZPKHALLPointRaceReqRanking{}
		return mm
	})
	mm := rmm.(*netproto.DZPKHALLPointRaceReqRanking)
	if mm == nil {
		return nil
	}

	d := GetDatabase()
	sp := "DZPKHALL_PrPs_PointsRace_GetRanking"
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", mm.GetUserID())
	ps.AddIntInput("intRankStart", mm.GetRankStart())
	ps.AddIntInput("intRankEnd", mm.GetRankEnd())
	ps.AddIntInput("intConfigID", mm.GetActivityID())
	proc := db.NewProcedure(sp, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("[PointsRaceHandler][DZPKHALL_PrPs_PointsRace_GetRanking] Error! " + err.Error())
		return nil
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_PointsRaceRankRet)
	data := utility.ParserPointsRaceRanking(ret)
	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

func (ph *PointsRaceHandler) getActivitConfig(userID int32) (int32, error) {
	d := GetDatabase()
	sp := "DZPKHALL_PrPs_PointsRace_GetActivitConfig"
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userID)
	ps.AddIntOutput("intConfigID", 0)
	proc := db.NewProcedure(sp, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("[PointsRaceHandler][DZPKHALL_PrPs_PointsRace_GetActivitConfig] Error! " + err.Error())
		return 0, err
	}
	retVal := ret.GetReturnValue()
	if retVal != 1 {
		mlog.Info("getActivitConfig return value = %d", retVal)
	}
	configID := ret.GetOutputParamValue("intConfigID").(int64)
	return int32(configID), nil
}

func (ph *PointsRaceHandler) updatePointsRaceRanking() error {
	d := GetDatabase()
	sp := "DZPKHALL_PrPs_PointsRace_GetAllUsersPoint"
	ps := db.NewSqlParameters()
	ps.AddIntInput("intActivityID", 1)
	ps.AddIntInput("intGroupID", 70001)
	proc := db.NewProcedure(sp, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("[PointsRaceHandler][DZPKHALL_PrPs_PointsRace_GetAllUsersPoint] Error! " + err.Error())
		return err
	}
	retVal := ret.GetReturnValue()
	if retVal != 1 {
		mlog.Info("updatePointsRaceRanking return value = %d", retVal)
	}
	return nil
}
