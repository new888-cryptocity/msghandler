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
)

type CpThirdGameHandler struct {
}

func (cp *CpThirdGameHandler) Init() {}
func (cp *CpThirdGameHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_DBServer) {
		return false
	}
	switch sclsid {
	case int32(netproto.DBServerClassID_GetCpGameInfo),
		int32(netproto.DBServerClassID_GetUserScore),
		int32(netproto.DBServerClassID_CpGameTransferScore),
		int32(netproto.DBServerClassID_CpGameCheckOrderState):
		return true
	}
	return false
}

func (cp *CpThirdGameHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	switch m.SClassID {
	case int32(netproto.DBServerClassID_GetCpGameInfo):
		return cp.GetCpGameConfig(m)
	case int32(netproto.DBServerClassID_GetUserScore):
		return cp.GetUserBalanceReq(m)
	case int32(netproto.DBServerClassID_CpGameTransferScore):
		return cp.UserTransferReq(m)
	case int32(netproto.DBServerClassID_CpGameCheckOrderState):
		return cp.GetUserTransferState(m)
	}
	return nil
}

//獲取商戶配置信息
func (cp *CpThirdGameHandler) GetCpGameConfig(m *msg.Message) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.DBServerClassID_GetCpGameInfoRet)
	gameConfigList := new(netproto.CpGameConfig)
	retmsg.MsgData = gameConfigList
	d := GetDatabase()
	ps := db.NewSqlParameters()
	proc := db.NewProcedure("PrGs_GetCpGameConfig", ps)
	res, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("執行存儲過程PrGs_GetCpGameConfig失敗,err:%v", err)
		for i := 0; err != nil && i < 10; i++ {
			mlog.Error("執行存儲過程PrGs_GetCpGameConfig失敗,第:%d次 err:%v", i, err)
			res, err = d.ExecProc(proc) //重試一次
			if err == nil {
				mlog.Error("執行存儲過程PrGs_GetCpGameConfig成功")
				break
			}
		}
	}

	if err == nil && res[0] != nil {
		info := res[0]
		for i := 0; i < len(info.Rows); i++ {
			gameConfig := new(netproto.CpGameInfo)
			gameConfig.ID = proto.Int32(int32(info.GetValueByColName(i, "ID").(int64)))
			gameConfig.Name = proto.String(info.GetValueByColName(i, "Name").(string))
			gameConfig.BusinessKey = proto.String(info.GetValueByColName(i, "BusinessKey").(string))
			gameConfig.SignKey = proto.String(info.GetValueByColName(i, "SignKey").(string))
			gameConfig.Part = proto.Int32(int32(info.GetValueByColName(i, "Part").(int64)))
			gameConfig.GroupID = proto.Int32(int32(info.GetValueByColName(i, "GroupID").(int64)))
			gameConfig.PlatformID = proto.Int32(int32(info.GetValueByColName(i, "PlatformID").(int64)))
			gameConfig.HostAddress = proto.String(info.GetValueByColName(i, "HostAddress").(string))
			gameConfigList.Res = append(gameConfigList.Res, gameConfig)
		}
	}
	return []*msg.Message{retmsg}
}

//查詢玩家的金幣庫存
func (cp *CpThirdGameHandler) GetUserBalanceReq(m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ReqUserScore{}
		return mm
	})
	mm := rmm.(*netproto.ReqUserScore)
	BusinessKey := mm.GetBusinessKey()
	UserID := mm.GetUserID()
	mlog.Info("BusinessKey:%v, UserID:%v", BusinessKey, UserID)
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.DBServerClassID_GetUserScoreRet)
	data := new(netproto.ResUserScore)
	data.UserID = proto.Int32(UserID)
	data.Score = proto.Int64(0)
	retmsg.MsgData = data
	d := GetDatabase()
	ps := db.NewSqlParameters()
	var errCode int32 = 0
	ps.AddIntInput("intUserID", UserID)
	ps.AddIntOutput("intErrCode", errCode)
	ps.AddVarcharOutput("chvErrMsg", "")
	proc := db.NewProcedure("PrGs_GetUserMoney", ps)
	res, err := d.ExecProc(proc)
	if err != nil {
		return []*msg.Message{retmsg}
	}
	data.ErrMsg = proto.String(res.GetOutputParamValue("chvErrMsg").(string))
	data.ErrCode = proto.Int32(int32(res.GetOutputParamValue("intErrCode").(int64)))
	if res.GetReturnValue() > 0 && res[0] != nil {
		info := res[0]
		for i := 0; i < len(info.Rows); i++ {
			data.Score = proto.Int64(int64(info.GetValueByColName(i, "Amount").(int64)))
		}
	}
	return []*msg.Message{retmsg}
}

//CP通路請求存取玩家的金幣
func (cp *CpThirdGameHandler) UserTransferReq(m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.CpGameUserTransferScore{}
		return mm
	})
	mm := rmm.(*netproto.CpGameUserTransferScore)

	msgid := rpc.GetQueueID()
	BusinessKey := mm.GetBusinessKey()
	UserID := mm.GetUserID()
	OrderID := mm.GetOrderID()
	TransferType := mm.GetTransferType()
	Amount := mm.GetAmount()
	ServerID := mm.GetServerID()
	createTime := mm.GetCreateTime()
	mlog.Info("BusinessKey:%v, UserID:%v", BusinessKey, UserID)
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.DBServerClassID_CpGameTransferScoreRet)
	data := new(netproto.CpGameUserTransferRes)
	data.UserID = proto.Int32(UserID)
	data.BusinessKey = proto.String(BusinessKey)
	data.OrderID = proto.String(OrderID)
	data.TransferType = proto.Int32(TransferType)
	data.Score = proto.Int64(0)
	data.UpdateAmount = proto.Int64(Amount)
	retmsg.MsgData = data
	d := GetDatabase()
	ps := db.NewSqlParameters()
	var errCode int32 = 0
	ps.AddIntInput("intMsgID", msgid)
	ps.AddIntInput("intUserID", UserID)
	ps.AddBigIntInput("lngUpdateAmount", Amount)
	ps.AddBigIntInput("lngTaxAmount", 0)
	ps.AddIntInput("intServerId", ServerID)
	ps.AddVarcharInput("chvBusinessKey", BusinessKey, 255)
	ps.AddVarcharInput("chvOrderID", OrderID, 255)
	ps.AddIntInput("intTransferType", TransferType)
	ps.AddVarcharInput("dtmCreateTime", createTime, 20)
	ps.AddIntOutput("intErrCode", errCode)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddBigIntOutput("lngCurrentAmount", 0)

	proc := db.NewProcedure("PrGs_CpGameTransferUserMoney", ps)
	res, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("執行存儲過程PrGs_CpGameTransferUserMoney失敗,err:%v", err)
		return []*msg.Message{retmsg}
	}

	data.ErrMsg = proto.String(res.GetOutputParamValue("chvErrMsg").(string))
	data.ErrCode = proto.Int32(int32(res.GetOutputParamValue("intErrCode").(int64)))

	if res.GetReturnValue() > 0 && res[0] != nil {
		data.Score = proto.Int64(res.GetOutputParamValue("lngCurrentAmount").(int64))
	}

	return []*msg.Message{retmsg}
}

//查詢玩家的訂單狀態
func (cp *CpThirdGameHandler) GetUserTransferState(m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.CpCheckOderState{}
		return mm
	})
	mm := rmm.(*netproto.CpCheckOderState)
	BusinessKey := mm.GetBusinessKey()
	OrderID := mm.GetOrderID()
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.DBServerClassID_CpGameCheckOrderStateRet)
	data := new(netproto.CpCheckOderStateRes)
	data.BusinessKey = proto.String(BusinessKey)
	data.OrderID = proto.String(OrderID)
	retmsg.MsgData = data
	d := GetDatabase()
	ps := db.NewSqlParameters()
	var errCode int32 = 0
	ps.AddVarcharInput("chvOrderID", OrderID, 255)
	ps.AddIntOutput("intErrCode", errCode)
	ps.AddVarcharOutput("chvErrMsg", "")
	proc := db.NewProcedure("PrGs_GetCpGameOrderInfo", ps)
	res, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("執行存儲過程PrGs_GetCpGameOrderInfo失敗,err:%v", err)
		return []*msg.Message{retmsg}
	}

	data.ErrMsg = proto.String(res.GetOutputParamValue("chvErrMsg").(string))
	data.ErrCode = proto.Int32(int32(res.GetOutputParamValue("intErrCode").(int64)))

	if res.GetReturnValue() > 0 && res[0] != nil {
		info := res[0]
		for i := 0; i < len(info.Rows); i++ {
			data.State = proto.Int32(int32(info.GetValueByColName(i, "State").(int64)))
		}
	}

	return []*msg.Message{retmsg}
}
