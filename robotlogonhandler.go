package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/new888-cryptocity/netproto"

	proto "github.com/golang/protobuf/proto"
)

//機器人登入處理器
type RobotLogonHandler struct {
}

func (h *RobotLogonHandler) Init() {}
func (h *RobotLogonHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_LogonRobotID)
}

func (h *RobotLogonHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LogonRobot{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.LogonRobot)

	ctx.Info("收到機器人登入數據%v", mm)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intServerID", mm.GetServerID())
	ps.AddIntInput("intCurrentRobotCount", mm.GetCurrentRobotCount())
	ps.AddBigIntInput("lngMinMoneyAmount", mm.GetMinMoneyAmount())
	ps.AddBigIntInput("lngMaxMoneyAmount", mm.GetMaxMoneyAmount())
	ps.AddIntOutput("intLogoutRobotCount", 0)
	ps.AddIntInput("intGroupID", mm.GetLianyunID())
	proc := db.NewProcedure("PrGs_RobotLogon", ps)
	dbret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_DBServer), int32(netproto.DBServerClassID_LogonRobotRetID), nil)
	retdata := new(netproto.LogonRobotRet)

	//如果出錯將數量設為0
	logonRobotCount := int32(0)
	if err != nil {
		strMsg := fmt.Sprintf("執行存儲過程PrGs_RobotLogon錯誤, %v", err)
		ctx.Error(strMsg)
	} else {
		logonRobotCount = int32(dbret.GetOutputParamValue("intLogoutRobotCount").(int64))
		if logonRobotCount > 0 && dbret.GetRetTableCount() > 1 {
			tbRobot := dbret[0]
			for i := 0; i < len(tbRobot.Rows); i++ {
				robotData := new(netproto.RobotData)
				robotData.UserID = proto.Int32(int32(tbRobot.GetValueByColName(i, "UserID").(int64)))
				robotData.UserType = proto.Int32(int32(tbRobot.GetValueByColName(i, "UserType").(int64)))
				robotData.NickName = proto.String(tbRobot.GetValueByColName(i, "NickName").(string))
				robotData.FaceID = proto.Int32(int32(tbRobot.GetValueByColName(i, "FaceID").(int64)))
				robotData.Sex = proto.Int32(int32(tbRobot.GetValueByColName(i, "Sex").(int64)))
				robotData.CashAmount = proto.Int64(tbRobot.GetValueByColName(i, "CashAmount").(int64))
				robotData.WinCount = proto.Int32(int32(tbRobot.GetValueByColName(i, "WinCount").(int64)))
				robotData.LoseCount = proto.Int32(int32(tbRobot.GetValueByColName(i, "LoseCount").(int64)))
				robotData.DrawCount = proto.Int32(int32(tbRobot.GetValueByColName(i, "DrawCount").(int64)))
				robotData.RobotLevelID = proto.Int32(int32(tbRobot.GetValueByColName(i, "LevelID").(int64)))
				robotData.TotalScore = proto.Int64(int64(tbRobot.GetValueByColName(i, "TotalScore").(int64)))

				retdata.UserData = append(retdata.UserData, robotData)
			}
		}

	}

	retdata.RobotCount = proto.Int32(logonRobotCount)
	retmsg.MsgData = retdata

	ctx.Info("返回機器人登入數據%v", retdata)

	return []*msg.Message{retmsg}
}
