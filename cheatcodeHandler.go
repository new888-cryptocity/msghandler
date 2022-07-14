package msghandler

import (
	"fmt"
	"strconv"
	"time"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

type CheatcodeHandler struct {
	loginGiftHandler  *LoginGiftHandler
	pointsRaceHandler *PointsRaceHandler
}

func NewCheatCodeHandler(lgh *LoginGiftHandler, prh *PointsRaceHandler) *CheatcodeHandler {
	handler := &CheatcodeHandler{
		loginGiftHandler:  lgh,
		pointsRaceHandler: prh,
	}
	return handler
}

func (cc *CheatcodeHandler) Init() {}
func (cc *CheatcodeHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && (sclsid == int32(netproto.PlatformCommonClassID_CheatCodeID))
}

func (cc *CheatcodeHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Debug("[Cheatcode OnMessage] %d %d", m.BClassID, m.SClassID)
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.CheatCodeMsg{}
		return mm
	})
	mm := rmm.(*netproto.CheatCodeMsg)
	if mm == nil {
		return nil
	}
	userid := rpc.GetUserID()

	return cc.Excute(userid, mm)
}

func (cc *CheatcodeHandler) Excute(userid int32, mm *netproto.CheatCodeMsg) []*msg.Message {
	mlog.Debug("Cheatcode type = %d", mm.Type)
	switch *mm.Type {
	case 200001: // 增加每日抽獎次數

		if len(mm.ParamInt) > 0 {
			count := mm.ParamInt[0]
			return cc.addDailySignLotteryCount(userid, count)
		}
		break
	case 200002: // 重置每日登入資訊
		return cc.resetUserSignInfo(*mm.Type, userid)
	case 200003: // 重置每日登入抽獎獎項數量
		return cc.resetDailyLotteryRewardCount(*mm.Type)
	case 200004: // 設定每日登狀態
		return cc.setUserSignInfo(*mm.Type, userid, mm.GetParamInt())
	case 200005: // 取得每日登入抽獎獎項數量
		return cc.getDailyLotteryRewardCount(*mm.Type)
	case 200006: // 取得每日登入抽獎獎項最大
		return cc.getDailyLotteryRewardMaxCount(*mm.Type)

	case 20101: // 刷新積分賽排名
		return cc.updatePointsRaceRanking(*mm.Type)

	case 888887: // 設定錢包位址
		if len(mm.ParamStr) > 0 {
			address := mm.ParamStr[0]
			return cc.setWalletAddress(*mm.Type, userid, address)
		}
		break
	case 888888: // 增加獎勵金額到玩家身上
		if len(mm.ParamInt) > 0 {
			count := mm.ParamInt[0]
			return cc.addRewardToUser(*mm.Type, userid, count)
		}
		break
	case 888889: // 設定玩家的免費虛點
		if len(mm.ParamInt) > 0 {
			count := mm.ParamInt[0]
			return cc.setFreePointToUser(*mm.Type, userid, count)
		}
		break

	case 878787: // 清除玩家本日的Money.Log
		return cc.cleanUserTodayMoneyLog(*mm.Type, userid)

	}

	return nil
}

func (cc *CheatcodeHandler) addDailySignLotteryCount(userid int32, count int32) []*msg.Message {
	//Hall_PrPs_SetSignLotteryCount
	totalCount := int32(0)
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intCount", count)
	ps.AddIntOutput("intTotalCount", totalCount)
	proc := db.NewProcedure("Hall_PrPs_SetSignLotteryCount", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error(err.Error())
		return nil
	}
	mlog.Debug("%v", ret)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(200001),
		ParamInt: []int32{totalCount},
		ParamStr: []string{"增加每日登入的抽獎次數成功!"},
	}

	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}

func (cc *CheatcodeHandler) resetUserSignInfo(cheatType int32, userid int32) []*msg.Message {
	//[Hall_PrPs_ResetUserSignActivityInfo]
	err := cc.loginGiftHandler.ResetUserSignActivityInfo(userid)
	if err != nil {
		mlog.Error(err.Error())
		return nil
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(cheatType),
		ParamStr: []string{"重置玩家每日登入資訊成功!"},
	}

	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}

func (cc *CheatcodeHandler) setUserSignInfo(cheatType int32, userid int32, status []int32) []*msg.Message {

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
	d := GetDatabase()
	if len(status) < 7 {
		resp := &netproto.CheatCodeMsg{
			Type:     proto.Int32(cheatType),
			ParamStr: []string{"重置每日登入資訊失敗 參數設定數量小於 7!"},
		}
		retmsg.MsgData = resp

		return []*msg.Message{retmsg}
	}

	str := fmt.Sprintf("[DayOneState] = %d, [DayTwoState] = %d, [DayThreeState] = %d, [DayFourState] = %d, [DayFiveState] = %d, [DaySixState] = %d, [DaySevenState] = %d ", status[0], status[1], status[2], status[3], status[4], status[5], status[6])
	ts := "'2022-03-01 12:00'"
	str1 := fmt.Sprintf("[DayOneLastUpdateTime] = %s,[DayTwoLastUpdateTime] = %s,[DayThreeLastUpdateTime] = %s,[DayFourLastUpdateTime] = %s,[DayFiveLastUpdateTime] = %s,[DaySixLastUpdateTime] = %s,[DaySevenLastUpdateTime] = %s", ts, ts, ts, ts, ts, ts, ts)
	str2 := "[DayOneAmount] = 0, [DayTwoAmount] = 0, [DayThreeAmount] = 0, [DayFourAmount] = 0, [DayFiveAmount] = 0, [DaySixAmount] = 0, [DaySevenAmount] = 0"
	sql := fmt.Sprintf("Update [CenterDB].[Users].[UserActivitySignData] Set %s, %s, %s Where UserID = %d", str, str1, str2, userid)
	ret, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error(err.Error())
		return nil
	}
	mlog.Debug("%v", ret)

	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(cheatType),
		ParamStr: []string{"設定每日登入資訊成功!"},
	}

	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}

func (cc *CheatcodeHandler) resetDailyLotteryRewardCount(cheatType int32) []*msg.Message {
	if cc.loginGiftHandler != nil {
		cc.loginGiftHandler.Reset()
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
		resp := &netproto.CheatCodeMsg{
			Type:     proto.Int32(cheatType),
			ParamStr: []string{"重置每日登入抽獎數量成功!"},
		}

		retmsg.MsgData = resp

		return []*msg.Message{retmsg}
	}
	return nil
}

func (cc *CheatcodeHandler) getDailyLotteryRewardCount(cheatType int32) []*msg.Message {
	if cc.loginGiftHandler != nil {
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
		resp := &netproto.CheatCodeMsg{
			Type:     proto.Int32(cheatType),
			ParamStr: []string{"取得每日登入抽獎數量成功!"},
		}
		for _, config := range cc.loginGiftHandler.lotteryConfig {
			resp.ParamInt = append(resp.ParamInt, config.count)
		}

		retmsg.MsgData = resp

		return []*msg.Message{retmsg}
	}
	return nil
}

func (cc *CheatcodeHandler) getDailyLotteryRewardMaxCount(cheatType int32) []*msg.Message {
	if cc.loginGiftHandler != nil {
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
		resp := &netproto.CheatCodeMsg{
			Type:     proto.Int32(cheatType),
			ParamStr: []string{"取得每日登入抽獎獎項的最大數量成功!"},
		}
		for _, config := range cc.loginGiftHandler.lotteryConfig {
			resp.ParamInt = append(resp.ParamInt, config.maxCount)
		}

		retmsg.MsgData = resp

		return []*msg.Message{retmsg}
	}
	return nil
}

func (cc *CheatcodeHandler) updatePointsRaceRanking(cheatType int32) []*msg.Message {
	if cc.pointsRaceHandler != nil {
		cc.pointsRaceHandler.updatePointsRaceRanking()
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
		resp := &netproto.CheatCodeMsg{
			Type:     proto.Int32(cheatType),
			ParamStr: []string{"刷新積分賽排行成功!"},
		}

		retmsg.MsgData = resp

		return []*msg.Message{retmsg}
	}
	return nil
}
func (cc *CheatcodeHandler) setWalletAddress(cheatType int32, userid int32, address string) []*msg.Message {
	d := GetDatabase()
	sql := fmt.Sprintf("update [CenterDB].[Money].[Cash] set WalletAddress='%s' where UserID=%d", address, userid)
	ret, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error(err.Error())
		return nil
	}
	mlog.Debug("%v", ret)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(cheatType),
		ParamStr: []string{"設定錢包位址為:" + address},
	}

	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}

func (cc *CheatcodeHandler) addRewardToUser(cheatType int32, userid int32, count int32) []*msg.Message {

	var amount int64 = 0
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intUpdateAmount", count)
	ps.AddIntInput("intSourceType", 18009)
	ps.AddBigIntOutput("lngCashAmount", amount)
	proc := db.NewProcedure("Hall_PrPs_UpdateUserMoney", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error(err.Error())
		return nil
	}
	mlog.Debug("%v", ret)
	amount = ret.GetOutputParamValue("lngCashAmount").(int64)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(cheatType),
		ParamStr: []string{"取得每日獎勵金:" + strconv.FormatInt(int64(count), 10)},
		ParamInt: []int32{int32(amount) + count},
	}

	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}

func (cc *CheatcodeHandler) setFreePointToUser(cheatType int32, userid int32, count int32) []*msg.Message {

	d := GetDatabase()
	sql := fmt.Sprintf("update [CenterDB].[Money].[Cash] set Point=%d where UserID=%d", count, userid)
	ret, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error(err.Error())
		return nil
	}
	mlog.Debug("%v", ret)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)
	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(cheatType),
		ParamStr: []string{"設定虛點為:" + strconv.FormatInt(int64(count), 10)},
		ParamInt: []int32{count},
	}

	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}

func (cc *CheatcodeHandler) cleanUserTodayMoneyLog(cheatType int32, userid int32) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.PlatformCommonClassID_CheatCodeID)

	current_time := time.Now()
	st := fmt.Sprintf("%d-%02d-%02d", current_time.Year(), current_time.Month(), current_time.Day())
	//mlog.Debug(time.Now().Format("2006-01-01"))

	current_time = current_time.AddDate(0, 0, 1)
	et := fmt.Sprintf("%d-%02d-%02d", current_time.Year(), current_time.Month(), current_time.Day())

	d := GetDatabase()
	sql := fmt.Sprintf("UPDATE [CenterDB].[dbo].[UpdateMoneyLog] SET CreateTime = '2020-02-20' where UserID = %d AND CreateTime Between '%s' AND '%s'", userid, st, et)
	_, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error(err.Error())
		resp := &netproto.CheatCodeMsg{
			Type:     proto.Int32(cheatType),
			ParamStr: []string{"清除今日玩家Money Log 失敗! ", err.Error()},
		}
		retmsg.MsgData = resp
		return []*msg.Message{retmsg}
	}

	resp := &netproto.CheatCodeMsg{
		Type:     proto.Int32(cheatType),
		ParamStr: []string{"清除今日玩家Money Log 成功! "},
	}
	retmsg.MsgData = resp

	return []*msg.Message{retmsg}
}
