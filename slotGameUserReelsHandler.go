package msghandler

import (
	"sync"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

type GetUserReelsControlHandler struct {
	sync.RWMutex
}

func (h *GetUserReelsControlHandler) Init() {}
func (h *GetUserReelsControlHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_GameRoom) {
		return false
	}

	switch sclsid {
	case int32(netproto.GameRoomClassID_GetUserReelsControl): // 讀取風控輪帶
		//mlog.Info("[GetUserReelsControlHandler][IsHook] hook了 bclsid:%d sclsid:%d", bclsid, sclsid)
		return true
	}
	return false
}

func (h *GetUserReelsControlHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[GetUserReelsControlHandler][OnMessage] BClassID:%d SClassID:%d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.GameRoomClassID_GetUserReelsControl): // 讀取風控輪帶
		return h.OnGetUserReelsControl(ctx, clt, m)
	}
	return nil
}

//讀取輪帶控製請求
func (h *GetUserReelsControlHandler) OnGetUserReelsControl(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	h.Lock()
	defer h.Unlock()
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_GetUserReelsControl)
	mlog.Info("retmsg = %v", retmsg)
	// data := new(netproto.SignActionConfigRet)

	d := GetDatabase()
	ps := db.NewSqlParameters()

	proc := db.NewProcedure("PrGs_Game_GetUserReelsControl", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrGs_Game_GetUserReelsControl error")
		return []*msg.Message{retmsg}
	}

	ControlData := new(netproto.UserReelsControlInfo)

	info := ret[0]
	mlog.Debug("執行存儲過程 PrGs_Game_GetUserReelsControl ret[0]:%v", ret[0])
	if ret.GetRetTableCount() > 0 {
		tbData := ret[0]
		for i := 0; i < len(tbData.Rows); i++ {
			signConfig := new(netproto.UserReelsControl)

			signConfig.UserID = proto.Int64(info.GetValueByColName(i, "UserID").(int64))
			signConfig.IsControl = proto.Int64(info.GetValueByColName(i, "IsControl").(int64))

			ControlData.UserReelsData = append(ControlData.UserReelsData, signConfig)
		}
	}
	mlog.Info("ControlData = %+v", ControlData)

	retmsg.MsgData = ControlData

	mlog.Info("retmsg = %v", []*msg.Message{retmsg})
	return []*msg.Message{retmsg}
}
