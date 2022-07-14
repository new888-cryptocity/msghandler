package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

//處理H5遊戲登出請求 (html5 game)
type H5GameLogoutHandler struct {
}

func (h *H5GameLogoutHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_H5GameLogoutID)
}

//登出請求 TODO:後面改登出邏輯 和 H5GameLogoutStart
func (h *H5GameLogoutHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.H5GameLoginStart{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	ip := rpc.GetIPAddress()

	mm := rmm.(*netproto.H5GameLoginStart)
	hdtype := mm.GetHDType()
	hdcode := mm.GetHDCode()
	siteid := mm.GetSiteID()
	version := mm.GetVersion()
	platformId := mm.GetPlatformID()
	serverid := rpc.GetRouteServerID()
	bunldID := mm.GetBunldID()

	ver := mm.GetVer()
	language := mm.GetLanguage()
	if language == "" {
		language = "zh_CN" //语言	zh_CN/en_US/zh_TW
	}

	token := mm.GetToken()
	thirdPartyUserID := mm.GetThirdPartyUserID()
	thirdPartyUserIDStr := mm.GetThirdPartyUserIDStr()
	thirdPartyCurrency := mm.GetThirdPartyCurrency()
	thirdPartyUserAmount := mm.GetThirdPartyUserAmount()
	browserType := mm.GetBrowserType()
	browserVer := mm.GetBrowserVer()

	mlog.Debug(`[DbServer] 處理 H5遊戲登入請求 開始
		Ver=%s, HDType=%d, HDCode=%s, ip=%s, ServerID=%d,
		SiteID(AreaID)=%d, BunldID=%s, Version=%s, PlatformID=%d, Language=%s
		token=%s, thirdPartyUserID=%d, thirdPartyUserIDStr=%s, thirdPartyCurrency=%s, thirdPartyUserAmount=%d,
		browserType=%d, browserVer=%s`,
		ver, hdtype, hdcode, ip, serverid,
		siteid, bunldID, version, platformId, language,
		token, thirdPartyUserID, thirdPartyUserIDStr, thirdPartyCurrency, thirdPartyUserAmount,
		browserType, browserVer)

	if thirdPartyUserID <= 0 && len(thirdPartyUserIDStr) == 0 {
		// 帳號資訊不能為空
		mlog.Error("收到非法 H5Game 登录1： thirdPartyUserID=%d, thirdPartyUserIDStr=%s", thirdPartyUserID, thirdPartyUserIDStr)

		return []*msg.Message{GetDBRetMsg(0)}
	}
	if thirdPartyUserID > 0 && len(thirdPartyUserIDStr) > 0 {
		// 帳號資訊只能二擇一
		mlog.Error("收到非法 H5Game 登录2： thirdPartyUserID=%d, thirdPartyUserIDStr=%s", thirdPartyUserID, thirdPartyUserIDStr)
		return []*msg.Message{GetDBRetMsg(0)}
	}
	if thirdPartyUserAmount <= 0 {
		// 錢幣=0 不給註冊和登入
		mlog.Error("收到非法 H5Game 登录3： thirdPartyUserID=%d, thirdPartyUserIDStr=%s, thirdPartyUserAmount=%d", thirdPartyUserID, thirdPartyUserIDStr, thirdPartyUserAmount)
		// 後面要幹嘛? 前人也沒製作 XD
		return []*msg.Message{GetDBRetMsg(0)}
	}

	proName := "PrPs_UserLogon_Fast_Token"
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvToken", token, 50)                             //上層網頁平台產生的Token
	ps.AddBigIntInput("intThirdPartyUserID", thirdPartyUserID)            //上層網頁平台的會員ID
	ps.AddVarcharInput("chvThirdPartyUserIDStr", thirdPartyUserIDStr, 50) //上層網頁平台的會員ID
	ps.AddVarcharInput("chvThirdPartyCurrency", thirdPartyCurrency, 10)   //上層網頁平台的會員幣值
	ps.AddBigIntInput("intThirdPartyUserAmount", thirdPartyUserAmount)    //上層網頁平台的會員錢幣數量
	ps.AddIntInput("chvBrowserType", browserType)                         //瀏覽器種類 0:unkonw 1:chrome 2:其他
	ps.AddVarcharInput("chvBrowserVer", browserVer, 16)                   //瀏覽器版本号
	ps.AddVarcharInput("chvVersion", version, 16)                         //客户端版本号
	ps.AddIntInput("tnyHDType", hdtype)                                   //硬件类型
	ps.AddVarcharInput("chvHDCode", hdcode, 64)                           //硬件码
	ps.AddVarcharInput("chvIPAddress", ip, 15)                            //IP地址
	ps.AddIntInput("sintPlazaServerID", serverid)                         //广场ID
	ps.AddIntInput("intAreaID", siteid)                                   //渠道ID
	ps.AddVarcharInput("chvBunldID", bunldID, 64)                         //BunldID
	ps.AddVarcharInput("chvVer", ver, 64)                                 //版本号
	ps.AddIntOutput("intUserID", 0)                                       //用户ID
	ps.AddVarcharOutput("chvCertification", "")                           //证书
	ps.AddVarcharOutput("chvErrMsg", "")

	mlog.Debug("[%s] token=%v", proName, token)
	mlog.Debug("[%s] thirdPartyUserID=%v", proName, thirdPartyUserID)
	mlog.Debug("[%s] thirdPartyUserIDStr=%v", proName, thirdPartyUserIDStr)
	mlog.Debug("[%s] thirdPartyCurrency=%v", proName, thirdPartyCurrency)
	mlog.Debug("[%s] browserType=%v", proName, browserType)
	mlog.Debug("[%s] browserVer=%v", proName, browserVer)
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

	//組合回傳封包 (回傳給hallServer)
	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_LoginRetID), nil)

	if err != nil {
		mlog.Error("执行存储过程%s出错%v", proName, err)
		logret := &netproto.UserLoginRet{}
		logret.Code = proto.Int32(0)
		logret.Message = proto.String(fmt.Sprintf("服务器错误."))
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
			//修改国际语言代码
			ExecUpdLanguage(logret.GetUserID(), language)
		}
	}

	return []*msg.Message{retmsg}
}
