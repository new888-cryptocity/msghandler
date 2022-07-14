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

type GetReelsControlHandler struct {
	sync.RWMutex
}

func (h *GetReelsControlHandler) Init() {}
func (h *GetReelsControlHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_GameRoom) {
		return false
	}

	switch sclsid {
	case int32(netproto.GameRoomClassID_GetReelsControlCYZS): // 讀取風控輪帶
		//mlog.Info("[GetReelsControlHandler][IsHook] hook了 bclsid:%d sclsid:%d", bclsid, sclsid)
		return true
	case int32(netproto.GameRoomClassID_GetReelsControlCSFFF): // 讀取風控輪帶
		//mlog.Info("[GetReelsControlHandler][IsHook] hook了 bclsid:%d sclsid:%d", bclsid, sclsid)
		return true
	}
	return false
}

func (h *GetReelsControlHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[GetReelsControlHandler][OnMessage] BClassID:%d SClassID:%d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.GameRoomClassID_GetReelsControlCYZS): // 讀取風控輪帶
		return h.OnGetReelsControlCYZS(ctx, clt, m)
	case int32(netproto.GameRoomClassID_GetReelsControlCSFFF): // 讀取風控輪帶
		return h.OnGetReelsControlCSFFF(ctx, clt, m)
	}
	return nil
}

//讀取輪帶控製請求
func (h *GetReelsControlHandler) OnGetReelsControlCYZS(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	h.Lock()
	defer h.Unlock()
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_GetReelsControlCYZS)
	mlog.Info("retmsg = %v", retmsg)

	d := GetDatabase()
	ps := db.NewSqlParameters()

	proc := db.NewProcedure("PrGs_Game_GetReelsControl", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrGs_Game_GetReelsControl error")
		return []*msg.Message{retmsg}
	}

	ControlData := new(netproto.ReelsControlInfo)

	info := ret[0]
	mlog.Debug("執行存儲過程 PrGs_Game_GetReelsControl ret[0]:%v", ret[0])
	if ret.GetRetTableCount() > 0 {
		tbData := ret[0]
		for i := 0; i < len(tbData.Rows); i++ {
			if info.GetValueByColName(i, "GameID").(int64) == 49 {

				signConfig := new(netproto.ReelsControl)

				signConfig.ID = proto.Int32(int32(i))
				signConfig.Rtprange = proto.String(info.GetValueByColName(i, "Rtprange").(string))
				signConfig.Checksym = proto.Int64(info.GetValueByColName(i, "Checksym").(int64))
				signConfig.Rtpreel = proto.Int64(info.GetValueByColName(i, "Rtpreel").(int64))
				signConfig.IsCtrl = proto.Int64(info.GetValueByColName(i, "IsCtrl").(int64))

				ControlData.ReelsData = append(ControlData.ReelsData, signConfig)
			}
		}
	}
	mlog.Info("ControlData = %+v", ControlData)

	retmsg.MsgData = ControlData

	mlog.Info("retmsg = %v", []*msg.Message{retmsg})
	return []*msg.Message{retmsg}
}

//讀取輪帶控製請求
func (h *GetReelsControlHandler) OnGetReelsControlCSFFF(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	h.Lock()
	defer h.Unlock()
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_GameRoom)
	retmsg.SClassID = int32(netproto.GameRoomClassID_GetReelsControlCSFFF)
	mlog.Info("retmsg = %v", retmsg)

	d := GetDatabase()
	ps := db.NewSqlParameters()

	proc := db.NewProcedure("PrGs_Game_GetReelsControl", ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("PrGs_Game_GetReelsControl error")
		return []*msg.Message{retmsg}
	}

	ControlData := new(netproto.ReelsControlInfo)

	info := ret[0]
	mlog.Debug("執行存儲過程 PrGs_Game_GetReelsControl ret[0]:%v", ret[0])
	if ret.GetRetTableCount() > 0 {
		tbData := ret[0]
		for i := 0; i < len(tbData.Rows); i++ {

			if info.GetValueByColName(i, "GameID").(int64) == 42 {
				signConfig := new(netproto.ReelsControl)

				signConfig.ID = proto.Int32(int32(i))
				signConfig.Rtprange = proto.String(info.GetValueByColName(i, "Rtprange").(string))
				signConfig.Checksym = proto.Int64(info.GetValueByColName(i, "Checksym").(int64))
				signConfig.Rtpreel = proto.Int64(info.GetValueByColName(i, "Rtpreel").(int64))
				signConfig.IsCtrl = proto.Int64(info.GetValueByColName(i, "IsCtrl").(int64))

				ControlData.ReelsData = append(ControlData.ReelsData, signConfig)
			}
		}
	}
	mlog.Info("ControlData = %+v", ControlData)

	retmsg.MsgData = ControlData

	mlog.Info("retmsg = %v", []*msg.Message{retmsg})
	return []*msg.Message{retmsg}
}
