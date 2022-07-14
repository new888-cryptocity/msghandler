package msghandler

import (
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

// 儲值活動
type DepositActivityHandle struct {
}

func (d *DepositActivityHandle) Init() {}
func (d *DepositActivityHandle) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_DBServer) {
		return false
	}
	switch sclsid {
	case int32(netproto.DBServerClassID_DepositActivityGet):
		return true

	}
	return false
}

func (d *DepositActivityHandle) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[DepositActivityHandle][OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.DBServerClassID_DepositActivityGet):
		return d.GetDepositActivityData(ctx, clt, m)
	}
	return nil
}

func (d *DepositActivityHandle) GetDepositActivityData(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LuckyWheelDBRequest{}
		return mm
	})

	mm := rmm.(*netproto.LuckyWheelDBRequest)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_DepositActivityGetRet)

	dbbase := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intGroupID", *mm.GroupID)
	proc := db.NewProcedure("PrGs_ActivityRechargeGet", ps)
	mlog.Info("PrGs_ActivityRechargeGet")
	ret, err := dbbase.ExecProc(proc)
	if err != nil {
		mlog.Error("PrGs_ActivityRechargeGet error")
		return []*msg.Message{retmsg}
	}
	return ParseDepositActivityData(ret, retmsg)
}

func ParseDepositActivityData(ret db.RecordSet, retmsg *msg.Message) []*msg.Message {
	data := new(netproto.DepositActivityData)
	data.Rewards = make([]*netproto.DZPKHALLDepositActivityRewardConfig, 0)
	data.UserData = make([]*netproto.DepositActivityUserData, 0)
	// 獎項設定
	if ret[0] != nil {
		info := ret[0]
		if len(info.Rows) > 0 {
			data.Config = &netproto.DZPKHALLDepositActivityConfig{}
			data.Config.BeginTime = proto.String(info.GetValueByColName(0, "BeginTime").(time.Time).String())
			data.Config.EndTime = proto.String(info.GetValueByColName(0, "EndTime").(time.Time).String())
		}
	}
	// 獎項清單
	if ret[1] != nil {
		info := ret[1]
		for i := 0; i < len(info.Rows); i++ {
			config := new(netproto.DZPKHALLDepositActivityRewardConfig)
			config.ID = proto.Int32(int32(info.GetValueByColName(i, "ID").(int64)))
			config.RewardAmount = proto.Int32(int32(info.GetValueByColName(i, "RewardAmount").(int64)))
			config.BetAmount = proto.Int32(int32(info.GetValueByColName(i, "BetAmount").(int64)))
			config.RechargeAmount = proto.Int32(int32(info.GetValueByColName(i, "RechargeAmount").(int64)))
			data.Rewards = append(data.Rewards, config)
		}
	}
	// 玩家清單
	if ret[2] != nil {
		info := ret[2]
		for i := 0; i < len(info.Rows); i++ {
			udata := new(netproto.DepositActivityUserData)
			udata.UserID = proto.Int32(int32(info.GetValueByColName(i, "UserId").(int64)))
			udata.BetAmount = proto.Int64(info.GetValueByColName(i, "BetAmount").(int64))
			udata.RechargeAmount = proto.Int64(info.GetValueByColName(i, "RechargeAmount").(int64))
			data.UserData = append(data.UserData, udata)
		}
	}
	retmsg.MsgData = data
	return []*msg.Message{retmsg}
}
