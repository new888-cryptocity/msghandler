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
	"github.com/golang/protobuf/ptypes/empty"
)

type ActivityHandler struct {
}

func (rh *ActivityHandler) Init() {}
func (rh *ActivityHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && (sclsid == int32(netproto.HallMsgClassID_ActivityEnroll) ||
		sclsid == int32(netproto.HallMsgClassID_ActivityGetReward) || sclsid == int32(netproto.HallMsgClassID_ActivityGetInfo) ||
		sclsid == int32(netproto.HallMsgClassID_ActivitGetAdvanceInfo) ||
		sclsid == int32(netproto.HallMsgClassID_ActivitGetRechargeInfo) || sclsid == int32(netproto.HallMsgClassID_ActivitGetRechargeReward))
}

func (rh *ActivityHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, _ := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &empty.Empty{}
		return mm
	})

	userid := rpcinfo.GetUserID()
	if userid <= 0 {
		mlog.Error("首充活動請求用戶ID錯誤：%v", userid)
		return nil
	}

	if m.SClassID == int32(netproto.HallMsgClassID_ActivityEnroll) { //報名參與活動
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_ActivityEnroll", ps)
		ret, err := d.ExecProc(proc)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_ActivityEnrollRet)
		if err != nil {
			ctx.Error("執行存儲過程PrPs_ActivityEnroll錯誤, %v", err)
			activityInfoRet := &netproto.ActivityInfoRet{}
			activityInfoRet.Code = proto.Int32(0)
			activityInfoRet.Message = proto.String("服務器錯誤.")
			retmsg.MsgData = activityInfoRet
		} else {
			activityInfoRet := parseActivityInfoRet(ret)
			retmsg.MsgData = activityInfoRet
		}

		return []*msg.Message{retmsg}
	}
	if m.SClassID == int32(netproto.HallMsgClassID_ActivityGetReward) { //領取活動獎勵
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_ActivityGetReward", ps)
		ret, err := d.ExecProc(proc)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_ActivityGetRewardRet)
		if err != nil {
			ctx.Error("執行存儲過程PrPs_ActivityGetReward錯誤, %v", err)
			activityInfoRet := &netproto.ActivityInfoRet{}
			activityInfoRet.Code = proto.Int32(0)
			activityInfoRet.Message = proto.String("服務器錯誤.")
			retmsg.MsgData = activityInfoRet
		} else {
			activityInfoRet := parseActivityInfoRet(ret)
			retmsg.MsgData = activityInfoRet
		}

		return []*msg.Message{retmsg}
	}
	if m.SClassID == int32(netproto.HallMsgClassID_ActivityGetInfo) { //獲取活動信息
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_GetActivityInfo", ps)
		ret, err := d.ExecProc(proc)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_ActivityGetInfoRet)
		if err != nil {
			ctx.Error("執行存儲過程PrPs_GetActivityInfo錯誤, %v", err)
			activityInfoRet := &netproto.ActivityInfoRet{}
			activityInfoRet.Code = proto.Int32(0)
			activityInfoRet.Message = proto.String("服務器錯誤.")
			retmsg.MsgData = activityInfoRet
		} else {
			activityInfoRet := parseActivityInfoRet(ret)
			retmsg.MsgData = activityInfoRet
		}

		return []*msg.Message{retmsg}
	}

	if m.SClassID == int32(netproto.HallMsgClassID_ActivitGetAdvanceInfo) { //獲取激情晉級信息
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddIntOutput("yesterdayAmount", 0) //昨天纍計贏金幣
		ps.AddIntOutput("nowadayAmount", 0)   //今天纍計贏取金幣
		ps.AddIntOutput("yesterdayReward", 0) //昨日獎勵
		ps.AddIntOutput("tomorrowAmount", 0)  //明日可領取獎勵
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_AdvanceActivityInfo", ps)
		ret, err := d.ExecProc(proc)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_ActivitGetAdvanceInfoRet)
		advanceInfoRet := &netproto.AdvanceInfoRet{}
		if err != nil {
			ctx.Error("執行存儲過程PrPs_AdvanceActivityInfo錯誤, %v", err)
			advanceInfoRet.Code = proto.Int32(0)
			advanceInfoRet.Message = proto.String("服務器錯誤.")
			retmsg.MsgData = advanceInfoRet
		} else {
			if ret.GetReturnValue() == 1 {
				advanceInfoRet.Code = proto.Int32(1)
				advanceInfoRet.Message = proto.String("獲取成功")
				advanceInfoRet.YesterdayAmount = proto.Int32(int32(ret.GetOutputParamValue("yesterdayAmount").(int64)))
				advanceInfoRet.NowadayAmount = proto.Int32(int32(ret.GetOutputParamValue("nowadayAmount").(int64)))
				advanceInfoRet.YesterdayReward = proto.Int32(int32(ret.GetOutputParamValue("yesterdayReward").(int64)))
				advanceInfoRet.TomorrowAmount = proto.Int32(int32(ret.GetOutputParamValue("tomorrowAmount").(int64)))

				if ret.GetRetTableCount() > 0 {
					tbAdvanceInfo := ret[0]
					for i := 0; i < len(tbAdvanceInfo.Rows); i++ {
						advanceConfig := new(netproto.AdvanceConfig)
						advanceConfig.AdvanceID = proto.Int32(int32(tbAdvanceInfo.GetValueByColName(i, "AdvanceID").(int64)))
						advanceConfig.TotalAmount = proto.Int32(int32(tbAdvanceInfo.GetValueByColName(i, "TotalAmount").(int64)))
						advanceConfig.RewardAmount = proto.Int32(int32(tbAdvanceInfo.GetValueByColName(i, "RewardAmount").(int64)))
						advanceConfig.Status = proto.Int32(0)
						advanceInfoRet.AdvanceConfig = append(advanceInfoRet.AdvanceConfig, advanceConfig)
					}
				}
				retmsg.MsgData = advanceInfoRet
			} else {
				advanceInfoRet.Code = proto.Int32(-1)
				advanceInfoRet.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
				retmsg.MsgData = advanceInfoRet
			}

		}
		return []*msg.Message{retmsg}
	}

	if m.SClassID == int32(netproto.HallMsgClassID_ActivitGetRechargeInfo) { //獲取充值活動信息
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddIntOutput("maxAmount", 0)            //贏額進度最大值
		ps.AddIntOutput("curAmount", 0)            //贏額進度當前值
		ps.AddIntOutput("recAmount", 0)            //已領取金額
		ps.AddIntOutput("availableAmoun", 0)       //可領取金額
		ps.AddVarcharOutput("activityEndTime", "") //活動結束時間
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_GetRechargeActivityInfo", ps)
		ret, err := d.ExecProc(proc)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_ActivitGetRechargeInfoRet)
		rechargeInfoRet := &netproto.RechargeActivityRet{}
		if err != nil {
			ctx.Error("執行存儲過程PrPs_GetRechargeActivityInfo錯誤, %v", err)
			rechargeInfoRet.Code = proto.Int32(0)
			rechargeInfoRet.Message = proto.String("服務器錯誤.")
			retmsg.MsgData = rechargeInfoRet
		} else {
			if ret.GetReturnValue() == 1 {
				rechargeInfoRet.Code = proto.Int32(1)
				rechargeInfoRet.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
				rechargeInfoRet.ActivityEndTime = proto.String(ret.GetOutputParamValue("activityEndTime").(string))
				rechargeInfoRet.MaxAmount = proto.Int32(int32(ret.GetOutputParamValue("maxAmount").(int64)))
				rechargeInfoRet.CurAmount = proto.Int32(int32(ret.GetOutputParamValue("curAmount").(int64)))
				rechargeInfoRet.RecAmount = proto.Int32(int32(ret.GetOutputParamValue("recAmount").(int64)))
				rechargeInfoRet.AvailableAmoun = proto.Int32(int32(ret.GetOutputParamValue("availableAmoun").(int64)))
				if ret.GetRetTableCount() > 0 {
					tbRechargeInfo := ret[0]
					for i := 0; i < len(tbRechargeInfo.Rows); i++ {
						rechargeConfig := new(netproto.RechargeConfig)
						rechargeConfig.RechargeID = proto.Int32(int32(tbRechargeInfo.GetValueByColName(i, "RechargeId").(int64)))
						rechargeConfig.RechargeAmount = proto.Int32(int32(tbRechargeInfo.GetValueByColName(i, "TotalRecharge").(int64)))
						rechargeConfig.WinAmount = proto.Int32(int32(tbRechargeInfo.GetValueByColName(i, "TotalWin").(int64)))
						rechargeConfig.RewardAmount = proto.Int32(int32(tbRechargeInfo.GetValueByColName(i, "Reward").(int64)))
						rechargeInfoRet.RechargeConfig = append(rechargeInfoRet.RechargeConfig, rechargeConfig)
					}
				}
				retmsg.MsgData = rechargeInfoRet
			} else {
				rechargeInfoRet.Code = proto.Int32(-1)
				rechargeInfoRet.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
				retmsg.MsgData = rechargeInfoRet
			}
		}
		return []*msg.Message{retmsg}
	}

	if m.SClassID == int32(netproto.HallMsgClassID_ActivitGetRechargeReward) { //領取充值活動獎勵
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharOutput("chvErrMsg", "")

		proc := db.NewProcedure("PrPs_ReceiveRechargeActivity", ps)
		ret, err := d.ExecProc(proc)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_ActivitGetRechargeRewardRet)
		rechargeInfoRet := &netproto.RechargeActivityRet{}
		if err != nil {
			ctx.Error("執行存儲過程PrPs_ReceiveRechargeActivity錯誤, %v", err)
			rechargeInfoRet.Code = proto.Int32(0)
			rechargeInfoRet.Message = proto.String("服務器錯誤.")
			retmsg.MsgData = rechargeInfoRet
		} else {
			if ret.GetReturnValue() == 1 {
				rechargeInfoRet.Code = proto.Int32(1)
			} else {
				rechargeInfoRet.Code = proto.Int32(-1)
			}
			rechargeInfoRet.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
			retmsg.MsgData = rechargeInfoRet
		}
		return []*msg.Message{retmsg}
	}

	return nil
}
