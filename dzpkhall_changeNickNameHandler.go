package msghandler

import (
	"regexp"
	"strings"
	"sync"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

//修改暱稱
type DZPKHALL_ChangeNickNameHandler struct {
	blackList []string
	once      sync.Once
	regexp    *regexp.Regexp
}

func (rh *DZPKHALL_ChangeNickNameHandler) Init() {}

func (rh *DZPKHALL_ChangeNickNameHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_DZPKHALL_ChangeNickName)
}

func (rh *DZPKHALL_ChangeNickNameHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rh.once.Do(func() {
		blackListString := ctx.Config.GetConfig("db").GetConfigValue("nicknameblacklist", "words")
		tmpArr := strings.Split(blackListString, ",")
		for _, v := range tmpArr {
			rh.blackList = append(rh.blackList, strings.ToUpper(v))
		}
		rh.regexp, _ = regexp.Compile("[^A-Za-z0-9]")
	})
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DZPKHALLChangeNickName{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.DZPKHALLChangeNickName)

	nickName := strings.ToUpper(mm.GetNickName())
	userID := rpc.GetUserID()

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_DZPKHALL_ChangeNickNameRet), nil)
	logret := &netproto.DZPKHALLChangeNickNameRet{}
	logret.CurrentNickName = &nickName
	logret.CurrentMoney = proto.Int64(0)

	// 暱稱長度檢查
	if len(nickName) < 4 || len(nickName) > 20 {
		logret.Code = proto.Int32(10144)
		logret.Message = proto.String("暱稱長度有問題")
		retmsg.MsgData = logret
		return []*msg.Message{retmsg}
	}

	// 暱稱字元檢查
	if rh.regexp.MatchString(nickName) {
		logret.Code = proto.Int32(10147)
		logret.Message = proto.String("暱稱有不合法字元")
		retmsg.MsgData = logret
		return []*msg.Message{retmsg}
	}

	// 禁用字檢查
	for i := 0; i < len(rh.blackList); i++ {
		if strings.Index(nickName, rh.blackList[i]) != -1 {
			logret.Code = proto.Int32(10145)
			logret.Message = proto.String("暱稱有禁用字")
			retmsg.MsgData = logret
			return []*msg.Message{retmsg}
		}
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvNickName", nickName, 255)
	ps.AddIntInput("intUserID", userID)
	ps.AddBigIntOutput("lngCurrentAmount", 0)
	ps.AddVarcharOutput("chvDBName", "")
	ps.AddIntOutput("intNextCost", 0)

	//ps.AddVarcharOutput("chvErrMsg", "")

	proName := "DZPKHALL_PrPs_ChangeNickName"
	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("執行存儲過程%s出錯%v", proName, err)
		logret.Code = proto.Int32(0)
		logret.Message = proto.String("服務器錯誤.")
		retmsg.MsgData = logret

	} else {
		mlog.Debug("[%s] 執行成功 ret=%v", proName, ret)
		code := ret.GetReturnValue()
		logret.Code = proto.Int32(int32(code))
		logret.CurrentMoney = proto.Int64(ret.GetOutputParamValue("lngCurrentAmount").(int64))
		logret.CurrentNickName = proto.String(ret.GetOutputParamValue("chvDBName").(string))
		logret.NextCost = proto.Int32(int32(ret.GetOutputParamValue("intNextCost").(int64)))
		retmsg.MsgData = logret
	}

	return []*msg.Message{retmsg}
}
