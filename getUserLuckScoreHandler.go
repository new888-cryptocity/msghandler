package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

type GetUserLuckScoreHandler struct {
}

func (rh *GetUserLuckScoreHandler) Init() {}
func (rh *GetUserLuckScoreHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Game) && sclsid == int32(netproto.GameMessageClassID_UserLuckScoreID)
}

func (rh *GetUserLuckScoreHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UserLuckScore{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.UserLuckScore)

	serverid := rpc.GetRouteServerID()
	userid := mm.GetUserID()
	score := mm.GetScore()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intServerID", serverid)
	ps.AddIntInput("intUserID", userid)
	ps.AddBigIntInput("lngScore", score)
	ps.AddVarcharOutput("chvValue", "")

	proc := db.NewProcedure("PrGs_GetUserLuckScore", ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		ctx.Error("執行存儲過程PrGs_ServerReg錯誤, %v", err)
		return nil
	} else {
		luckValue := ret.GetOutputParamValue("chvValue").(string)
		retData := new(netproto.UserLuckScoreRet)
		retData.Value = proto.String(luckValue)
		retData.UserID = proto.Int32(userid)
		retmsg := new(msg.Message)
		retmsg.BClassID = int32(netproto.MessageBClassID_Game)
		retmsg.SClassID = int32(netproto.GameMessageClassID_UserLuckScoreRetID)
		retmsg.MsgData = retData
		return []*msg.Message{retmsg}
	}

}
