package msghandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//修改頭像
type DZPKHALL_BindAccountHandler struct {
}

func (rh *DZPKHALL_BindAccountHandler) Init() {}
func (rh *DZPKHALL_BindAccountHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_DZPKHALL_BindAccount)
}

func (rh *DZPKHALL_BindAccountHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DZPKHALLBindAccountLogin{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.DZPKHALLBindAccountLogin)

	validType := mm.GetValidType()
	email := mm.GetEmail()
	token := mm.GetValidToken()

	validError := DZPKHALL_ValidTypeHandle(validType, email, token)
	if validError != nil {
		return DZPKHALL_BindGuestAccountResponse(0, "TEST 第三方驗証錯誤")
	}

	hdtype := mm.GetHDType()
	hdcode := mm.GetHDCode()
	ip := rpc.GetIPAddress()
	siteid := mm.GetSiteID()
	version := mm.GetVersion()
	platformId := mm.GetPlatformID()
	serverid := rpc.GetRouteServerID()

	bunldID := mm.GetBunldID()
	ver := mm.GetVer()
	language := mm.GetLanguage()
	if language == "" {
		language = "zh_CN" //語言	zh_CN/en_US/zh_TW
	}

	mlog.Debug("[DbServer] 處理第三方登入請求開始 Ver=%s, HDType=%d, HDCode=%s, ip=%s, ServerID=%d, SiteID(AreaID)=%d, BunldID=%s, Version=%s, PlatformID=%d, Language=%s",
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
	//ps.AddVarcharInput("chvBunldID", bunldID, 64)
	//ps.AddVarcharInput("chvVer", ver, 64)
	ps.AddIntOutput("intUserID", 0)
	ps.AddVarcharOutput("chvCertification", "")
	ps.AddVarcharOutput("chvErrMsg", "")

	ps.AddVarcharInput("chvValidType", validType, 10)
	ps.AddVarcharInput("chvEmail", email, 255)

	proName := "DZPKHALL_PrPs_ThirdBindLogon"
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

		mlog.Debug("[DbServer] 處理第三方登入請求成功 Ver=%s, HDType=%d, HDCode=%s, ip=%s, ServerID=%d, SiteID(AreaID)=%d, BunldID=%s, Version=%s, PlatformID=%d, Language=%s, UserID=%d",
			ver, hdtype, hdcode, ip, serverid,
			siteid, bunldID, version,
			platformId, language, logret.UserID)

		if logret.GetCode() == 1 {
			//修改國際語言代碼
			ExecUpdLanguage(logret.GetUserID(), language)
		}
	}

	return []*msg.Message{retmsg}

	/*
		var code int32 = 0
		retmessage := ""
		if err != nil {
			ctx.Error("執行存儲過程DZPKHALL_PrPs_ThirdBindLogon錯誤, %v", err)
			code = 0
			retmessage = "服務器錯誤"
		} else {
			if ret.GetReturnValue() == 1 {
				code = 1
			}
			retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
		}

		return DZPKHALL_BindGuestAccountResponse(code, retmessage)
	*/
}

func DZPKHALL_BindGuestAccountResponse(code int32, retmessage string) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_BindGuestAccountRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}

// 判斷要驗証的方式，如果都不是回傳 nil
func DZPKHALL_ValidTypeHandle(validType string, email string, token string) error {
	switch validType {
	case "facebook", "google":
		return DZPKHALL_SendBackendValidAccount(validType, email, token)
	default:
		return nil
	}
}

// 送Api給後臺確認帳號
func DZPKHALL_SendBackendValidAccount(validName string, email string, token string) error {
	resp, err := http.Get(fmt.Sprintf("http://192.168.3.100/verifyemailapi/%s/%s/%s", email, validName, token))
	if err != nil {
		mlog.Info("BindGuest backend Error %s", err)
		return err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mlog.Info("BindGuest backend body Error %s", err)
		return err
	}

	// 後臺回應的內容
	//mlog.Info(string(content))

	r := &dzpkhall_backendResponse{}
	err = json.Unmarshal(content, &r)
	if err != nil {
		return err
	} else if r.Code == 101 {
		return nil
	} else {
		return fmt.Errorf("backend response code:%d", r.Code)
	}
}

// 後臺回傳的結構
type dzpkhall_backendResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}
