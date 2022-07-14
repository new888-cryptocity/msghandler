package msghandler

import (
	"regexp"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"666.com/gameserver/netproto"
	"github.com/golang/protobuf/proto"
)

type BindZhifubaoHandler struct {
}

func (rh *BindZhifubaoHandler) Init() {}
func (rh *BindZhifubaoHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_Hall) && sclsid == int32(netproto.HallMsgClassID_BindZhifubaoID)
}

func (rh *BindZhifubaoHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpcinfo, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.ZhifubaoInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.ZhifubaoInfo)
	zhifubao := mm.GetZhifubao()
	realname := mm.GetRealName()

	NeedVcode := mm.GetNeedVcode()
	VCode := mm.GetVcode()

	userid := rpcinfo.GetUserID()
	cer := rpcinfo.GetCer()

	var code int32 = 0
	retmessage := ""
	if tret, err := regexp.Match(".*[!-+<>(){}=!;:&'\"]+.*", []byte(zhifubao)); tret && err == nil {
		ctx.Error("参数包含非法字符, %v", zhifubao)
		code = 0
		retmessage = "账号包含非法字符"
	} else if tret, err := regexp.Match(".*[!-+<>(){}=!;:&'\"]+.*", []byte(realname)); tret && err == nil {
		ctx.Error("参数包含非法字符, %v", realname)
		code = 0
		retmessage = "名字包含非法字符"
	} else {
		d := GetDatabase()
		ps := db.NewSqlParameters()
		ps.AddIntInput("intUserID", userid)
		ps.AddVarcharInput("chvCer", cer, 32)
		ps.AddVarcharInput("chvRealName", realname, 64)
		ps.AddVarcharInput("chvAlipayAccount", zhifubao, 64)
		ps.AddVarcharOutput("chvErrMsg", "")

		ps.AddIntInput("intNeedVcode", NeedVcode)
		ps.AddVarcharInput("chvVcode", VCode, 10)

		proc := db.NewProcedure("PrPs_BindAlipayAccount", ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			ctx.Error("执行存储过程PrPs_BindAlipayAccount错误, %v", err)
			code = 0
			retmessage = "服务器错误"
		} else {
			if ret.GetReturnValue() == 1 {
				code = 1
			}
			retmessage = ret.GetOutputParamValue("chvErrMsg").(string)
		}
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Hall)
	retmsg.SClassID = int32(netproto.HallMsgClassID_BindZhifubaoRetID)
	retmsg.MsgData = getRetMsg(code, retmessage)

	return []*msg.Message{retmsg}
}
