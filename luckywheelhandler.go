package msghandler

import (
	"strings"
	"time"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	"github.com/golang/protobuf/proto"
)

type LuckyWheelHandler struct {
}

func (lw *LuckyWheelHandler) Init() {}

func (lw *LuckyWheelHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_DBServer) {
		return false
	}
	switch sclsid {
	case int32(netproto.DBServerClassID_LoadLuckyWheelConfig),
		int32(netproto.DBServerClassID_LoadUserLuckyWheelData),
		int32(netproto.DBServerClassID_LuckyWheelGetReward),
		int32(netproto.DBServerClassID_LuckyWheelReset),
		int32(netproto.DBServerClassID_LuckyWheelForceList),
		int32(netproto.DBServerClassID_LuckyWheelUserForceUpdate):
		return true

	}
	return false
}

func (lw *LuckyWheelHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.DBServerClassID_LoadLuckyWheelConfig):
		return lw.LoadLuckyWheelConfig(ctx, clt, m)
	case int32(netproto.DBServerClassID_LoadUserLuckyWheelData):
		return lw.LoadUserLuckyWheel(ctx, clt, m)
	case int32(netproto.DBServerClassID_LuckyWheelGetReward):
		return lw.LuckyWheelGetReward(ctx, clt, m)
	case int32(netproto.DBServerClassID_LuckyWheelReset):
		return lw.LuckyWheelReset(ctx, clt, m)
	case int32(netproto.DBServerClassID_LuckyWheelForceList):
		return lw.LuckyWheelForceList(ctx, clt, m)
	case int32(netproto.DBServerClassID_LuckyWheelUserForceUpdate):
		return lw.LuckyWheelUserForceUpdate(ctx, clt, m)
	}
	return nil
}

// 加载幸運轉輪設定
func (lw *LuckyWheelHandler) LoadLuckyWheelConfig(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelDBRequest{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_LoadLuckyWheelConfigRet)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	mm := rmm.(*netproto.LuckyWheelDBRequest)
	ps.AddIntInput("intGroupID", *mm.GroupID)
	proc := db.NewProcedure("PrPs_LoadLuckyWheelConfig", ps)
	mlog.Info("PrPs_LoadLuckyWheelConfig")
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_LoadLuckyWheelConfig error")
		return []*msg.Message{retmsg}
	}

	return ParseLuckyWheelConfig(ret, retmsg)
}

// 處理幸運轉輪玩家資料
func (lw *LuckyWheelHandler) LoadUserLuckyWheel(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelDBRequest{}
		return mm
	})

	rpcInfo := rpcmsg.GetRPCInfo(m)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.DBServerClassID_LoadUserLuckyWheelDataRet)

	d := GetDatabase()
	ps := db.NewSqlParameters()

	UserID := *rpcInfo.UserID
	mm := rmm.(*netproto.LuckyWheelDBRequest)

	ps.AddIntInput("intGroupID", *mm.GroupID)
	ps.AddIntInput("intUserID", UserID)

	proc := db.NewProcedure("PrPs_LoadLuckyWheelUserData", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("PrPs_LoadLuckyWheelUserData error:%v", err)
		return []*msg.Message{retmsg}
	}

	if ret[0] != nil {
		info := ret[0]
		retData := &netproto.LuckyWheelUserData{}
		retData.SpinWeight = proto.Int32(int32(info.GetValueByColName(0, "SpinWeight").(int64)))
		retData.SpinCount = proto.Int32(int32(info.GetValueByColName(0, "SpinCount").(int64)))
		retData.CanSpin = proto.Int32(int32(info.GetValueByColName(0, "CanSpin").(int64)))
		updateTime := strings.Split(info.GetValueByColName(0, "UpdateTime").(time.Time).String(), " +")[0]
		retData.UpdateTime = proto.String(updateTime)
		retmsg.MsgData = retData
		return []*msg.Message{retmsg}
	}

	return nil
}

// 處理幸運轉輪抽獎
func (sa *LuckyWheelHandler) LuckyWheelGetReward(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcInfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelGetReward{}
		return mm
	})
	mm := rmm.(*netproto.LuckyWheelGetReward)
	if mm == nil {
		return nil
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_LuckyWheelSpinRet)
	UserID := *rpcInfo.UserID

	d := GetDatabase()
	ps := db.NewSqlParameters()

	ps.AddIntInput("intUserID", UserID)
	ps.AddIntInput("intReward", mm.GetRewardId())
	ps.AddIntInput("intRewardAmount", mm.GetRewardAmount())
	ps.AddIntInput("intSpinWeight", mm.GetSpinWeight())
	ps.AddIntInput("intGroupID", mm.GetGroupID())

	proc := db.NewProcedure("PrPs_LuckyWheelGetReward", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("PrPs_LuckyWheelGetReward error:%v", err)
		return []*msg.Message{retmsg}
	}
	retData := &netproto.DZPKHALLLuckyWheelSpinRet{}
	currentMoney := ret.GetOutputParamValue("CurrentMoney").(int64)
	retData.Code = proto.Int32(0)
	retData.RewardType = mm.RewardType
	retData.RewardAmount = mm.RewardAmount
	retData.CurrentMoney = proto.Int64(currentMoney)
	retmsg.MsgData = retData
	return []*msg.Message{retmsg}
}

// 幸運轉輪重置
func (lw *LuckyWheelHandler) LuckyWheelReset(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelDBRequest{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_LoadLuckyWheelConfigRet)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	mm := rmm.(*netproto.LuckyWheelDBRequest)
	ps.AddIntInput("intGroupID", *mm.GroupID)
	proc := db.NewProcedure("PrPs_LoadLuckyWheelReset", ps)
	mlog.Info("PrPs_LoadLuckyWheelReset")
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_LoadLuckyWheelReset error")
		return []*msg.Message{retmsg}
	}

	return ParseLuckyWheelConfig(ret, retmsg)
}

func ParseLuckyWheelConfig(ret db.RecordSet, retmsg *msg.Message) []*msg.Message {
	data := new(netproto.LuckyWheelConfig)
	rewardList := []*netproto.LuckyWheelReward{}
	// 獎項清單
	if ret[0] != nil {
		info := ret[0]
		for i := 0; i < len(info.Rows); i++ {
			reward := new(netproto.LuckyWheelReward)
			reward.Id = proto.Int32(int32(info.GetValueByColName(i, "Id").(int64)))
			reward.RewardName = proto.String(info.GetValueByColName(i, "RewardName").(string))
			reward.RewardType = proto.Int32(int32(info.GetValueByColName(i, "RewardType").(int64)))
			reward.RewardAmount = proto.Int32(int32(info.GetValueByColName(i, "RewardAmount").(int64)))
			reward.LimitAmount = proto.Int32(int32(info.GetValueByColName(i, "LimitAmount").(int64)))
			reward.AssignAmount = proto.Int32(int32(info.GetValueByColName(i, "AssignAmount").(int64)))
			reward.PosIndex = proto.String(info.GetValueByColName(i, "PosIndex").(string))
			reward.BulletType = proto.Int32(int32(info.GetValueByColName(i, "BulletType").(int64)))
			reward.RewardWeight = proto.Int32(int32(info.GetValueByColName(i, "RewardWeight").(int64)))
			rewardList = append(rewardList, reward)
		}
	}
	data.Rewards = rewardList
	probabilityList := []*netproto.LuckyWheelProbability{}
	// 獎項機率
	if ret[1] != nil {
		info := ret[1]
		for i := 0; i < len(info.Rows); i++ {
			probability := new(netproto.LuckyWheelProbability)
			probability.Id = proto.Int32(int32(info.GetValueByColName(i, "Id").(int64)))
			probability.RewardId = proto.Int32(int32(info.GetValueByColName(i, "RewardId").(int64)))
			probability.ProbabilityGroup = proto.Int32(int32(info.GetValueByColName(i, "ProbabilityGroup").(int64)))
			probability.Probability = proto.Int32(int32(info.GetValueByColName(i, "Probability").(int64)))
			probability.LevelWeightLimitDown = proto.Int32(int32(info.GetValueByColName(i, "LevelWeightLimitDown").(int64)))
			probabilityList = append(probabilityList, probability)
		}
	}
	data.Probabilitys = probabilityList
	if ret[2] != nil {
		info := ret[2]
		data.ResetTime = proto.String(info.GetValueByColName(0, "ResetTime").(string))
	}

	data.Records = &netproto.DZPKHALLLuckyWheelUserRewardRecordsRet{}
	if ret[3] != nil {
		info := ret[3]
		// 獲獎紀錄反轉，因為從DB拿出來是倒序
		records := make([]*netproto.DZPKHALLLuckyWheelUserRewardRecord, 0)
		for i := 0; i < len(info.Rows); i++ {
			record := new(netproto.DZPKHALLLuckyWheelUserRewardRecord)
			record.NickName = proto.String(info.GetValueByColName(i, "NickName").(string))
			record.RewardType = proto.Int32(int32(info.GetValueByColName(i, "RewardType").(int64)))
			record.RewardAmount = proto.Int32(int32(info.GetValueByColName(i, "RewardAmount").(int64)))
			record.BulletType = proto.Int32(int32(info.GetValueByColName(i, "BulletType").(int64)))
			records = append(records, record)
		}
		recordCount := len(records) - 1
		for i := recordCount; i >= 0; i-- {
			data.Records.RewardList = append(data.Records.RewardList, records[i])
		}
	}
	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

// 幸運轉輪指定派獎名單
func (lw *LuckyWheelHandler) LuckyWheelForceList(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelDBRequest{}
		return mm
	})

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_LuckyWheelForceListRet)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	mm := rmm.(*netproto.LuckyWheelDBRequest)
	ps.AddIntInput("intGroupID", *mm.GroupID)
	proc := db.NewProcedure("PrPs_LoadLuckyWheelForceList", ps)
	mlog.Info("PrPs_LoadLuckyWheelForceList")
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_LoadLuckyWheelForceList error")
		return []*msg.Message{retmsg}
	}

	flist := &netproto.LuckyWheelForceList{List: make([]*netproto.LuckyWheelForce, 0)}

	for i := 0; i < len(ret[0].Rows); i++ {
		force := new(netproto.LuckyWheelForce)
		force.ID = proto.Int32(int32(ret[0].GetValueByColName(i, "Id").(int64)))
		force.UserID = proto.Int32(int32(ret[0].GetValueByColName(i, "UserId").(int64)))
		force.RewardID = proto.Int32(int32(ret[0].GetValueByColName(i, "RewardId").(int64)))
		flist.List = append(flist.List, force)
	}
	retmsg.MsgData = flist

	return []*msg.Message{retmsg}
}

func (lw *LuckyWheelHandler) LuckyWheelUserForceUpdate(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelForce{}
		return mm
	})

	d := GetDatabase()
	ps := db.NewSqlParameters()
	mm := rmm.(*netproto.LuckyWheelForce)
	ps.AddIntInput("intID", mm.GetID())
	proc := db.NewProcedure("PrPs_LuckyWheelUserForceUpdate", ps)
	mlog.Info("PrPs_LuckyWheelUserForceUpdate")
	_, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_LuckyWheelUserForceUpdate error")
		return nil
	}

	return nil
}
