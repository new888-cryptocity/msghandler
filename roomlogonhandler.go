package msghandler

import (
	"encoding/json"
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	"github.com/golang/protobuf/proto"
)

//登入房間(進入遊戲伺服器)
type RoomLogonHandler struct {
}

func (h *RoomLogonHandler) Init() {}
func (h *RoomLogonHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_GameRoom) && sclsid == int32(netproto.GameRoomClassID_LoginRoomID)
}

func (h *RoomLogonHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LoginGameRoomInfo{}
		return mm
	})

	mm := rmm.(*netproto.LoginGameRoomInfo)

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	ctx.Info("RoomLogonHandler::OnMessage user logon room:%v", mm)

	//取得傳入參數
	serverid := rpc.GetRouteServerID()
	userid := mm.GetUserID()
	ip := rpc.GetIPAddress()
	cer := mm.GetCer()
	hdtype := mm.GetHDType()
	hdcode := mm.GetHDCode()

	//準備呼叫存儲過程參數
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("sintServerID", serverid)
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddVarcharInput("chvCer", cer, 32)
	ps.AddIntInput("tnyHDType", hdtype)
	ps.AddVarcharInput("chvHDCode", hdcode, 64)
	ps.AddVarcharOutput("chvErrMsg", "")

	mlog.Debug("執行存儲過程 PrGs_UserLogin serverid:[%d] userid:[%d] hdtype:[%d] hdcode[%s] ip[%s] cer[%s]",
		serverid, userid, hdtype, hdcode, ip, cer)

	proName := "PrGs_UserLogin"
	mlog.Debug("[%s] serverid=%v", proName, serverid)
	mlog.Debug("[%s] userid=%v", proName, userid)
	mlog.Debug("[%s] ip=%v", proName, ip)
	mlog.Debug("[%s] cer=%v", proName, cer)
	mlog.Debug("[%s] hdtype=%v", proName, hdtype)
	mlog.Debug("[%s] hdcode=%v", proName, hdcode)

	//執行存儲過程
	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	//組合回傳封包
	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_GameRoom), int32(netproto.GameRoomClassID_LoginRoomRetID), nil)

	if err != nil {
		mlog.Error("執行存儲過程 %s 出錯%v", proName, err)
		logret := &netproto.LoginGameRoomRet{}
		logret.Code = proto.Int32(0)
		logret.Message = proto.String(fmt.Sprintf("服務器錯誤."))
		retmsg.MsgData = logret

	} else {
		var code int64 = 0
		if ret.GetReturnValue() == 1 {
			code = 1
		} else {
			code = ret.GetReturnValue()
		}

		mlog.Debug("執行存儲過程 %s code:%d, TableCount:%d", proName, code, ret.GetRetTableCount())

		logret := &netproto.LoginGameRoomRet{}
		logret.Code = proto.Int32(int32(code))
		logret.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
		logret.LoginRequestData = mm

		// for i, j := range ret {
		// 	mlog.Debug("看細節 i:%v, Rows:%v, j:%v", i, j.Rows, j)
		// }

		//if ret.GetRetTableCount() > 1 { // old 用這會不太準 因為有錯誤代碼回傳時 資料會是 [遊戲GameID無效]] [chvErrMsg] + [10141]] []
		if code == 1 { // leo  修改, 因為看SP, 成功是回傳1, 其他錯誤都是 大於0的數

			userdata := new(netproto.UserRoomLogonData)

			userdata.UserID = proto.Int32(ret[0].GetSingleValueInt32("UserID"))
			userdata.UserType = proto.Int32(ret[0].GetSingleValueInt32("UserType"))
			userdata.NickName = proto.String(ret[0].GetSingleValue("NickName").(string))
			userdata.FaceID = proto.Int32(ret[0].GetSingleValueInt32("FaceID"))
			userdata.Sex = proto.Int32(ret[0].GetSingleValueInt32("Sex"))

			userdata.Currency = proto.String(ret[0].GetSingleValue("Currency").(string))    //幣值
			userdata.Denomination = proto.Int32(ret[0].GetSingleValueInt32("Denomination")) //放大倍率
			userdata.CashAmount = proto.Int64(ret[0].GetSingleValueInt64("CashAmount"))     //現金點數
			userdata.FreePoint = proto.Int64(ret[0].GetSingleValueInt64("FreePoint"))       //虛點

			userdata.WinCount = proto.Int32(ret[0].GetSingleValueInt32("WinCount"))
			userdata.LoseCount = proto.Int32(ret[0].GetSingleValueInt32("LoseCount"))
			userdata.DrawCount = proto.Int32(ret[0].GetSingleValueInt32("DrawCount"))
			userdata.ServerAddr = proto.String(ret[0].GetSingleValue("ServerAddr").(string))
			userdata.GameBuff = proto.String(ret[0].GetSingleValue("GameBuff").(string))
			userdata.TotalScore = proto.Int64(ret[0].GetSingleValueInt64("TotalScore"))
			userdata.XiuXianScore = proto.Int64(ret[0].GetSingleValueInt64("XiuXianScore"))
			userdata.IsSuperUser = proto.Bool(ret[0].GetSingleValueInt64("IsSuperUser") == 1)
			userdata.TracedUserID = proto.Int32(ret[0].GetSingleValueInt32("TracedUserID"))
			userdata.TodayScoreDan = proto.Int32(ret[0].GetSingleValueInt32("TodayScoreDan"))
			userdata.TotalWinDan = proto.Int32(ret[0].GetSingleValueInt32("TotalWinDan"))
			userdata.ChargeDan = proto.Int32(ret[0].GetSingleValueInt32("ChargeDan"))
			userdata.WinRateDan = proto.Int32(ret[0].GetSingleValueInt32("WinRateDan"))
			userdata.GameTimeDan = proto.Int32(ret[0].GetSingleValueInt32("GameTimeDan"))
			userdata.IsNewBee = proto.Bool(ret[0].GetSingleValueInt64("IsNewBee") == 1)

			userdata.IsNewbiePro = proto.Bool(ret[0].GetSingleValueInt32("IsNewbiePro") == 1)
			userdata.NextWinRate = proto.Int32(ret[0].GetSingleValueInt32("NextWinRate"))
			userdata.NextWinLimit = proto.Int32(ret[0].GetSingleValueInt32("NextWinLimit"))
			userdata.NextAddRate = proto.Int32(ret[0].GetSingleValueInt32("NextAddRate"))
			userdata.DiffRunAcc = proto.Int64(ret[0].GetSingleValueInt64("DiffRunAcc"))
			userdata.WinLoseParam = proto.Int32(ret[0].GetSingleValueInt32("WinLoseParam"))
			userdata.TotalWinLose = proto.Int64(ret[0].GetSingleValueInt64("TotalWinLose"))
			userdata.TotalCharge = proto.Int64(ret[0].GetSingleValueInt64("TotalCharge"))
			userdata.BetLimit = proto.Int64(ret[0].GetSingleValueInt64("BetLimit"))
			userdata.BankMoney = proto.Int64(ret[0].GetSingleValueInt64("BankMoney")) //銀行點數

			userdata.CurrGameLoseWin = proto.Int64(ret[0].GetSingleValueInt64("CurrGameLoseWin"))       // 當前遊戲總輸贏
			userdata.FirstGameTimestamp = proto.Int64(ret[0].GetSingleValueInt64("FirstGameTimestamp")) // 首次玩該遊戲時間戳
			userdata.DayRecharge = proto.Int64(ret[0].GetSingleValueInt64("DayRecharge"))               // 日充值

			mlog.Debug("執行存儲過程 %s [進入遊戲伺服器結果] UserID:%d, UserType:%d, Currency:%s, Denomination:%d, CashAmount:%d, BankMoney:%d",
				proName, *userdata.UserID, *userdata.UserType, *userdata.Currency, *userdata.Denomination, *userdata.CashAmount, *userdata.BankMoney)

			logret.UserData = userdata
			tv, ts := json.Marshal(userdata)
			mlog.Info("執行存儲過程 %s 成功 %v %v", proName, string(tv), ts)
		} else {
			mlog.Warn("執行存儲過程 %s 失敗 Code:%v, Message:%v", proName, *logret.Code, *logret.Message)
		}

		retmsg.MsgData = logret
	}

	return []*msg.Message{retmsg}
}
