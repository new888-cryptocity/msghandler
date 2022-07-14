package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

//大廳啓動
type HallServerStartHandler struct {
}

func (rh *HallServerStartHandler) Init() {}
func (rh *HallServerStartHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_HallServerStartID)
}

func (rh *HallServerStartHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.HallServerStart{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.HallServerStart)

	serverid := mm.GetServerID()
	groupid := mm.GetLianyunID()
	clientaddr := mm.GetClientAddr()
	notifyaddr := mm.GetNotifyAddr()
	httpaddr := mm.GetHttpAddr()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("sintServerID", serverid)
	ps.AddIntInput("intGroupID", groupid)
	ps.AddVarcharInput("chvClientAddr", clientaddr, 32)
	ps.AddVarcharInput("chvNotifyAddr", notifyaddr, 32)
	ps.AddVarcharInput("chvHttpAddr", httpaddr, 32)

	proc := db.NewProcedure("PrPs_ServerStart", ps)
	dbret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_DBServer), int32(netproto.DBServerClassID_HallServerStartRetID), nil)
	var code int32
	code = 1
	message := "服務器啓動成功"
	if err != nil {
		message = fmt.Sprintf("執行存儲過程PrPs_ServerStart錯誤, %v", err)
		ctx.Error(message)
		code = 0
	}

	retdata := new(netproto.HallServerStartRet)
	retdata.Code = proto.Int32(code)
	retdata.Message = proto.String(message)

	if dbret.GetRetTableCount() >= 1 {
		retdata.AndroidVersion = proto.String(dbret[0].GetSingleValue("AndroidVersion").(string))
		retdata.IOSVersion = proto.String(dbret[0].GetSingleValue("IOSVersion").(string))
	} else {
		retdata.Code = proto.Int32(0)
		retdata.Message = proto.String("未獲取到平臺版本信息")
	}
	retmsg.MsgData = retdata

	return []*msg.Message{retmsg}
}
