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
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//取得錢包位址
type DZPKHALL_WalletWithdrawHandler struct {
}

func (rh *DZPKHALL_WalletWithdrawHandler) Init() {}
func (rh *DZPKHALL_WalletWithdrawHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_DZPKHALL_WalletWithdraw)
}

func (rh *DZPKHALL_WalletWithdrawHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DZPKHALLWalletWithdrawReq{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.DZPKHALLWalletWithdrawReq)
	userid := mm.GetUserID()
	towalletaddress := mm.GetWalletAddress()
	amount := mm.GetAmount()
	OrderID := ""
	var WithDrawAmount int64 = 0
	var PendingAmount int64 = 0
	var CurrAmount int64 = 0

	mlog.Debug("[DbServer] HallMsgClassID_DZPKHALLWalletWithdraw")

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intUserID", userid)
	ps.AddVarcharInput("chvToWalletaddress", towalletaddress, 128)
	ps.AddBigIntInput("lngAmount", amount)
	ps.AddIntOutput("intCode", -1)
	ps.AddVarcharOutput("chvOrderID", OrderID)
	ps.AddBigIntOutput("lngWithDrawAmount", WithDrawAmount)
	ps.AddBigIntOutput("lngPendingAmount", PendingAmount)
	ps.AddBigIntOutput("lngCurrAmount", CurrAmount)

	debugtest := ctx.Config.GetConfig("db").GetConfigValue("conn", "debugtest")

	proName := "DZPKHALL_PrPs_Money_WalletWithdraw"
	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		ctx.Error("執行存儲過程DZPKHALL_PrPs_Money_WalletWithdraw錯誤1, %v", err)
		return nil
	}

	code := ret[0].GetSingleValueInt32("intCode")

	// 產生提領訂單成功.
	if code == 1 {
		// 打api給金庫.
		OrderID = ret[0].GetSingleValueString("chvOrderID")
		DZPKHALL_SendWithdraw(debugtest, userid, OrderID, towalletaddress, amount)

		// 回傳client.
		retData := new(netproto.DZPKHALLWalletWithdrawRet)
		retData.Code = proto.Int32(code)
		retData.Amount = proto.Int64(ret[0].GetSingleValueInt64("lngCurrAmount")) //proto.Int64(int64(ret.GetOutputParamValue("lngAmount").(int64)))
		retData.WithdrawAmount = proto.Int64(ret[0].GetSingleValueInt64("lngWithdrawAmount"))
		retData.PendingAmount = proto.Int64(ret[0].GetSingleValueInt64("lngPendingAmount"))

		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_WalletWithdrawRet)
		retmsg.MsgData = retData
		return []*msg.Message{retmsg}
	} else if code == -4 { // 未達最小提領金額 無法提領
		// 回傳client.
		retData := new(netproto.DZPKHALLWalletWithdrawRet)
		retData.Code = proto.Int32(code)
		retData.Amount = proto.Int64(ret[0].GetSingleValueInt64("lngCurrAmount")) //proto.Int64(int64(ret.GetOutputParamValue("lngAmount").(int64)))

		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_WalletWithdrawRet)
		retmsg.MsgData = retData
		return []*msg.Message{retmsg}
	} else if code == -5 { // 已達單日最高提領金額 今日無法再提領
		// 回傳client.
		retData := new(netproto.DZPKHALLWalletWithdrawRet)
		retData.Code = proto.Int32(code)
		retData.Amount = proto.Int64(ret[0].GetSingleValueInt64("lngCurrAmount")) //proto.Int64(int64(ret.GetOutputParamValue("lngAmount").(int64)))

		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_WalletWithdrawRet)
		retmsg.MsgData = retData
		return []*msg.Message{retmsg}
	}

	return []*msg.Message{GetDBRetMsg(code)}

	/*
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
		retmsg.SClassID = int32(netproto.HallMsgClassID_DZPKHALL_WalletWithdraw)
		retmsg.MsgData = GetDBRetMsg(0)

		return []*msg.Message{retmsg}*/
}

// 送Api給金庫提領金額.
func DZPKHALL_SendWithdraw(debugtest string, UserID int32, OrderID string, ToAddress string, WithdrawAmount int64) string {

	if debugtest == "1" {
		mlog.Info("DZPKHALL_SendWithdraw test0 deubg:%s", debugtest)
		return ""
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", "/wallets/123456/sender/transactions",
		strings.NewReader("{\"requests\":[{ \"order_id\": \""+fmt.Sprintf("%s%d", OrderID, UserID)+"\",  \"address\": \""+fmt.Sprintf("%s", ToAddress)+"\",    \"amount\": \""+fmt.Sprintf("%f", (float64(WithdrawAmount)/100))+"\",    \"user_id\": \""+fmt.Sprintf("%d", UserID)+"\"     }]}"))

	mlog.Info("DZPKHALL_SendWithdraw test %s deubg:%s", req, debugtest)

	if err != nil {
		mlog.Info("DZPKHALL_SendWithdraw Error %s", err)
		return ""
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-API-CODE", "J4D3as54dfjk")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mlog.Info("DZPKHALL_SendWithdraw Error %s", err)
		return ""
	}

	// 後臺回應的內容
	mlog.Info(string(content))

	var data map[string]interface{}
	json.Unmarshal([]byte(content), &data)

	// 失敗要退回penging中的錢
	if data["results"] == nil {
		mlog.Info("DZPKHALL_SendBackendCreateWalletAddress results is null ")

		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", UserID)
		ps.AddVarcharInput("chvOrderID", OrderID, 128)

		proName := "DZPKHALL_PrPs_Money_WalletWithdrawFail"
		proc := db.NewProcedure(proName, ps)
		_, err := d.ExecProc(proc)

		if err != nil {
			mlog.Info("執行存儲過程DZPKHALL_PrPs_Money_WalletWithdrawFail錯誤1, %v", err)
			return ""
		}
		return ""
	}

	return ""
}
