package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"
	proto "github.com/golang/protobuf/proto"
)

//修改頭像
type DZPKHALL_ChangeFaceHandler struct {
}

func (rh *DZPKHALL_ChangeFaceHandler) Init() {}
func (rh *DZPKHALL_ChangeFaceHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_DZPKHALL_ChangeFace)
}

func (rh *DZPKHALL_ChangeFaceHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.DZPKHALLChangeFace{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.DZPKHALLChangeFace)

	faceUrl := mm.GetFaceUrl()
	userID := rpc.GetUserID()

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvFaceUrl", faceUrl, 255)
	ps.AddIntInput("intUserID", userID)
	ps.AddBigIntOutput("lngCurrentAmount", 0)
	ps.AddVarcharOutput("chvDBFaceUrl", "")
	ps.AddIntOutput("intNextCost", 0)

	//ps.AddVarcharOutput("chvErrMsg", "")

	proName := "DZPKHALL_PrPs_ChangeFace"
	proc := db.NewProcedure(proName, ps)
	ret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_Hall), int32(netproto.HallMsgClassID_DZPKHALL_ChangeFaceRet), nil)
	if err != nil {
		mlog.Error("執行存儲過程%s出錯%v", proName, err)
		logret := &netproto.DZPKHALLChangeFaceRet{}
		logret.Code = proto.Int32(0)
		logret.Message = proto.String("服務器錯誤.")
		retmsg.MsgData = logret

	} else {
		mlog.Debug("[%s] 執行成功 ret=%v", proName, ret)
		logret := &netproto.DZPKHALLChangeFaceRet{}
		code := ret.GetReturnValue()
		logret.Code = proto.Int32(int32(code))
		logret.CurrentMoney = proto.Int64(ret.GetOutputParamValue("lngCurrentAmount").(int64))
		logret.CurrentFaceUrl = proto.String(ret.GetOutputParamValue("chvDBFaceUrl").(string))
		logret.NextCost = proto.Int32(int32(ret.GetOutputParamValue("intNextCost").(int64)))
		retmsg.MsgData = logret
	}

	return []*msg.Message{retmsg}
}
