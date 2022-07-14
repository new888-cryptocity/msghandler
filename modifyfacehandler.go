package msghandler

import (
	"reflect"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	proto "github.com/golang/protobuf/proto"
)

//修改頭像
type ModifyFaceHandler struct {
}

func (rh *ModifyFaceHandler) Init() {}
func (rh *ModifyFaceHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_ModifyFaceID)
}

func (rh *ModifyFaceHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ModifyFace{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.ModifyFace)

	userid := rpcinfo.GetUserID()
	ip := rpcinfo.GetIPAddress()
	faceid := mm.GetFaceID()

	d := GetDatabase()
	ps := db.NewSqlParameters()

	isNewApi := false
	checkType := reflect.TypeOf(*mm)
	if _, ok := checkType.FieldByName("FaceFrameID"); ok {
		//有頭像框字段
		faceFrameID := mm.GetFaceFrameID()
		ps.AddIntInput("intFaceFrameID", faceFrameID)
		isNewApi = true
	}
	//faceFrameID := mm.GetFaceFrameID()
	ps.AddIntInput("intUserID", userid)
	ps.AddIntInput("intFaceID", faceid)
	ps.AddVarcharInput("chvIPAddress", ip, 15)
	ps.AddVarcharOutput("chvErrMsg", "")

	proFunName := "PrPs_User_ModifyFaceID"
	if isNewApi {
		//使用新修改頭像接口
		proFunName = "PrPs_User_ModifyFaceIDNew"
	}
	proc := db.NewProcedure(proFunName, ps)
	ret, err := d.ExecProc(proc)

	var code int32 = 0
	retmessage := ""
	if err != nil {
		ctx.Error("執行存儲過程PrPs_User_ModifyFaceID錯誤, %v", err)
		code = 0
		retmessage = "服務器錯誤"
	} else {
		if ret.GetReturnValue() == 1 {
			code = 1
		}
		retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_ModifyFaceRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
