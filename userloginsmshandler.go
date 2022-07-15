package msghandler

import (
	"time"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"

	proto "github.com/golang/protobuf/proto"
)

func NewUserLoginSmsHandler(hook func(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) bool) *UserLoginSmsHandler {
	handle := &UserLoginSmsHandler{}
	handle.hook = hook
	return handle
}

//用户登录
type UserLoginSmsHandler struct {
	hook func(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) bool
}

func (h *UserLoginSmsHandler) Init() {}
func (h *UserLoginSmsHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_DZPKHALL_LoginSms)
}

func (h *UserLoginSmsHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UserLoginSms{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	sid := netproto.HallMsgClassID_DZPKHALL_LoginSmsRet

	mm := rmm.(*netproto.UserLoginSms)

	hdtype := mm.GetHDType()
	hdcode := mm.GetHDCode()
	ip := mm.GetClientIP()
	if ip == "" {
		ip = "unknow"
	}
	siteid := mm.GetSiteID()
	version := mm.GetVersion()
	serverid := rpc.GetRouteServerID()

	loginname := mm.GetLoginName()
	password := mm.GetPassword()

	bunldID := mm.GetBunldID()
	countryCode := mm.GetCountryCode()
	ver := mm.GetVer()
	vcode := mm.GetVCode()

	var userid int32

	mlog.Debug(`[DbServer] 處理用户登入sms請求開始 loginname=%s, password=%s, version=%s, hdtype=%d, hdcode=%s, serverid=%d, ip=%s
		siteid=%d, bunldID=%s`,
		loginname, password, version, hdtype, hdcode, serverid, ip,
		siteid, bunldID)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvLoginName", loginname, 32)
	ps.AddVarcharInput("chvPassword", password, 32)
	ps.AddIntInput("tnyHDType", hdtype)
	ps.AddVarcharInput("chvHDCode", hdcode, 64)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddIntInput("intAreaID", siteid)
	ps.AddVarcharInput("chvCode", vcode, 16)
	ps.AddIntInput("intCountryCode", countryCode)

	ps.AddIntOutput("intUserID", userid)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddVarcharOutput("chvHttpAddr", "")

	proName := "PrPs_UserLogin_SMS"

	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("执行存储过程 PrPs_UserLogin_SMS 出错%v", err)
		return UserLoginResponseError(0, "服务器错误.", sid)

	} else {
		// ret = 10154 玩家還在遊戲中，所以要先通知GS讓他斷線
		retCode := ret.GetReturnValue()
		if retCode == 10154 {
			result := SendGSkickUser(ret)
			// result true 玩家在遊戲中踢掉
			if result {
				go func() {
					time.Sleep(time.Duration(1) * time.Second)
					// 轉換protocal
					loginData := &netproto.UserLogin{}
					loginData.LoginName = mm.LoginName
					loginData.Password = mm.Password
					loginData.BunldID = mm.BunldID
					loginData.HDType = mm.HDType
					loginData.HDCode = mm.HDCode
					loginData.SiteID = mm.SiteID
					loginData.Version = mm.Version
					loginData.Ver = mm.Ver
					loginData.CountryCode = mm.CountryCode
					loginData.PlatformID = mm.PlatformID

					m.SClassID = int32(netproto.HallMsgClassID_UserLoginID)

					rData, _ := m.MsgData.(*rpcmsg.RPCMessage)
					rData.MsgData, _ = proto.Marshal(loginData)
					m.MsgData = rData

					h.hook(ctx, clt, m)
				}()

				return nil
			}
		} else if retCode != 1 {
			return UserLoginResponse(ret, sid)
		}
	}
	userid = int32(ret.GetOutputParamValue("intUserID").(int64))
	return UserLoginCommon(siteid, userid, hdtype, hdcode, version, serverid, ip, bunldID, ver, "", sid)
}
