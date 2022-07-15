package msghandler

import (
	"666.com/gameserver/dbserver/src/dal/utility"
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	"github.com/golang/protobuf/proto"
)

type OtherMsg struct {
}

func (h *OtherMsg) Init() {}
func (h *OtherMsg) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_Hall) {
		return false
	}

	switch sclsid {
	case int32(netproto.HallMsgClassID_UpdLanguageID),
		int32(netproto.HallMsgClassID_RequestUserLoginList):
		return true
	}
	return false
}

func (h *OtherMsg) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	switch m.SClassID {
	case int32(netproto.HallMsgClassID_UpdLanguageID):
		return h.OnUpdLanguage(m)
	case int32(netproto.HallMsgClassID_RequestUserLoginList):
		return h.RequestUserLoginList(m)
	}
	return nil
}

func (h *OtherMsg) OnUpdLanguage(m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UpdLanguage{}
		return mm
	})

	if rmm == nil {
		return nil
	}
	mm := rmm.(*netproto.UpdLanguage)

	userid := mm.GetUserID()
	language := mm.GetLanguage()

	code, retmessage := ExecUpdLanguage(userid, language)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_UpdLanguageIDRet)
	retmsg.MsgData = getRetMsg(code, retmessage)
	return []*msg.Message{retmsg}
}

func ExecUpdLanguage(userid int32, language string) (int32, string) {
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharInput("chrLanguage", language, 16)
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrPs_UpdateUserLanguage", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("PrPs_UpdateUserLanguage 錯誤%v", err)
		return 10, "exec proc update language error"
	}

	code := int32(ret.GetReturnValue())
	if code != 1 {
		mlog.Error("執行存儲過程PrPs_UpdateUserLanguage出錯, errcode:%v", code)
	}
	retmessage := ret.GetOutputParamValue("chvErrMsg").(string)
	return code, retmessage
}

func (h *OtherMsg) RequestUserLoginList(m *msg.Message) []*msg.Message {
	rmm, ok := m.MsgData.(*rpcmsg.RPCMessage)
	if !ok {
		return nil
	}

	ip := rmm.RPCInfo.GetIPAddress()
	d := GetDatabase()
	sp := "Hall_PrPs_GetAllUserInfo"
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("varIP", ip, 15)

	proc := db.NewProcedure(sp, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("RequestUserLoginList Error! " + err.Error())
		return nil
	}

	utility.ParserUserInfoList(ret)

	users := utility.ParserUserInfoList(ret)
	data := &netproto.UserLoginList{
		Users: users,
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_RequestUserLoginList)
	retmsg.MsgData = data
	return []*msg.Message{retmsg}

}
