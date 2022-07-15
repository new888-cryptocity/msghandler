package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//遊戲服務器啓動
type GameServerStartHandler struct {
}

func (rh *GameServerStartHandler) Init() {}
func (rh *GameServerStartHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_GameServerStartID)
}

func (rh *GameServerStartHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GameServerStart{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.GameServerStart)

	serverid := mm.GetServerID()
	groupid := mm.GetLianyunID()
	clientaddr := mm.GetClientAddr()
	notifyaddr := mm.GetNotifyAddr()
	httpaddr := mm.GetHttpAddr()
	gameid := mm.GetGameID()
	servername := mm.GetServerName()
	showname := mm.GetShowName()
	loginmoney := mm.GetLoginMoney()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("sintServerID", serverid)
	ps.AddIntInput("sintGameID", gameid)
	ps.AddIntInput("intGroupID", groupid)
	ps.AddVarcharInput("chvServerName", servername, 64)
	ps.AddVarcharInput("chvShowName", showname, 64)
	ps.AddVarcharInput("chvClientAddr", clientaddr, 32)
	ps.AddVarcharInput("chvNotifyAddr", notifyaddr, 32)
	ps.AddVarcharInput("chvHttpAddr", httpaddr, 32)
	ps.AddIntInput("lngLoginMoney", loginmoney)
	ps.AddVarcharOutput("chvErrMsg", "")

	proName := "PrGs_ServerStart"
	mlog.Debug("[%s] gameid=%v", proName, gameid)
	mlog.Debug("[%s] servername=%v", proName, servername)
	mlog.Debug("[%s] showname=%v", proName, showname)
	mlog.Debug("[%s] clientaddr=%v", proName, clientaddr)
	mlog.Debug("[%s] notifyaddr=%v", proName, notifyaddr)
	mlog.Debug("[%s] httpaddr=%v", proName, httpaddr)
	mlog.Debug("[%s] loginmoney=%v", proName, loginmoney)
	mlog.Debug("[%s] serverid=%v", proName, serverid)
	mlog.Debug("[%s] groupid=%v", proName, groupid)

	proc := db.NewProcedure(proName, ps)
	dbret, err := d.ExecProc(proc)

	var code int32
	code = 1
	message := "服務器啓動成功"

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_DBServer), int32(netproto.DBServerClassID_GameServerStartRetID), nil)
	retdata := new(netproto.GameServerStartRet)

	if err != nil {
		message = fmt.Sprintf("執行存儲過程 %s 錯誤, %v", proName, err)
		ctx.Error(message)
		code = 0
	} else {
		retdata.AndroidVersion = proto.String(dbret[0].GetSingleValue("AndroidVersion").(string))
		retdata.IOSVersion = proto.String(dbret[0].GetSingleValue("IOSVersion").(string))
		retdata.Status = proto.Int32(dbret[0].GetSingleValueInt32("Status"))
		retdata.StatusMessage = proto.String(dbret[0].GetSingleValue("StatusMessage").(string))
		code = int32(dbret.GetReturnValue())
		message = dbret.GetOutputParamValue("chvErrMsg").(string)
	}

	retdata.Code = proto.Int32(code)
	retdata.Message = proto.String(message)
	retmsg.MsgData = retdata

	return []*msg.Message{retmsg}
}
