package msghandler

import (
	"sync"

	"666.com/gameserver/dbserver/src/dal/utility"
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/framework/util"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

//抽獎獎池配置信息 2022.03.11
type LotteryPool struct {
	index       int32 //索引
	rewardMoney int64 //獎勵
	weights     int32 //中獎權重
	maxCount    int32 //數量上限
	count       int32 //目前數量
}

func (l *LotteryPool) IsFull() bool {
	if l.weights <= 0 {
		return true
	}
	if l.maxCount < 0 {
		return false
	} else if l.count < l.maxCount {
		return false
	} else {
		return true
	}
}

func (l *LotteryPool) AddCount(count int32) {
	if l.maxCount < 0 {
		return
	}
	l.count += count
}

func (l *LotteryPool) ResetCount() {
	l.count = 0
}

func (l *LotteryPool) GetReward() (int32, int64) {
	return 1, 10
}

// 每日登入獎勵 2022.03.08
type LoginGiftHandler struct {
	lotteryConfig []*LotteryPool //抽獎獎池配置信息
	lotteryIndex  int32
	lotteryReward int32
	lotteryMutex  sync.Mutex
}

func (lh *LoginGiftHandler) Init() {
	lh.lotteryConfig = make([]*LotteryPool, 0)
	lh.lotteryReward = 0
}

func (lh *LoginGiftHandler) IsHook(bclsid int32, sclsid int32) bool {
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

func (lh *LoginGiftHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.DBServerClassID_LoadSignActionConfig):
		return lh.LoadSignConfig(ctx, clt, m)
	case int32(netproto.DBServerClassID_LoadUserSignPro):
		return lh.LoadUserSignPro(ctx, clt, m)
	case int32(netproto.DBServerClassID_ReceiveSignRewardReqID):
		return lh.ReceiveSignReward(ctx, clt, m)
	case int32(netproto.DBServerClassID_AddSignLotteryRewardReqID):
		return lh.ReceiveSignLotteryReward(ctx, clt, m)
	}
	return nil
}

//加載簽到配置
func (lh *LoginGiftHandler) LoadSignConfig(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_LoadSignActionConfigRet)
	data := new(netproto.SignActionConfigRet)

	ret, err := lh.RequestSignConfig()
	if err != nil {
		mlog.Error("PrPs_LoadSignConfig error")
		return []*msg.Message{retmsg}
	}

	signConList := []*netproto.SignProConfig{}
	//簽到進度配置   FROM [CenterDB].[Activity].[ActivityLoginGiftSignConfig]
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
	if len(lh.lotteryConfig) > 0 {
		lh.lotteryConfig = make([]*LotteryPool, 0)
	}
	//抽獎獎池配置 FROM [CenterDB].[Activity].[ActivityLoginGiftTurnConfig]
	if ret[1] != nil {
		info := ret[1]
		for i := 0; i < len(info.Rows); i++ {
			poolConfig := new(netproto.LotteryPool)
			poolConfig.Index = proto.Int32(int32(info.GetValueByColName(i, "ID").(int64)))
			poolConfig.Reward = proto.Int64(info.GetValueByColName(i, "Reward").(int64))
			poolConfig.Weights = proto.Int32(int32(info.GetValueByColName(i, "Weights").(int64)))
			poolList = append(poolList, poolConfig)

			pool := &LotteryPool{
				index:       int32(info.GetValueByColName(i, "ID").(int64)),
				rewardMoney: info.GetValueByColName(i, "Reward").(int64),
				weights:     int32(info.GetValueByColName(i, "Weights").(int64)),
				count:       int32(info.GetValueByColName(i, "Count").(int64)),
				maxCount:    int32(info.GetValueByColName(i, "MaxCount").(int64)),
			}
			lh.lotteryConfig = append(lh.lotteryConfig, pool)
		}
	}
	data.Pool = poolList
	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

//加載玩家的簽到進度信息
func (lh *LoginGiftHandler) LoadUserSignPro(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
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

	ret, err := lh.RequestUserSignActivityInfo(UserID)
	if err != nil {
		mlog.Error("PrPs_GetUserSignPro error:%v", err)
		return []*msg.Message{retmsg}
	}

	signPro := utility.ParserUserSignActivityInfo(ret, UserID)
	retmsg.MsgData = signPro

	//加載配置數據
	ret1, err := lh.RequestSignConfig()
	if err != nil {
		mlog.Error("PrPs_LoadSignConfig error")
		return []*msg.Message{retmsg}
	}

	signConList, poolList := utility.ParserSignConfig(ret1)

	signPro.SignCon = signConList
	signPro.Pool = poolList

	return []*msg.Message{retmsg}
}

//領取簽到獎勵
func (lh *LoginGiftHandler) ReceiveSignReward(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
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
	ps.AddIntOutput("intLotteryAmount", 0)

	proc := db.NewProcedure("Hall_PrPs_GetSignActionReward", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("Hall_PrPs_GetSignActionReward error:%v", err)
		retmsg.MsgData = data
		return []*msg.Message{retmsg}
	}

	code := ret.GetReturnValue()
	if code == 1 && ret[0] != nil {
		data.BankAmount = proto.Int64(ret.GetOutputParamValue("intBankAmount").(int64))
		data.RewardAmount = proto.Int32(int32(ret.GetOutputParamValue("intRewardAmount").(int64)))
		data.LotteryAmount = proto.Int32(int32(ret.GetOutputParamValue("intLotteryAmount").(int64)))
	} else {
		data.Msg = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
		data.Code = proto.Int32(int32(code))
	}

	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}

//領取簽到抽獎獎勵
func (lh *LoginGiftHandler) ReceiveSignLotteryReward(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	lh.lotteryMutex.Lock()
	defer lh.lotteryMutex.Unlock()
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.AddSignLotteryRewardReq{}
		return mm
	})
	mm := rmm.(*netproto.AddSignLotteryRewardReq)
	if mm == nil {
		return nil
	}
	UserID := mm.GetUserID()

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_AddSignLotteryRewardResID)
	data := new(netproto.AddSignLotteryRewardRes)
	data.UserID = proto.Int32(UserID)
	data.Code = proto.Int32(0)

	if len(lh.lotteryConfig) == 0 {
		lh.LoadSignConfig(nil, nil, nil)
	}

	lh.lotteryIndex, lh.lotteryReward = lh.GetLotteryReward()
	if lh.lotteryReward <= 0 {
		mlog.Error("ReceiveSignLotteryReward GetLotteryReward Error!")
		data.Code = proto.Int32(-2)
		retmsg.MsgData = data
		return []*msg.Message{retmsg}
	}

	data.Index = proto.Int32(lh.lotteryIndex)
	data.AddAmount = proto.Int32(lh.lotteryReward)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", UserID)
	ps.AddIntInput("intRewardIndex", lh.lotteryIndex)
	ps.AddIntInput("intUpdateAmount", int32(lh.lotteryReward))
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntOutput("intErrCode", 0)
	ps.AddIntOutput("intBankAmount", 0)

	proc := db.NewProcedure("Hall_PrPs_GetSignLotteryReward", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("Hall_PrPs_GetSignLotteryReward error")
		data.Code = proto.Int32(-3)
		retmsg.MsgData = data
		return []*msg.Message{retmsg}
	}

	code := ret.GetReturnValue()
	if code == 1 {
		data.BankAmount = proto.Int64(ret.GetOutputParamValue("intBankAmount").(int64))

		for _, config := range lh.lotteryConfig {
			if config.index == lh.lotteryIndex {
				config.count += 1
				break
			}
		}

	} else {
		data.Msg = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
		data.Code = proto.Int32(int32(code))
	}

	retmsg.MsgData = data

	return []*msg.Message{retmsg}
}

// 每日登入活動的配置數據請求
func (lh *LoginGiftHandler) RequestSignConfig() (db.RecordSet, error) {
	// 讀取DB中以下兩個表格的資料
	// [CenterDB].[Activity].[ActivityLoginGiftSignConfig] 每日登入獎勵表
	// [CenterDB].[Activity].[ActivityLoginGiftTurnConfig]  樂透機率表
	d := GetDatabase()
	ps := db.NewSqlParameters()
	proc := db.NewProcedure("PrPs_LoadSignConfig", ps)
	return d.ExecProc(proc)
}
func (lh *LoginGiftHandler) RequestUserSignActivityInfo(userId int32) (db.RecordSet, error) {
	d := GetDatabase()
	ps := db.NewSqlParameters()

	ps.AddIntInput("intUserID", userId)

	ps.AddIntOutput("intCurrDay", 1)
	ps.AddVarcharOutput("chvEndTime", "")
	ps.AddIntOutput("intErrCode", 0)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddVarcharOutput("chvActiveDes", "")

	proc := db.NewProcedure("Hall_PrPs_GetUserSignActivityInfo", ps)
	return d.ExecProc(proc)
}

// 依照目前獎項及機率取得可獲得的獎勵
func (lh *LoginGiftHandler) GetLotteryReward() (int32, int32) {

	var totalWeight int32 = 0
	for _, config := range lh.lotteryConfig {
		if config.IsFull() {
			continue
		}
		totalWeight += config.weights
	}

	randValueArr := util.RndNInt(0, int(totalWeight), 1)
	if len(randValueArr) <= 0 {
		mlog.Error("抽獎操作隨機數失敗 totalWeight:%d", totalWeight)
		//data.Code = proto.Int32(Error_No_Lottery_Count)
		//data.Msg = proto.String("抽獎操作失敗")
		//rsmsg.MsgData = data
		return 0, 0
	}

	randV := randValueArr[0]
	mlog.Info("randV:%v", randV)
	var currWight int32
	var isSuccess = false
	var reward int32 = 0
	var index int32 = 0
	for _, config := range lh.lotteryConfig {
		if config.IsFull() {
			continue
		}
		currWight += config.weights
		if currWight >= int32(randV) {
			isSuccess = true
			reward = int32(config.rewardMoney)
			index = config.index
			break
		}
	}

	if isSuccess && reward > 0 {
		return index, reward
	} else {
		mlog.Error("抽獎操作隨機數失敗 totalWeight:%d", totalWeight)
		return 0, 0
	}
}

// 重置活動
func (lh *LoginGiftHandler) Reset() {
	mlog.Info("重置每日簽到活動抽獎的獎品數量")
	d := GetDatabase()
	ps := db.NewSqlParameters()
	proc := db.NewProcedure("Hall_PrPs_ResetActivityLoginGiftTurnConfig", ps)
	_, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("LoginGiftHandler Reset Error! " + err.Error())
		return
	}

	lh.LoadSignConfig(nil, nil, nil)

	/*
		for _, config := range lh.lotteryConfig {
			config.count = 0
		}
	*/

}

func (lh *LoginGiftHandler) ResetUserSignActivityInfo(userID int32) error {
	// Hall_PrPs_ResetUserSignActivityInfo
	mlog.Info("重置玩家每日簽到活動")
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userID)
	ps.AddBigIntInput("intDayOneAmount", 0)
	proc := db.NewProcedure("Hall_PrPs_ResetUserSignActivityInfo", ps)
	_, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("ResetUserSignActivityInfo Error! " + err.Error())
	}
	return err
}
