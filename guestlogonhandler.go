package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	"github.com/golang/protobuf/proto"
)

//處理遊客登入請求 (unity) (若第一次登入, 會註冊到 Users.UserInfo 錶)
type GuestLogonHandler struct {
}

func (h *GuestLogonHandler) Init() {}
func (h *GuestLogonHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_GuestLoginID)
}

func (h *GuestLogonHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GuestLogin{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.GuestLogin)

	hdtype := mm.GetHDType()
	hdcode := mm.GetHDCode()
	ip := mm.GetClientIP()
	if ip == "" {
		ip = "unknow"
	}
	siteid := mm.GetSiteID()
	version := mm.GetVersion()
	platformId := mm.GetPlatformID()
	serverid := rpc.GetRouteServerID()

	wxOpenID := mm.GetWxOpenID()
	wxUnionID := mm.GetWxUnionID()

	bunldID := mm.GetBunldID()
	ver := mm.GetVer()
	language := mm.GetLanguage()
	if language == "" {
		language = "en-US" //語言	zh_CN/en_US/zh_TW
	}

	mlog.Debug("[DbServer] 處理遊客登入請求開始 Ver=%s, HDType=%d, HDCode=%s, ip=%s, ServerID=%d, SiteID(AreaID)=%d, BunldID=%s, Version=%s, PlatformID=%d, Language=%s",
		ver, hdtype, hdcode, ip, serverid,
		siteid, bunldID, version,
		platformId, language)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvVersion", version, 16)
	ps.AddIntInput("tnyHDType", hdtype)
	ps.AddVarcharInput("chvHDCode", hdcode, 64)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddIntInput("sintPlazaServerID", serverid)
	ps.AddIntInput("intAreaID", siteid)
	ps.AddVarcharInput("chvBunldID", bunldID, 64)
	ps.AddVarcharInput("chvVer", ver, 64)
	ps.AddIntOutput("intUserID", 0)
	ps.AddVarcharOutput("chvCertification", "")
	ps.AddVarcharOutput("chvErrMsg", "")

	proName := "PrPs_UserLogon_Fast"
	if len(wxOpenID) >= 10 && len(wxOpenID) <= 100 { //客戶端微信登入
		ps.AddVarcharInput("WxOpenID", wxOpenID, 100)
		ps.AddVarcharInput("WxUnionID", wxUnionID, 100)
		proName = "PrPs_UserLogon_Fast_Wx"
	} else if len(wxOpenID) > 0 {
		mlog.Error("收到非法微信openID登入：%s", wxOpenID)
	}

	mlog.Debug("[%s] chvVersion=%v", proName, version)
	mlog.Debug("[%s] tnyHDType=%v", proName, hdtype)
	mlog.Debug("[%s] chvHDCode=%v", proName, hdcode)
	mlog.Debug("[%s] chvIPAddress=%v", proName, ip)
	mlog.Debug("[%s] sintPlazaServerID=%v", proName, serverid)
	mlog.Debug("[%s] intAreaID=%v", proName, siteid)
	mlog.Debug("[%s] chvBunldID=%v", proName, bunldID)
	mlog.Debug("[%s] chvVer=%v", proName, ver)

	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_LoginRetID), nil)

	if err != nil {
		mlog.Error("執行存儲過程%s出錯%v", proName, err)
		logret := &netproto.UserLoginRet{}
		logret.Code = proto.Int32(0)
		logret.Message = proto.String(fmt.Sprintf("服務器錯誤."))
		retmsg.MsgData = logret

	} else {

		mlog.Debug("[%s] 執行成功 ret=%v", proName, ret)

		logret := parseUserLogonRet(ret) //PrPs_UserLogon_Fast
		logret.HDCode = proto.String(hdcode)
		logret.HDType = proto.Int32(hdtype)
		retmsg.MsgData = logret

		mlog.Debug("[DbServer] 處理遊客登入請求成功 Ver=%s, HDType=%d, HDCode=%s, ip=%s, ServerID=%d, SiteID(AreaID)=%d, BunldID=%s, Version=%s, PlatformID=%d, Language=%s, UserID=%d",
			ver, hdtype, hdcode, ip, serverid,
			siteid, bunldID, version,
			platformId, language, logret.UserID)

		if logret.GetCode() == 1 {
			//修改國際語言代碼
			ExecUpdLanguage(logret.GetUserID(), language)
		}
	}

	return []*msg.Message{retmsg}
}
