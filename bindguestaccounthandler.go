package msghandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"

	"crypto/rand"
	"strings"
)

//修改頭像
type BindGuestAccountHandler struct {
}

func (rh *BindGuestAccountHandler) Init() {}
func (rh *BindGuestAccountHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_BindGuestAccountID)
}

func (rh *BindGuestAccountHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BindGuestAccount{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BindGuestAccount)

	validType := mm.GetValidType()
	email := mm.GetEmail()
	token := mm.GetValidToken()

	validError := ValidTypeHandle(validType, email, token)
	if validError != nil {
		return BindGuestAccountResponse(0, "TEST 第三方驗証錯誤", "")
	}

	userid := rpcinfo.GetUserID()
	tel := mm.GetTel()
	vcode := mm.GetVCode()
	passwd := mm.GetPassword()
	userIDAgent := mm.GetAgentUserID()
	countryCode := mm.GetCountryCode()
	/* 取消預設值 2022.04.06
	if countryCode == 0 {
		countryCode = 86
	}
	*/

	strTemp := ""
	TempInviteCode := ""
	for iIndex := 0; iIndex < 8-len(strconv.FormatInt(int64(rpcinfo.GetUserID()), 10)); iIndex++ {
		strTemp += "0"
	}
	invitecode := strTemp + strconv.FormatInt(int64(rpcinfo.GetUserID()), 10)
	TempInviteCode = invitecode[0:4]
	TempInviteCode = strings.Replace(TempInviteCode, "0", "B", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "1", "R", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "2", "U", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "3", "C", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "4", "E", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "5", "W", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "6", "A", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "7", "Y", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "8", "N", -1)
	TempInviteCode = strings.Replace(TempInviteCode, "9", "M", -1)
	invitecode = TempInviteCode + invitecode[4:8]

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharInput("chvCode", vcode, 32)
	ps.AddVarcharInput("chvTel", tel, 32)
	ps.AddVarcharInput("chvPassword", passwd, 32)
	ps.AddVarcharOutput("chvErrMsg", "")
	ps.AddVarcharInput("chvAgentUserID", userIDAgent, 16)
	ps.AddIntInput("intCountryCode", countryCode)

	ps.AddVarcharInput("chvValidType", validType, 10)
	ps.AddVarcharInput("chvEmail", email, 255)
	ps.AddVarcharOutput("chvInviteCode", invitecode)
	proc := db.NewProcedure("PrPs_User_BindGuest", ps)
	ret, err := d.ExecProc(proc)

	var code int32 = 0
	retmessage := ""
	retInviteCode := ""
	if err != nil {
		ctx.Error("執行存儲過程PrPs_User_BindGuest錯誤, %v", err)
		code = 0
		retmessage = "DB Process Error!"
	} else {
		code = int32(ret.GetReturnValue())
		retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
		retInviteCode = ret.GetOutputParamValue("chvInviteCode").(string)
	}

	return BindGuestAccountResponse(code, retmessage, retInviteCode)
}

func BindGuestAccountResponse(code int32, retmessage string, retInviteCode string) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_BindGuestAccountRetID)
	retmsg.MsgData = getRetMsgAddInvite(code, retmessage, retInviteCode)
	return []*msg.Message{retmsg}
}

// 判斷要驗証的方式，如果都不是回傳 nil
func ValidTypeHandle(validType string, email string, token string) error {
	switch validType {
	case "facebook", "google":
		return SendBackendValidAccount(validType, email, token)
	default:
		return nil
	}
}

// 送Api給後臺確認帳號
func SendBackendValidAccount(validName string, email string, token string) error {
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

	r := &backendResponse{}
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
type backendResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func GetRandNumber(min int64, max int64) string {
	iNumber := big.NewInt(max - min)
	iBigIntRandNumber, _ := rand.Int(rand.Reader, iNumber)
	strRandIndex := iBigIntRandNumber.String()
	iRandNumber, _ := strconv.ParseInt(strRandIndex, 10, 32)
	iRandNumber += min
	return strconv.FormatInt(iRandNumber, 10)
}
