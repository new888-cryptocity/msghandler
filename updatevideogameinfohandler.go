package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

type UpdateVideoGameInfoHandler struct {
}

func (rh *UpdateVideoGameInfoHandler) Init() {}
func (rh *UpdateVideoGameInfoHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_UpdateVideoGameInfoID)
}

func (rh *UpdateVideoGameInfoHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.UpdateVideoGameInfo{}
		return mm
	})

	//組織回包數據
	ret := new(netproto.UpdateVideoGameInfoRet)
	ret.Code = proto.Int32(0)

	tmpmsg := new(msg.Message)
	tmpmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	tmpmsg.SClassID = int32(netproto.DBServerClassID_UpdateVideoGameInfoID)
	tmpmsg.MsgData = ret

	if rpc == nil {
		return []*msg.Message{tmpmsg}
	}

	mm := rmm.(*netproto.UpdateVideoGameInfo)
	ctx.Info("update score [video] %v", mm)

	//msgid := rpc.GetQueueID()
	userid := mm.GetUserID()
	vgameid := mm.GetVGameID()
	cashnum := mm.GetCashNum()
	inorout := mm.GetInOrOut()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	//ps.AddIntInput("intMsgID", msgid)
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intVGameID", vgameid)
	ps.AddIntInput("intCashNum", cashnum)
	ps.AddIntInput("intInOrOut", inorout)

	proc := db.NewProcedure("PrGs_UpdateShixunScore", ps)
	dbcode, err := d.ExecProc(proc)

	var retcode int32 = 1
	if err != nil {
		ctx.Error("執行存儲過程PrGs_UpdateShixunScore錯誤, %v", err)
		dbcode, err = d.ExecProc(proc) //重試一次
		if err != nil {
			retcode = 0
			if cashnum > 0 {
				ctx.Error("/*【budan3_start需補單】*/ USE CenterDB; %s \nGO\n/*【budan3_end需補單】*/", proc.GetSqlString())
			}
			ctx.Error("PrGs_UpdateShixunScore失敗,intInOrOut: %d, intUserID: %d, intVGameID: %d, intCashNum: %d", inorout, userid, vgameid, cashnum)
		} else {
			retcode = int32(dbcode.GetReturnValue())
			ctx.Error("重試執行存儲過程PrGs_UpdateShixunScore成功")
		}
	} else {
		retcode = int32(dbcode.GetReturnValue())
	}

	//返回數據
	ret.Code = proto.Int32(retcode)
	ret.Uvgi = mm
	tmpmsg.MsgData = ret

	return []*msg.Message{tmpmsg}
}
