package msghandler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

func NewUserLogonHandler(hook func(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) bool) *UserLogonHandler {
	handle := &UserLogonHandler{}
	handle.hook = hook
	return handle
}

//用户登录
type UserLogonHandler struct {
	hook func(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) bool
}

func (h *UserLogonHandler) Init() {}
func (h *UserLogonHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_UserLoginID)
}

func (h *UserLogonHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UserLogin{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	sid := netproto.HallMsgClassID_LoginRetID

	mm := rmm.(*netproto.UserLogin)

	hdtype := mm.GetHDType()
	hdcode := mm.GetHDCode()
	ip := mm.GetClientIP()
	if ip == "" {
		ip = "unknow"
	}
	siteid := mm.GetSiteID()
	version := mm.GetVersion()
	serverid := rpc.GetRouteServerID()

	userid := mm.GetUserID()
	cer := mm.GetCer()
	loginname := mm.GetLoginName()
	password := mm.GetPassword()

	bunldID := mm.GetBunldID()
	ver := mm.GetVer()
	countryCode := mm.GetCountryCode()
	if countryCode == 0 {
		countryCode = 86
	}

	validType := mm.GetValidType()
	email := mm.GetEmail()
	token := mm.GetValidToken()

	validError := ValidTypeHandle(validType, email, token)
	if validError != nil {
		return UserLoginResponseError(0, "TEST 第三方驗証錯誤", sid)
	}

	mlog.Debug(`[DbServer] 處理用户登录請求開始 loginname=%s, password=%s, version=%s, hdtype=%d, hdcode=%s, serverid=%d, ip=%s
		siteid=%d, bunldID=%s, ver=%s`,
		loginname, password, version, hdtype, hdcode, serverid, ip,
		siteid, bunldID, ver)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("tnyHDType", hdtype)
	ps.AddVarcharInput("chvHDCode", hdcode, 64)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddIntInput("intAreaID", siteid)
	ps.AddVarcharInput("chvValidType", validType, 10)
	ps.AddVarcharInput("chvEmail", email, 255)
	ps.AddIntInput("intCountryCode", countryCode)
	if loginname != "" && password != "" {
		ps.AddVarcharInput("chvLoginName", loginname, 32)
		ps.AddVarcharInput("chvPassword", password, 32)
	}

	ps.AddIntOutput("intUserID", userid)
	ps.AddVarcharOutput("chvCertification", cer)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddVarcharOutput("chvHttpAddr", "")

	proName := "PrPs_UserLogon_Normal"

	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("执行存储过程 %s 出错%v", proName, err)
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
					h.hook(ctx, clt, m)
				}()

				return nil
			}
		} else if retCode != 1 {
			return UserLoginResponse(ret, sid)
		}
	}
	userid = int32(ret.GetOutputParamValue("intUserID").(int64))
	return UserLoginCommon(siteid, userid, hdtype, hdcode, version, serverid, ip, bunldID, ver, cer, sid)
}

func UserLoginResponse(ret db.RecordSet, sid netproto.HallMsgClassID) []*msg.Message {
	logret := parseUserLogonRet(ret)
	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(sid), nil)
	retmsg.MsgData = logret
	return []*msg.Message{retmsg}
}

func UserLoginResponseError(code int32, retmessage string, sid netproto.HallMsgClassID) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(sid)

	logret := &netproto.UserLoginRet{}
	logret.Code = proto.Int32(code)
	logret.Message = proto.String(retmessage)
	retmsg.MsgData = logret

	return []*msg.Message{retmsg}
}

func UserLoginCommon(areaId int32, userId int32, hdType int32, hdcode string, version string, serverId int32, ip string, bunldId string, ver string, cer string,
	sid netproto.HallMsgClassID) []*msg.Message {
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intAreaID", areaId)
	//ps.AddIntInput("intUserID", userId)
	ps.AddIntInput("tnyLogonType", 0)
	ps.AddIntInput("tnyHDType", hdType)
	ps.AddVarcharInput("chvHDCode", hdcode, 64)
	ps.AddVarcharInput("chvVersion", version, 16)
	ps.AddIntInput("sintPlazaServerID", serverId)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddVarcharInput("chvBunldID", bunldId, 64)
	ps.AddVarcharInput("chvVer", ver, 64)

	ps.AddVarcharOutput("chvCertification", cer)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddIntOutput("intUserID", userId)

	proName := "PrPs_UserLogon_Common"
	mlog.Debug("[%s] intAreaID=%v", proName, areaId)
	mlog.Debug("[%s] intUserID=%v", proName, userId)
	mlog.Debug("[%s] tnyHDType=%v", proName, hdType)
	mlog.Debug("[%s] chvHDCode=%v", proName, hdcode)
	mlog.Debug("[%s] chvVersion=%v", proName, version)
	mlog.Debug("[%s] sintPlazaServerID=%v", proName, serverId)
	mlog.Debug("[%s] chvIPAddress=%v", proName, ip)
	mlog.Debug("[%s] chvBunldID=%v", proName, bunldId)
	mlog.Debug("[%s] chvVer=%v", proName, ver)
	mlog.Debug("[%s] chvCertification=%v", proName, cer)

	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(sid), nil)

	if err != nil {
		mlog.Error("执行存储过程 %v 出错%v", proName, err)
		return UserLoginResponseError(0, "服务器错误.", sid)

	} else {
		logret := parseUserLogonRet(ret) //PrPs_UserLogon_Normal
		if logret.UserID == nil {
			logret.UserID = &userId
		}
		retmsg.MsgData = logret

		mlog.Debug(`[DbServer] 處理帳號通用登入請求成功 version=%s, hdtype=%d, hdcode=%s, serverid=%d, ip=%s, siteid=%d, bunldID=%s, ver=%s, UserID=%d`,
			version, hdType, hdcode, serverId, ip, areaId, bunldId, ver, logret.UserID)

		return []*msg.Message{retmsg}
	}
}

func SendGSkickUser(ret db.RecordSet) bool {
	iterfaceAddr := ret.GetOutputParamValue("chvHttpAddr")
	httpAddr := ""
	if iterfaceAddr == nil {
		return false
	} else {
		httpAddr = iterfaceAddr.(string)
	}
	userid := int32(ret.GetOutputParamValue("intUserID").(int64))
	client := &http.Client{}
	vcode := fmt.Sprintf("actionkilloutuseruid%v%v", userid, "abc123")
	h := md5.New()
	h.Write([]byte(vcode))
	vcode = hex.EncodeToString(h.Sum(nil))

	req, err := http.NewRequest("POST", "http://"+httpAddr+"/game", strings.NewReader(fmt.Sprintf("action=killoutuser&uid=%v&vcode=%v", userid, vcode)))
	if err != nil {
		mlog.Info("Login NewRequest Error %s", err)
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		mlog.Info("Login Request Do Error %s", err)
		return false
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mlog.Info("Login Request body Error %s", err)
		return false
	}
	r := &netproto.Http_Result{}
	if err := json.Unmarshal(content, r); err == nil {
		return *r.Result
	} else {
		return false
	}
}
