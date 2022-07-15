package msghandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	proto "github.com/golang/protobuf/proto"
)

//取得錢包位址
type DZPKHALL_WalletAddressHandler struct {
}

func (rh *DZPKHALL_WalletAddressHandler) Init() {}
func (rh *DZPKHALL_WalletAddressHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_DZPKHALL_WalletAddress)
}

func (rh *DZPKHALL_WalletAddressHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DZPKHALLWalletAddressGet{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.DZPKHALLWalletAddressGet)

	userid := mm.GetUserID()
	code := mm.GetCode()
	walletaddress := mm.GetWalletAddress()
	amount := mm.GetAmount()
	WithdrawAmount := mm.GetWithdrawAmount()
	PendingAmount := mm.GetPendingAmount()
	WithdrawFee := mm.GetWithdrawFee()
	MinWithdrawAmount := int64(0)
	DailyMaxWithdrawAmount := int64(0)

	mlog.Debug("[DbServer] HallMsgClassID_DZPKHALL_WalletAddress")

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntOutput("intCode", code)
	ps.AddVarcharOutput("chvWalletaddress", walletaddress)
	ps.AddBigIntOutput("lngAmount", amount)
	ps.AddBigIntOutput("lngWithdrawAmount", WithdrawAmount)
	ps.AddBigIntOutput("lngPendingAmount", PendingAmount)
	ps.AddIntOutput("intWithdrawFee", WithdrawFee)
	ps.AddBigIntOutput("intMinWithdrawAmount", MinWithdrawAmount)
	ps.AddBigIntOutput("intDailyMaxWithdrawAmount", DailyMaxWithdrawAmount)

	debugtest := ctx.Config.GetConfig("db").GetConfigValue("conn", "debugtest")

	proName := "DZPKHALL_PrPs_Money_Wallet"
	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		ctx.Error("執行存儲過程DZPKHALL_PrPs_Money_Wallet錯誤1, %v", err)
		return nil
	} else {

		code = ret[0].GetSingleValueInt32("intCode")

		// 產生玩家錢包位址.
		if code != 1 {
			walletaddress = DZPKHALL_SendBackendCreateWalletAddress(debugtest, userid)
			if len(walletaddress) > 0 {
				mlog.Debug("[DbServer] HallMsgClassID_DZPKHALL_WalletAddress1 code:%d %s", code, walletaddress)
				code = 2
				ps = db.NewSqlParameters()
				ps.AddIntInput("intUserID", userid)
				ps.AddIntOutput("intCode", code)
				ps.AddVarcharOutput("chvWalletaddress", walletaddress)
				ps.AddBigIntOutput("lngAmount", amount)
				ps.AddBigIntOutput("lngWithdrawAmount", WithdrawAmount)
				ps.AddBigIntOutput("lngPendingAmount", PendingAmount)
				ps.AddIntOutput("intWithdrawFee", WithdrawFee)
				ps.AddBigIntOutput("intMinWithdrawAmount", MinWithdrawAmount)
				ps.AddBigIntOutput("intDailyMaxWithdrawAmount", DailyMaxWithdrawAmount)

				proc = db.NewProcedure(proName, ps)
				ret, err = d.ExecProc(proc)

				walletaddress = ret[0].GetSingleValueString("chvWalletaddress")
				if err != nil {
					ctx.Error("執行存儲過程DZPKHALL_PrPs_Money_Wallet錯誤2, %v", err)
					return nil
				} else if int32(len(walletaddress)) == 0 {
					ctx.Error("執行存儲過程DZPKHALL_PrPs_Money_Wallet錯誤3, %v", err)
					return nil
				}
				mlog.Debug("[DbServer] HallMsgClassID_DZPKHALL_WalletAddress2 code:%d %s", code, walletaddress)
				//return rh.OnMessage(ctx, clt, m)
			} else {
				ctx.Error("產生玩家錢包位址錯誤")
			}
		}

		retData := new(netproto.DZPKHALLWalletAddressGet)
		retData.UserID = proto.Int32(userid)
		retData.Code = proto.Int32(ret[0].GetSingleValueInt32("intCode"))
		retData.WalletAddress = proto.String(ret[0].GetSingleValueString("chvWalletaddress"))
		//retData.Amount = proto.Int64(int64(ret.GetOutputParamValue("lngAmount").(int64)))
		retData.Amount = proto.Int64(ret[0].GetSingleValueInt64("lngAmount"))
		retData.WithdrawAmount = proto.Int64(ret[0].GetSingleValueInt64("lngWithdrawAmount"))
		retData.PendingAmount = proto.Int64(ret[0].GetSingleValueInt64("lngPendingAmount"))
		retData.WithdrawFee = proto.Int32(ret[0].GetSingleValueInt32("intWithdrawFee"))
		retData.MinWithdrawAmount = proto.Int64(ret[0].GetSingleValueInt64("intMinWithdrawAmount"))
		retData.DailyMaxWithdrawAmount = proto.Int64(ret[0].GetSingleValueInt64("intDailyMaxWithdrawAmount"))

		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_WalletAddress)
		retmsg.MsgData = retData

		mlog.Debug("[DbServer] HallMsgClassID_DZPKHALL_WalletAddress finish code:%d %s", ret[0].GetSingleValueInt32("intCode"), ret[0].GetSingleValueString("chvWalletaddress"))
		return []*msg.Message{retmsg}
	}

}

func DZPKHALL_WalletAddressResponse(code int32, retmessage string) []*msg.Message {
	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_WalletAddress)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}

// 送Api給後臺建立錢包位址
func DZPKHALL_SendBackendCreateWalletAddress(debugtest string, UserID int32) string {

	if debugtest == "1" {
		mlog.Info("DZPKHALL_SendBackendCreateWalletAddress test0 deubg:%s", debugtest)
		return ""
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "/wallets/123456/addresses", strings.NewReader("{\"count\": 1, \"labels\":[ \""+fmt.Sprintf("%d", UserID)+"\" ]}"))
	if err != nil {
		mlog.Info("BindGuest backend Error %s", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-API-CODE", "gf2dkl054Skj")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mlog.Info("BindGuest backend body Error %s", err)
		return ""
	}

	// 後臺回應的內容
	mlog.Info(string(content))

	var in dzpkhall_backendCreateWalletAddressResponse
	err = json.Unmarshal([]byte(content), &in)
	if err != nil {
		mlog.Info("DZPKHALL_SendBackendCreateWalletAddress body Error %s", err)
		return ""
	}

	mlog.Info(in.Address[0])
	return in.Address[0]

}

// 後臺回傳的結構
type dzpkhall_backendCreateWalletAddressResponse struct {
	Address StringSlice `json:"addresses"`
}

type StringSlice []string
