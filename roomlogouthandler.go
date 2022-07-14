package msghandler

import (
	"fmt"
	"time"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"https://github.com/new888-cryptocity/netproto"
	"github.com/golang/protobuf/proto"
)

type RoomLogoutHandler struct {
}

func (h *RoomLogoutHandler) Init() {}
func (h *RoomLogoutHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid == int32(netproto.MessageBClassID_DBServer) {
		if sclsid == int32(netproto.DBServerClassID_LogoutRoomID) {
			return true
		}
	} else if bclsid == int32(netproto.MessageBClassID_Hall) {
		if sclsid == int32(netproto.HallMsgClassID_CleanUserGameListID) || sclsid == int32(netproto.HallMsgClassID_GetUserGameListID) {
			return true
		}
	}
	return false

	//return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_LogoutRoomID)
}

func (h *RoomLogoutHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	if m.BClassID == int32(netproto.MessageBClassID_DBServer) {
		return h.logoutRoom(ctx, clt, m)
	} else if m.BClassID == int32(netproto.MessageBClassID_Hall) {
		if m.SClassID == int32(netproto.HallMsgClassID_CleanUserGameListID) {
			return h.cleanUserGameList(ctx, clt, m)
		} else if m.SClassID == int32(netproto.HallMsgClassID_GetUserGameListID) {
			return h.getUserGameList(ctx, clt, m)
		}
	}

	return nil
}

func (h *RoomLogoutHandler) logoutRoom(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.LogoutGameRoomInfo{}
		return mm
	})

	if rpc == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.LogoutGameRoomInfo)

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intMsgID", rpc.GetQueueID())
	ps.AddIntInput("intUserID", mm.GetUserID())
	ps.AddBigIntInput("lngUpdateAmount", mm.GetUpdateAmount())
	ps.AddBigIntInput("lngTaxAmount", mm.GetTaxAmount())
	ps.AddIntInput("intWinCount", mm.GetWinCount())
	ps.AddIntInput("intLoseCount", mm.GetLoseCount())
	ps.AddIntInput("intDrawCount", mm.GetDrawCount())
	ps.AddVarcharInput("dtmLoginTime", mm.GetLoginTime(), 32)
	ps.AddVarcharInput("dtmLogoutTime", mm.GetLogoutTime(), 32)
	ps.AddVarcharInput("chvIPAddress", mm.GetIPAddress(), 15)
	ps.AddIntInput("tnyHDType", mm.GetHDType())
	ps.AddVarcharInput("chvHDCode", mm.GetHDCode(), 64)
	ps.AddIntInput("intGameTime", mm.GetGameTime())
	ps.AddIntInput("sintGameID", mm.GetGameID())
	ps.AddIntInput("sintServerID", mm.GetServerID())
	ps.AddBigIntInput("lngLoginMoney", mm.GetLoginMoney())
	ps.AddIntInput("tnyScoreUpdateMode", mm.GetScoreUpdateMode())
	ps.AddVarcharOutput("chvErrMsg", "")

	proc := db.NewProcedure("PrGs_UserLogout", ps)
	mlog.Debug(`執行存儲過程 PrGs_UserLogout intMsgID:[%v] intUserID:[%v], lngUpdateAmount:[%v], lngTaxAmount:[%v], intWinCount:[%v], intLoseCount:[%v], intDrawCount:[%v], 
	dtmLoginTime:[%v], dtmLogoutTime:[%v], chvIPAddress:[%v], tnyHDType:[%v], chvHDCode:[%v], intGameTime:[%v], sintGameID:[%v], sintServerID:[%v]
	lngLoginMoney:[%v], tnyScoreUpdateMode:[%v]`,
		rpc.GetQueueID(), mm.GetUserID(), mm.GetUpdateAmount(), mm.GetTaxAmount(), mm.GetWinCount(), mm.GetLoseCount(), mm.GetDrawCount(),
		mm.GetLoginTime(), mm.GetLogoutTime(), mm.GetIPAddress(), mm.GetHDType(), mm.GetHDCode(), mm.GetGameTime(), mm.GetGameID(), mm.GetServerID(),
		mm.GetLoginMoney(), mm.GetScoreUpdateMode())

	_, err := d.ExecProc(proc)
	var retcode int32 = 1
	if err != nil {
		ctx.Error("執行存儲過程PrGs_UserLogout錯誤, %v", err)
		ctx.Error("PrGs_UserLogout失敗,intMsgID: %d, LogoutGameRoomInfo: %v", rpc.GetQueueID(), mm)
		retcode = 0

		for i := 0; err != nil && i < 100 && mm.GetServerID()/100 == 30; i++ { //鬥地主比賽退房間失敗需重試
			_, err = d.ExecProc(proc) //重試一次
			if err == nil {
				retcode = 1
				ctx.Error("第%v次%v鬥地主比賽退房間重試成功", i+1, mm.GetUserID())
				break
			}
		}
	}

	if err != nil && mm.GetServerID()/100 == 30 {
		ctx.Error("/*【budan0_start需補單】*/ USE CenterDB; %s \nGO\n/*【budan0_end需補單】*/", proc.GetSqlString())
	}

	return []*msg.Message{GetDBRetMsg(retcode)}
}

func (h *RoomLogoutHandler) cleanUserGameList(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.CleanUserGameList{}
		return mm
	})

	if rpc == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.CleanUserGameList)

	d := GetDatabase()
	sql := fmt.Sprintf("DELETE FROM [CenterDB].[Game].[GameUserList] WHERE UserID = %d AND GroupID = %d", mm.GetUserID(), mm.GetChannel())
	_, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error("CleanUserGameList Err:", err.Error())
		return []*msg.Message{GetDBRetMsg(1)}
	}
	return []*msg.Message{GetDBRetMsg(0)}
}

func (h *RoomLogoutHandler) getUserGameList(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	rpc, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.GetUserGameList{}
		return mm
	})

	if rpc == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.GetUserGameList)

	data := new(netproto.GetUserGameList)
	data.UserID = proto.Int32(mm.GetUserID())

	resp := new(msg.Message)
	resp.BClassID = int32(netproto.MessageBClassID_Hall)
	resp.SClassID = int32(netproto.HallMsgClassID_GetUserGameListID)
	resp.MsgData = data

	d := GetDatabase()
	sql := fmt.Sprintf("SELECT ServerID, LoginTime From [CenterDB].[Game].[GameUserList] WHERE UserID = %d AND GroupID = %d", mm.GetUserID(), mm.GetChannel())
	ret, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error("getUserGameList Err:", err.Error())
	}

	if len(ret) == 0 {
		return []*msg.Message{resp}
	}

	ret[0].ForEachRow(func(r db.RecordRow) {
		serverID := r[0].(int32)
		loginTime := r[1].(time.Time)
		loginTimeUnix := loginTime.Unix()
		mlog.Debug("ServerID = %d", serverID)
		mlog.Debug("loginTime = %d", loginTimeUnix)

		sql = fmt.Sprintf("SELECT GameID From [CenterDB].[ServerInfo].[GameServerListBase] WHERE ServerID = %d ", serverID)
		ret, err = d.ExecSql(sql)
		if err != nil {
			mlog.Error("getUserGameList Err:", err.Error())
		}

		if len(ret) == 0 {
			return
		}
		gameID := ret[0].GetSingleValueInt32("GameID")
		mlog.Debug("GameID =  %d", gameID)
		data.GameList = append(data.GameList, &netproto.UserGameListInfo{GameID: &gameID, LoginTime: &loginTimeUnix})

	})
	/*
			serverID := ret[0].GetSingleValueInt32("ServerID")
			loginTime := ret[0].GetSingleValue("LoginTime").(time.Time)
			loginTimeUnix := loginTime.Unix()
			mlog.Debug("ServerID = %d", serverID)
			mlog.Debug("loginTime = %d", loginTimeUnix)


		sql = fmt.Sprintf("SELECT GameID From [CenterDB].[ServerInfo].[GameServerListBase] WHERE ServerID = %d ", serverID)
		ret, err = d.ExecSql(sql)
		if err != nil {
			mlog.Error("getUserGameList Err:", err.Error())
		}

		if len(ret) == 0 {
			return []*msg.Message{resp}
		}
		gameID := ret[0].GetSingleValueInt32("GameID")
		mlog.Debug("GameID =  %d", gameID)
		data.GameList = append(data.GameList, &netproto.UserGameListInfo{GameID: &gameID, LoginTime: &loginTimeUnix})
	*/
	return []*msg.Message{resp}

}
