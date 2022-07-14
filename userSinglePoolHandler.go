package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"

	proto "github.com/golang/protobuf/proto"
)

type UserSinglePoolHandler struct {
}

func (h *UserSinglePoolHandler) Init() {}
func (h *UserSinglePoolHandler) IsHook(bclsid int32, sclsid int32) bool {
	//return bclsid == int32(netproto.MessageBClassID_GameRoom) && sclsid == int32(netproto.GameRoomClassID_GetSinglePool)
	if bclsid != int32(netproto.MessageBClassID_GameRoom) {
		return false
	}

	//mlog.Info("[UserSinglePoolHandler][IsHook] 準備hook bclsid:%d sclsid:%d", bclsid, sclsid)

	switch sclsid {
	case int32(netproto.GameRoomClassID_GetSinglePool), //用戶個人水池數據請求
		int32(netproto.GameRoomClassID_UpdateSinglePool): //用戶個人水池數據更新 (如果冇資料會寫入一筆新的)
		//mlog.Info("[UserSinglePoolHandler][IsHook] hook了 bclsid:%d sclsid:%d", bclsid, sclsid)
		return true
	}
	return false
}

func (h *UserSinglePoolHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[UserSinglePoolHandler][OnMessage] BClassID:%d SClassID:%d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.GameRoomClassID_GetSinglePool): //用戶個人水池數據請求
		return h.OnGetSinglePool(ctx, clt, m)
	case int32(netproto.GameRoomClassID_UpdateSinglePool): //用戶個人水池數據更新
		return h.OnUpdateSinglePool(ctx, clt, m)
	}
	return nil
}

//用戶個人水池數據請求
func (h *UserSinglePoolHandler) OnGetSinglePool(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GetUserSinglePoolReq{}
		return mm
	})

	if rmm == nil {
		return nil
	}
	mm := rmm.(*netproto.GetUserSinglePoolReq)

	//接收資料
	gameid := mm.GetGameID()       //遊戲ID
	userid := mm.GetUserID()       //玩家ID
	lianyunID := mm.GetLianyunID() //聯運ID

	//回傳格式
	resData := new(netproto.UserSinglePoolRes)
	resData.IsSuccess = proto.Bool(false)
	resData.GameID = proto.Int32(0)
	resData.UserID = proto.Int32(0)
	resData.PoolValue = proto.Int64(0)
	resData.LineNum = proto.Int32(0)
	resData.RTP = proto.Int32(0)
	resData.LianyunID = proto.Int32(0)
	//回傳格式
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_GetSinglePool)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intGameID", gameid)
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intLianyunID", lianyunID)

	mlog.Debug("執行存儲過程 PrGs_Game_GetSinglePool gameid:[%d], userid=[%d], lianyunID:[%d]", gameid, userid, lianyunID)

	proc := db.NewProcedure("PrGs_Game_GetSinglePool", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("執行存儲過程 PrGs_Game_GetSinglePool 出錯%v", err)
		return nil
	}

	if ret.GetReturnValue() != 1 {
		mlog.Error("執行存儲過程 PrGs_Game_GetSinglePool 出錯, errcode:%v", ret.GetReturnValue())
		return nil
	}

	if ret.GetRetTableCount() > 0 && len(ret[0].Rows) > 0 {
		//組合回傳數據
		mlog.Debug("執行存儲過程 PrGs_Game_GetSinglePool ret[0]:%v", ret[0])

		resData.IsSuccess = proto.Bool(true)
		resData.GameID = proto.Int32(ret[0].GetSingleValueInt32("GameID"))
		resData.UserID = proto.Int32(userid)
		resData.PoolValue = proto.Int64(ret[0].GetSingleValueInt64("PoolValue"))
		resData.LineNum = proto.Int32(ret[0].GetSingleValueInt32("LineNum"))
		resData.RTP = proto.Int32(ret[0].GetSingleValueInt32("RTP"))
		resData.LianyunID = proto.Int32(ret[0].GetSingleValueInt32("LianyunID"))
		mlog.Debug("執行存儲過程 PrGs_Game_GetSinglePool 組合回傳數據:%v", resData)

		retmsg.MsgData = resData

	} else {
		resData.IsSuccess = proto.Bool(false)
		retmsg.MsgData = resData
		mlog.Debug("執行存儲過程 PrGs_Game_GetSinglePool 失敗, 查無此資料 gameid:[%d], userid=[%d], lianyunID:[%d]",
			gameid, userid, lianyunID)
	}
	return []*msg.Message{retmsg}
}

//用戶個人水池數據更新
func (h *UserSinglePoolHandler) OnUpdateSinglePool(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UpdateUserSinglePoolReq{}
		return mm
	})

	if rmm == nil {
		return nil
	}
	mm := rmm.(*netproto.UpdateUserSinglePoolReq)

	//接收資料
	gameid := mm.GetGameID()                   //遊戲ID
	userid := mm.GetUserID()                   //玩家ID
	changePoolValue := mm.GetChangePoolValue() //變化水池量
	LineNum := mm.GetLineNum()                 //水位線
	rtp := mm.GetRTP()                         //RTP
	lianyunID := mm.GetLianyunID()             //聯運ID

	//回傳格式
	resData := new(netproto.UserSinglePoolRes)
	resData.IsSuccess = proto.Bool(false)
	resData.GameID = proto.Int32(0)
	resData.UserID = proto.Int32(0)
	resData.PoolValue = proto.Int64(0)
	resData.LineNum = proto.Int32(0)
	resData.RTP = proto.Int32(0)
	resData.LianyunID = proto.Int32(0)
	//回傳格式
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_UpdateSinglePool)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intGameID", gameid)
	ps.AddIntInput("intUserID", userid)

	ps.AddBigIntInput("intChangePoolValue", changePoolValue)
	ps.AddIntInput("intLineNum", LineNum)
	ps.AddIntInput("floatRTP", rtp)
	ps.AddIntInput("intLianyunID", lianyunID)

	mlog.Debug("執行存儲過程 PrGs_Game_UpdateSinglePool gameid:%d, userid:%v, changePoolValue=%d, lianyunID=%d",
		gameid, userid, changePoolValue, lianyunID)

	proc := db.NewProcedure("PrGs_Game_UpdateSinglePool", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("執行存儲過程 PrGs_Game_UpdateSinglePool 出錯%v", err)
		return nil
	}

	if ret.GetReturnValue() != 1 {
		mlog.Error("執行存儲過程 PrGs_Game_UpdateSinglePool 出錯, errcode:%v", ret.GetReturnValue())
		return nil
	}

	if ret.GetRetTableCount() > 0 && len(ret[0].Rows) > 0 {
		//組合回傳數據
		mlog.Debug("執行存儲過程 PrGs_Game_UpdateSinglePool ret[0]:%v", ret[0])

		//組合回傳數據
		resData.IsSuccess = proto.Bool(true)
		resData.GameID = proto.Int32(ret[0].GetSingleValueInt32("GameID"))
		resData.UserID = proto.Int32(userid)
		resData.PoolValue = proto.Int64(ret[0].GetSingleValueInt64("PoolValue"))
		resData.LineNum = proto.Int32(ret[0].GetSingleValueInt32("LineNum"))
		resData.RTP = proto.Int32(ret[0].GetSingleValueInt32("RTP"))
		resData.LianyunID = proto.Int32(ret[0].GetSingleValueInt32("LianyunID"))

		mlog.Debug("執行存儲過程 PrGs_Game_UpdateSinglePool 組合回傳數據:%v", resData)

		retmsg.MsgData = resData

	} else {
		resData.IsSuccess = proto.Bool(false)
		retmsg.MsgData = resData
		mlog.Debug("執行存儲過程 PrGs_Game_UpdateSinglePool 失敗, 查無此資料 gameid:[%d], userid=[%d], lianyunID:[%d]",
			gameid, userid, lianyunID)
	}
	return []*msg.Message{retmsg}
}
