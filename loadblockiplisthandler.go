package msghandler

import (
	"fmt"

	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/netproto"
)

//加載封鎖IP
type LoadBlockIPListHandler struct {
}

func (rh *LoadBlockIPListHandler) Init() {}
func (rh *LoadBlockIPListHandler) IsHook(bclsid int32, sclsid int32) bool {
	return bclsid == int32(netproto.MessageBClassID_DBServer) && sclsid == int32(netproto.DBServerClassID_LoadBlockIPID)
}

func (rh *LoadBlockIPListHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	d := GetDatabase()
	ps := db.NewSqlParameters()
	proc := db.NewProcedure("PrSys_LoadBlockIPList", ps)
	dbret, err := d.ExecProc(proc)

	retmsg := msg.NewMessage(int32(netproto.MessageBClassID_DBServer), int32(netproto.DBServerClassID_BlockIPListID), nil)
	retdata := new(netproto.BlockIPList)

	if err != nil {
		strMsg := fmt.Sprintf("執行存儲過程PrSys_LoadBlockIPList錯誤, %v", err)
		ctx.Error(strMsg)
	} else {
		if dbret.GetRetTableCount() > 0 {
			tbIPLst := dbret[0]
			for i := 0; i < len(tbIPLst.Rows); i++ {
				ips := tbIPLst.GetValueByColName(i, "IPAddress").(string)

				retdata.IPAddress = append(retdata.IPAddress, ips)
			}
		}
	}

	retmsg.MsgData = retdata

	return []*msg.Message{retmsg}
}
