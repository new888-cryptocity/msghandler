package msghandler

import (
	"666.com/gameserver/netproto"
)

//CD21 新增/更新系統水池输赢记录
type BJSysPoolHandler struct{}

func (fsp *BJSysPoolHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Game) && sclsid == int32(netproto.BJ_GameMessageClassID_BJSysPoolID)
}

/*
func (fsp *BJSysPoolHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	//傳送syspool data
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJSysPoolData{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BJSysPoolData)
	ps := db.NewSqlParameters()
	ps.AddIntInput("intFlagID", mm.GetFlagID())
	ps.AddBigIntInput("lngBetMoney", mm.GetSumBetMoney())
	ps.AddBigIntInput("lngBetCounts", mm.GetSumBetCount())
	ps.AddBigIntInput("lngPoolAmount", mm.GetPoolAmount())
	proc := db.NewProcedure("PrGs_SuperTool_TwentyOne_UpdSystemPoolWinLost", ps)
	//mlog.System("-------:",proc.GetSqlString())
		ret, err := DbUtil.ExecProc(proc)
	retcode := int32(ret.GetReturnValue())

	if err != nil {
		ctx.Error("执行存储过程[PrGs_SuperTool_TwentyOne_UpdSystemPoolWinLost]错误, %v, %s", err, proc.GetSqlString())
		retcode = int32(0)
	}

	var retmsg *msg.Message
	retmsg = msg.NewMessage(int32(netproto.MessageBClassID_Game), int32(netproto.BJ_GameMessageClassID_BJSysPoolID), nil)

	//get syspooldata
	detail := new(netproto.BJSysPoolData)
	if retcode == 1 {
		if ret.GetRetTableCount() > 0 {
			tbData := ret[0]
			for i := 0; i < len(tbData.Rows); i++ {
				detail.FlagID = proto.Int32(int32(tbData.GetValueByColName(i, "FlagID").(int64)))
				detail.SumBetMoney = proto.Int64(tbData.GetValueByColName(i, "BetMoney").(int64))
				detail.SumBetCount = proto.Int64(tbData.GetValueByColName(i, "BetCounts").(int64))
				detail.PoolAmount = proto.Int64(tbData.GetValueByColName(i, "PoolAmount").(int64))
				mlog.Debug("[linda] flag[%v] sum[%v]", detail.GetFlagID(), detail.GetPoolAmount())
				break
			}
		}
	}

	retmsg.MsgData = detail

	return []*msg.Message{retmsg}
}
*/
