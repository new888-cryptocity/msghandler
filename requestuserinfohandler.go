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

//請求用戶大廳信息
type RequestUserInfoHandler struct {
}

func (rh *RequestUserInfoHandler) Init() {}
func (h *RequestUserInfoHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_RequestUserHallInfoID)
}

func (h *RequestUserInfoHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.RequestUserHallInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.RequestUserHallInfo)

	userid := rpc.GetUserID()
	areaid := mm.GetSiteID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntOutput("intUserID", userid)
	ps.AddIntInput("intAreaID", areaid)
	ps.AddVarcharOutput("chvCertification", "")
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrPs_User_GetUserInfo", ps)
	ret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_UserHallInfoID), nil)

	if err != nil || ret.GetRetTableCount() <= 0 {
		mlog.Error("執行存儲過程PrPs_User_GetUserInfo出錯%v", err)
		logret := &netproto.UserLoginRet{}
		logret.Code = proto.Int32(0)
		logret.Message = proto.String("服務器錯誤.")
		retmsg.MsgData = logret

	} else {
		logret := parseUserLogonRet(ret) //PrPs_User_GetUserInfo
		retmsg.MsgData = logret
	}

	return []*msg.Message{retmsg}
}
