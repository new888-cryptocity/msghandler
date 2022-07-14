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

type ReliefHandler struct {
}

func (rh *ReliefHandler) Init() {}
func (rh *ReliefHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_Hall) {
		return false
	}

	switch sclsid {
	case int32(netproto.HallMsgClassID_ReliefConfigReqID),
		int32(netproto.HallMsgClassID_ReliefCollectReqID):
		return true
	}
	return false
}

func (rh *ReliefHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.HallMsgClassID_ReliefConfigReqID):
		return rh.OnReliefConfigReq(ctx, clt, m)
	case int32(netproto.HallMsgClassID_ReliefCollectReqID):
		return rh.OnReliefCollectReq(ctx, clt, m)
	}
	return nil
}

func (rh *ReliefHandler) OnReliefConfigReq(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, _ := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &empty.Empty{}
		return mm
	})

	userid := rpcinfo.GetUserID()
	if userid <= 0 {
		mlog.Error("救濟金請求失敗 %v", userid)
		return nil
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharOutput("chvErrMsg", "")
	proc := db.NewProcedure("PrPs_Activity_RelifGetConfig", ps)
	ret, err := d.ExecProc(proc)
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_ReliefConfigRetID)
	if err != nil {
		ctx.Error("執行存儲過程PrPs_Activity_RelifGetConfig錯誤, %v", err)
		return nil
	} else {
		if ret.GetReturnValue() != 1 {
			return nil
		}
		data := &netproto.ReliefConfigRet{}
		data.Reward = proto.Int32(ret[0].GetSingleValueInt32("Reward"))
		data.LessThanMoney = proto.Int32(ret[0].GetSingleValueInt32("LimitMoney"))
		data.MaxCollectTimes = proto.Int32(ret[0].GetSingleValueInt32("MaxTimes"))
		data.DayCollectTimes = proto.Int32(ret[0].GetSingleValueInt32("CollectTimes"))
		data.Desc = proto.String(ret[0].GetSingleValue("ReclifDesc").(string))
		retmsg.MsgData = data
	}
	return []*msg.Message{retmsg}
}

func (rh *ReliefHandler) OnReliefCollectReq(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, _ := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &empty.Empty{}
		return mm
	})

	userid := rpcinfo.GetUserID()
	if userid <= 0 {
		mlog.Error("救濟金請求失敗 %v", userid)
		return nil
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrPs_Activity_RelifCollect", ps)
	ret, err := d.ExecProc(proc)
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_ReliefCollectRetID)
	if err != nil {
		ctx.Error("執行存儲過程PrPs_Activity_RelifCollect錯誤, %v", err)
		return nil
	} else {
		data := &netproto.ReliefCollectRet{}
		if ret.GetReturnValue() != 1 {
			data.Code = proto.Int32(0)
			data.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))
			retmsg.MsgData = data
			return []*msg.Message{retmsg}
		}
		data.Code = proto.Int32(1)
		data.Reward = proto.Int32(ret[0].GetSingleValueInt32("Reward"))
		data.DayCollectTimes = proto.Int32(ret[0].GetSingleValueInt32("CollectTimes"))
		retmsg.MsgData = data
	}
	return []*msg.Message{retmsg}
}
