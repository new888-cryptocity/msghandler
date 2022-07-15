package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/framework/util"
	"github.com/new888-cryptocity/netproto"
	"github.com/golang/protobuf/proto"
)

type SignActionHandler struct {
}

func (sa *SignActionHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_DBServer) {
		return false
	}
	switch sclsid {
	case int32(netproto.DBServerClassID_LoadSignActionConfig),
		int32(netproto.DBServerClassID_LoadUserSignPro),
		int32(netproto.DBServerClassID_ReceiveSignRewardReqID),
		int32(netproto.DBServerClassID_AddSignLotteryRewardReqID):
		return true

	}
	return false
}

func (sa *SignActionHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.DBServerClassID_LoadSignActionConfig):
		return sa.LoadSignConfig(ctx, clt, m)
	case int32(netproto.DBServerClassID_LoadUserSignPro):
		return sa.LoadUserSignPro(ctx, clt, m)
	case int32(netproto.DBServerClassID_ReceiveSignRewardReqID):
		return sa.ReceiveSignReward(ctx, clt, m)
	case int32(netproto.DBServerClassID_AddSignLotteryRewardReqID):
		return sa.ReceiveSignLotteryReward(ctx, clt, m)
	}
	return nil
}

//加載簽到配置
func (sa *SignActionHandler) LoadSignConfig(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_LoadSignActionConfigRet)
	data := new(netproto.SignActionConfigRet)

	// 讀取DB中以下兩個表格的資料
	// [CenterDB].[Activity].[SignProConfig] 每日登入獎勵表
	// [CenterDB].[Activity].[SignLotteryPool]  樂透機率表
	d := GetDatabase()
	ps := db.NewSqlParameters()
	proc := db.NewProcedure("PrPs_LoadSignConfig", ps)
	mlog.Info("PrPs_LoadSignConfig")
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_LoadSignConfig error")
		return []*msg.Message{retmsg}
	}

	signConList := []*netproto.SignProConfig{}
	//簽到進度配置   FROM [CenterDB].[Activity].[SignProConfig]
	if ret[0] != nil {
		info := ret[0]
		for i := 0; i < len(info.Rows); i++ {
			signConfig := new(netproto.SignProConfig)
			signConfig.Day = proto.Int32(int32(info.GetValueByColName(i, "Day").(int64)))
			signConfig.Reward = proto.Int64(info.GetValueByColName(i, "Reward").(int64))
			signConfig.TargetCostAmount = proto.Int64(info.GetValueByColName(i, "TCostAmount").(int64))
			signConList = append(signConList, signConfig)
		}
	}
	data.Con = signConList
	poolList := []*netproto.LotteryPool{}
	//抽獎獎池配置 FROM [CenterDB].[Activity].[SignLotteryPool]
	if ret[1] != nil {
		info := ret[1]
		for i := 0; i < len(info.Rows); i++ {
			poolConfig := new(netproto.LotteryPool)
			poolConfig.Index = proto.Int32(int32(info.GetValueByColName(i, "ID").(int64)))
			poolConfig.Reward = proto.Int64(info.GetValueByColName(i, "Reward").(int64))
			poolConfig.Weights = proto.Int32(int32(info.GetValueByColName(i, "Weights").(int64)))
			poolList = append(poolList, poolConfig)
		}
	}
	data.Pool = poolList
	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

//加載玩家的簽到進度信息
func (sa *SignActionHandler) LoadUserSignPro(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LoadUserSignProReq{}
		return mm
	})
	mm := rmm.(*netproto.LoadUserSignProReq)
	if mm == nil {
		return nil
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_LoadUserSignProRes)
	UserID := mm.GetUserID()

	d := GetDatabase()
	ps := db.NewSqlParameters()

	ps.AddIntInput("intUserID", UserID)
	ps.AddIntOutput("intCurrDay", 1)
	ps.AddVarcharOutput("chvEndTime", "")
	ps.AddIntOutput("intErrCode", 0)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddVarcharOutput("chvActiveDes", "")

	proc := db.NewProcedure("PrPs_GetUserSignPro", ps)
	ret, err := d.ExecProc(proc)

	signPro := new(netproto.LoadUserSignProRet)
	signPro.UserID = proto.Int32(UserID)
	signPro.LotteryCount = proto.Int32(0)
	signPro.SignDay = proto.Int32(1)
	signPro.DayOneAmount = proto.Int32(0)
	signPro.DayOneState = proto.Int32(0)
	signPro.DayTwoAmount = proto.Int32(0)
	signPro.DayTwoState = proto.Int32(0)
	signPro.DayThreeAmount = proto.Int32(0)
	signPro.DayThreeState = proto.Int32(0)
	signPro.DayFourAmount = proto.Int32(0)
	signPro.DayFourState = proto.Int32(0)
	signPro.DayFiveAmount = proto.Int32(0)
	signPro.DayFiveState = proto.Int32(0)
	signPro.DaySixAmount = proto.Int32(0)
	signPro.DaySixState = proto.Int32(0)
	signPro.DaySevenAmount = proto.Int32(0)
	signPro.DaySevenState = proto.Int32(0)
	signPro.EndTime = proto.Int32(0)
	if err != nil {
		mlog.Error("PrPs_GetUserSignPro error:%v", err)
		return []*msg.Message{retmsg}
	}
	signPro.Code = proto.Int32(int32(ret.GetOutputParamValue("intErrCode").(int64)))
	signPro.Msg = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
	if ret.GetReturnValue() < 0 {
		retmsg.MsgData = signPro
		return []*msg.Message{retmsg}
	}
	//解析進度數據
	if ret[0] != nil {
		info := ret[0]
		endTimeStr := ret.GetOutputParamValue("chvEndTime").(string)
		signPro.EndTime = proto.Int32(util.FormatTimeStamp(endTimeStr))
		signPro.SignDay = proto.Int32(int32(ret.GetOutputParamValue("intCurrDay").(int64)))
		signPro.ActiveDes = proto.String(ret.GetOutputParamValue("chvActiveDes").(string))
		for i := 0; i < len(info.Rows); i++ {
			signPro.LotteryCount = proto.Int32(int32(info.GetValueByColName(i, "LotteryCount").(int64)))
			signPro.DayOneAmount = proto.Int32(int32(info.GetValueByColName(i, "DayOneAmount").(int64)))
			signPro.DayOneState = proto.Int32(int32(info.GetValueByColName(i, "DayOneState").(int64)))
			signPro.DayTwoAmount = proto.Int32(int32(info.GetValueByColName(i, "DayTwoAmount").(int64)))
			signPro.DayTwoState = proto.Int32(int32(info.GetValueByColName(i, "DayTwoState").(int64)))
			signPro.DayThreeAmount = proto.Int32(int32(info.GetValueByColName(i, "DayThreeAmount").(int64)))
			signPro.DayThreeState = proto.Int32(int32(info.GetValueByColName(i, "DayThreeState").(int64)))
			signPro.DayFourAmount = proto.Int32(int32(info.GetValueByColName(i, "DayFourAmount").(int64)))
			signPro.DayFourState = proto.Int32(int32(info.GetValueByColName(i, "DayFourState").(int64)))
			signPro.DayFiveAmount = proto.Int32(int32(info.GetValueByColName(i, "DayFiveAmount").(int64)))
			signPro.DayFiveState = proto.Int32(int32(info.GetValueByColName(i, "DayFiveState").(int64)))
			signPro.DaySixAmount = proto.Int32(int32(info.GetValueByColName(i, "DaySixAmount").(int64)))
			signPro.DaySixState = proto.Int32(int32(info.GetValueByColName(i, "DaySixState").(int64)))
			signPro.DaySevenAmount = proto.Int32(int32(info.GetValueByColName(i, "DaySevenAmount").(int64)))
			signPro.DaySevenState = proto.Int32(int32(info.GetValueByColName(i, "DaySevenState").(int64)))
		}
	}

	//加載配置數據
	ps1 := db.NewSqlParameters()
	proc1 := db.NewProcedure("PrPs_LoadSignConfig", ps1)
	ret1, err := d.ExecProc(proc1)
	if err != nil {
		mlog.Error("PrPs_LoadSignConfig error")
		return []*msg.Message{retmsg}
	}

	signConList := []*netproto.SignProConfig{}
	//簽到進度配置
	if ret1[0] != nil {
		info := ret1[0]
		for i := 0; i < len(info.Rows); i++ {
			signConfig := new(netproto.SignProConfig)
			signConfig.Day = proto.Int32(int32(info.GetValueByColName(i, "Day").(int64)))
			signConfig.Reward = proto.Int64(info.GetValueByColName(i, "Reward").(int64))
			signConfig.TargetCostAmount = proto.Int64(info.GetValueByColName(i, "TCostAmount").(int64))
			signConList = append(signConList, signConfig)
		}
	}
	signPro.SignCon = signConList
	poolList := []*netproto.LotteryPool{}
	//抽獎獎池配置
	if ret[1] != nil {
		info := ret1[1]
		for i := 0; i < len(info.Rows); i++ {
			poolConfig := new(netproto.LotteryPool)
			poolConfig.Index = proto.Int32(int32(info.GetValueByColName(i, "ID").(int64)))
			poolConfig.Reward = proto.Int64(info.GetValueByColName(i, "Reward").(int64))
			poolConfig.Weights = proto.Int32(int32(info.GetValueByColName(i, "Weights").(int64)))
			poolList = append(poolList, poolConfig)
		}
	}
	signPro.Pool = poolList
	retmsg.MsgData = signPro
	return []*msg.Message{retmsg}
}

//領取簽到獎勵
func (sa *SignActionHandler) ReceiveSignReward(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ReceiveSignRewardReq{}
		return mm
	})
	mm := rmm.(*netproto.ReceiveSignRewardReq)
	if mm == nil {
		return nil
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_ReceiveSignRewardResID)
	data := new(netproto.ReceiveSignRewardRes)

	UserID := mm.GetUserID()
	Day := mm.GetDay()

	data.UserID = proto.Int32(UserID)
	data.Day = proto.Int32(Day)
	data.Code = proto.Int32(0)

	d := GetDatabase()
	ps := db.NewSqlParameters()

	ps.AddIntInput("intUserID", UserID)
	ps.AddIntInput("intDay", Day)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntOutput("intErrCode", 0)
	ps.AddIntOutput("intBankAmount", 0)
	ps.AddIntOutput("intRewardAmount", 0)

	proc := db.NewProcedure("PrPs_GetSignActionReward", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_GetSignActionReward error:%v", err)
		retmsg.MsgData = data
		return []*msg.Message{retmsg}
	}

	code := ret.GetReturnValue()
	if code == 1 && ret[0] != nil {
		data.BankAmount = proto.Int64(ret.GetOutputParamValue("intBankAmount").(int64))
		data.RewardAmount = proto.Int32(int32(ret.GetOutputParamValue("intRewardAmount").(int64)))
	} else {
		data.Msg = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
		data.Code = proto.Int32(int32(code))
	}

	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

//領取簽到抽獎獎勵
func (sa *SignActionHandler) ReceiveSignLotteryReward(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.AddSignLotteryRewardReq{}
		return mm
	})
	mm := rmm.(*netproto.AddSignLotteryRewardReq)
	if mm == nil {
		return nil
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_AddSignLotteryRewardResID)
	data := new(netproto.AddSignLotteryRewardRes)

	UserID := mm.GetUserID()
	addAmount := mm.GetAddAmount()
	data.UserID = proto.Int32(UserID)
	data.Code = proto.Int32(0)
	data.Index = proto.Int32(mm.GetIndex())
	data.AddAmount = proto.Int32(addAmount)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", UserID)
	ps.AddIntInput("intUpdateAmount", int32(addAmount))
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntOutput("intErrCode", 0)
	ps.AddIntOutput("intBankAmount", 0)

	proc := db.NewProcedure("PrPs_GetSignLotteryReward", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrPs_GetSignLotteryReward error")
		retmsg.MsgData = data
		return []*msg.Message{retmsg}
	}

	code := ret.GetReturnValue()
	if code == 1 {
		data.BankAmount = proto.Int64(ret.GetOutputParamValue("intBankAmount").(int64))
	} else {
		data.Msg = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
		data.Code = proto.Int32(int32(code))
	}

	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}
