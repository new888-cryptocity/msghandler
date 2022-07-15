package msghandler

import (
	"crypto/sha256"
	"fmt"

	"github.com/xinholib/cybavo"
	"666.com/gameserver/dbserver/src/dal/utility"
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/module"
	"666.com/gameserver/framework/msg"
	"666.com/gameserver/framework/network"
	"666.com/gameserver/framework/rpcmsg"
	"github.com/xinholib/netproto"
	proto "github.com/golang/protobuf/proto"
)

//百家樂牌組操作
type BJL_TableCardHandler struct {
}

func (b *BJL_TableCardHandler) Init() {}
func (b *BJL_TableCardHandler) IsHook(bclsid int32, sclsid int32) bool {
	if bclsid != int32(netproto.MessageBClassID_DBServer) {
		return false
	}
	switch sclsid {
	case int32(netproto.DBServerClassID_BJL_SetCard),
		int32(netproto.DBServerClassID_BJL_GetConfig),
		int32(netproto.DBServerClassID_BJL_OpenCard),
		int32(netproto.DBServerClassID_BJL_SetPoker),
		int32(netproto.DBServerClassID_BJL_GetShoe),
		int32(netproto.DBServerClassID_BJL_SetShoe):
		return true

	}
	return false
}

func (b *BJL_TableCardHandler) OnMessage(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	mlog.Info("[BJL_TableCardHandler][OnMessage] %d %d", m.BClassID, m.SClassID)
	switch m.SClassID {
	case int32(netproto.DBServerClassID_BJL_SetCard):
		return b.SetCard(ctx, clt, m)
	case int32(netproto.DBServerClassID_BJL_GetConfig):
		return b.GetConfig(ctx, clt, m)
	case int32(netproto.DBServerClassID_BJL_OpenCard):
		return b.OpenCard(ctx, clt, m)
	case int32(netproto.DBServerClassID_BJL_SetPoker):
		return b.SetPoker(ctx, clt, m)
	case int32(netproto.DBServerClassID_BJL_SetShoe):
		return b.SetShoe(ctx, clt, m)
	case int32(netproto.DBServerClassID_BJL_GetShoe):
		return b.GetShoe(ctx, clt, m)
	}
	return nil
}

func (b *BJL_TableCardHandler) SetCard(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJL_DBCardInfoList{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BJL_DBCardInfoList)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_BJL_SetCard)
	retmsg.MsgData = mm

	d := GetDatabase()

	// 判斷是否將牌組加密資料上鏈
	isChain := mm.GetIsBlockchain()
	if isChain {
		isChain = ctx.Config.GetConfig("db").GetConfigInt32("Blockchain", "enable") > 0
	}
	var count int32 = 0
	if isChain == true {
		count = b.GetPokerUseCount(d, mm.GetTableNo(), mm.GetShoeNo(), mm.GetGroupID())
	}
	pokerEncode := ""
	spName := "BJL_PrGs_SetCard"
	if mm.GetFlag() == 0 {
		spName += "_Guest"
	}
	for i, tableInfo := range mm.Cards {

		if isChain == true && i < int(count) {
			if i == 0 {
				tableSN := utility.GenTableSN(mm.GetTableNo(), mm.GetShoeNo())
				pokerEncode = fmt.Sprintf("%s,%d,%s", tableSN, count, tableInfo.GetCode())
			} else {
				pokerEncode += fmt.Sprintf(",%s", tableInfo.GetCode())
			}
		}

		ps := db.NewSqlParameters()
		id := fmt.Sprintf("%d-%d-%d", mm.GetTableNo(), mm.GetShoeNo(), tableInfo.GetCardNo())
		ps.AddVarcharInput("ID", id, 50)
		ps.AddIntInput("TableNo", mm.GetTableNo())
		ps.AddIntInput("ShoeNo", mm.GetShoeNo())
		ps.AddIntInput("CardNo", tableInfo.GetCardNo())
		ps.AddVarcharInput("CardCode", tableInfo.GetCode(), 128)
		ps.AddVarcharInput("CardEncode", tableInfo.GetEncode(), 128)
		ps.AddVarcharInput("Card", tableInfo.GetCard(), 50)
		channelID := int32(mm.GetGroupID())
		ps.AddIntInput("ChannelID", channelID)
		cardTableName := mm.GetGameSN()
		ps.AddVarcharInput("CardTableName", cardTableName, 50)

		proc := db.NewProcedure(spName, ps)
		_, err := d.ExecProc(proc)

		if err != nil {
			mlog.Error("SetCard : %s error:%v", spName, err)
		}
	}

	if pokerEncode != "" {
		// 將牌組上鏈
		go func() {
			// 取得牌組加密資料
			h := sha256.New()
			h.Write([]byte(pokerEncode))
			hex := fmt.Sprintf("%x", h.Sum(nil))

			// 取得牌靴網址
			ip := ctx.Config.GetConfig("db").GetConfigValue("web", "shoeIP")
			url := utility.GenShoeURL(ip, mm.GetGameID(), mm.GetTableNo(), mm.GetShoeNo())

			// 取得 orderID
			orderID := cybavo.GenOrderID(mm.GetGameID(), mm.GetGroupID(), mm.GetTableNo(), mm.GetShoeNo())

			// 上鏈
			result, err := cybavo.AddingPokerBlocksToChain(orderID, url, hex)
			if err != nil {
				mlog.Error("cybavo.AddingPokerBlocksToChain Error!", err.Error())
			}

			utility.SaveBlocksInfoToFile(orderID, url, hex, pokerEncode, result)
			utility.SetBlokcsInfoToDB(d, "BJL_Shoe", orderID, mm.GetTableNo(), mm.GetShoeNo(), hex, pokerEncode)
		}()
	}

	return []*msg.Message{retmsg}
}

func (b *BJL_TableCardHandler) OpenCard(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {

	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJL_DBCardInfoList{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_BJL_OpenCard)

	mm := rmm.(*netproto.BJL_DBCardInfoList)
	retmsg.MsgData = mm
	maxCardIndex := int32(0)
	d := GetDatabase()
	for _, tableInfo := range mm.GetCards() {

		if tableInfo.GetCardNo() > maxCardIndex {
			maxCardIndex = tableInfo.GetCardNo()
		}

		ps := db.NewSqlParameters()
		id := fmt.Sprintf("%d-%d-%d", mm.GetTableNo(), mm.GetShoeNo(), tableInfo.GetCardNo())
		ps.AddVarcharInput("ID", id, 50)
		proc := db.NewProcedure("BJL_PrGs_OpenCard", ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			mlog.Error("OpenCard : BJL_PrGs_OpenCard error:%v", err)
		}

		b.parseBJLCardRet(ret, retmsg)
	}

	return nil
}
func (b *BJL_TableCardHandler) SetPoker(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJL_DBPokerInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BJL_DBPokerInfo)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Game)
	retmsg.SClassID = int32(netproto.DBServerClassID_BJL_SetPoker)
	retmsg.MsgData = mm
	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("TableSN", mm.GetGameSN(), 50)
	ps.AddIntInput("TotalCount", mm.GetTotalCount())
	ps.AddIntInput("UseCount", mm.GetUseCount())
	ps.AddIntInput("CurrCount", mm.GetCurrCount())
	ps.AddIntInput("Games", mm.GetGames())
	channelID := int32(mm.GetGroupID())
	ps.AddIntInput("ChannelID", channelID)

	spName := ""
	if mm.GetFlag() == 0 {
		spName = "BJL_PrGs_SetPoker_Guest"
	} else {
		spName = "BJL_PrGs_SetPoker"
	}

	proc := db.NewProcedure(spName, ps)
	ret, err := d.ExecProc(proc)

	if err != nil {
		mlog.Error("%s error:%v", spName, err)
	} else {
		mlog.Info("[BJL_TableCardHandler][SetPoker] TableSN=%s CurrCount=%d", mm.GetGameSN(), mm.GetCurrCount())

		if ret != nil && len(ret[0].Rows[0]) > 0 {
			cardTableName := ret[0].Rows[0][0].(string)
			mm.GameSN = &cardTableName
		}

	}

	return []*msg.Message{retmsg}
}

func (b *BJL_TableCardHandler) SetShoe(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJL_DBTableShoeInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BJL_DBTableShoeInfo)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_DBServer)
	retmsg.SClassID = int32(netproto.DBServerClassID_BJL_SetShoe)
	retmsg.MsgData = mm
	d := GetDatabase()

	spName := ""
	if mm.GetFlag() == 0 {
		spName = "BJL_PrGs_SetShoe_Guest"
	} else {
		spName = "BJL_PrGs_SetShoe"
	}

	for i := 0; i < len(mm.GetTableNo()); i++ {
		ps := db.NewSqlParameters()
		ps.AddIntInput("TableNo", mm.GetTableNo()[i])
		ps.AddIntInput("ShoeNo", mm.GetShoeNo()[i])
		ps.AddIntInput("Channel", mm.GetGroupID())
		proc := db.NewProcedure(spName, ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			mlog.Error("%s error:%v", spName, err)
		}

		b.parseBJLCardRet(ret, retmsg)
	}

	return []*msg.Message{retmsg}
}

func (b *BJL_TableCardHandler) GetShoe(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJL_DBTableShoeInfo{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BJL_DBTableShoeInfo)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Game)
	retmsg.SClassID = int32(netproto.DBServerClassID_BJL_GetShoe)
	data := new(netproto.BJL_DBTableShoeInfo)
	data.TableNo = make([]int32, 0)
	data.ShoeNo = make([]int32, 0)

	d := GetDatabase()

	spName := ""
	if mm.GetFlag() == 0 {
		spName = "BJL_PrGs_SetShoe_Guest"
	} else {
		spName = "BJL_PrGs_SetShoe"
	}

	for i := 0; i < len(mm.GetTableNo()); i++ {
		ps := db.NewSqlParameters()
		tableNo := mm.GetTableNo()[i]
		ps.AddIntInput("TableNo", tableNo)
		ps.AddIntInput("ShoeNo", 0) // ShoeNo = 0，表示要取得此TableNo 的 ShoeNo
		ps.AddIntInput("Channel", mm.GetGroupID())
		proc := db.NewProcedure(spName, ps)
		ret, err := d.ExecProc(proc)

		if err != nil {
			mlog.Error("%s error:%v", spName, err)
			continue
		}

		if ret != nil && len(ret[0].Rows[0]) > 0 {
			shoe := ret[0].Rows[0][0].(int64)
			data.TableNo = append(data.TableNo, tableNo)
			data.ShoeNo = append(data.ShoeNo, int32(shoe))
		}
	}
	retmsg.MsgData = data

	return []*msg.Message{retmsg}
}

// GetConfig 取得牌桌設定
func (b *BJL_TableCardHandler) GetConfig(ctx *module.MouduleCtx, clt *network.ConnInfo, m *msg.Message) []*msg.Message {
	_, rmm := rpcmsg.ParseRPCMessage(m, func() proto.Message {
		mm := &netproto.BJL_DBTableConfig{}
		return mm
	})

	if rmm == nil {
		return []*msg.Message{GetDBRetMsg(0)}
	}

	mm := rmm.(*netproto.BJL_DBTableConfig)

	retmsg := new(msg.Message)
	retmsg.BClassID = int32(netproto.MessageBClassID_Game)
	retmsg.SClassID = int32(netproto.DBServerClassID_BJL_GetConfig)

	spName := ""
	if mm.GetFlag() == 0 {
		spName = "BJL_PrGs_GetConfig_Guest"
	} else {
		spName = "BJL_PrGs_GetConfig"
	}

	d := GetDatabase()
	ps := db.NewSqlParameters()
	ps.AddIntInput("intID", mm.GetTableID())
	ps.AddIntInput("intConfigID", mm.GetChannel())
	proc := db.NewProcedure(spName, ps)
	ret, err := d.ExecProc(proc)
	if err != nil {
		mlog.Error("%s error:%v", spName, err)
	}

	data := utility.ParserBJLConfig(ret)

	retmsg.MsgData = data

	return []*msg.Message{retmsg}
}

func (b *BJL_TableCardHandler) GetPokerEncodeFromDB(d *db.Database, tableNo int32, shoeNo int32, channel int32) (int32, string, error) {
	output := ""
	sql := fmt.Sprintf("SELECT TOP (1) [CardTableName],[TotalCount],[UseCount],[CurrCount] FROM [BJL_Shoe].[Table].[Poker] where TableSN ='%d-%d' and Channel = %d", tableNo, shoeNo, channel)
	ret, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error("GetPokerEncode Err:", err.Error())
		return 0, output, err
	}

	cardTableName := ret[0].GetValueByColName(0, "CardTableName").(string)
	count := ret[0].GetValueByColName(0, "UseCount").(int64)
	mlog.Debug("cardTableName = " + cardTableName)

	sql = fmt.Sprintf("SELECT TOP (%d) [CardEncode] FROM [BJL_Shoe].[Table].[%s] where TableNo = %d and ShoeNo = %d and Channel = %d order by CardNo", count, cardTableName, tableNo, shoeNo, channel)
	ret, err = d.ExecSql(sql)
	if err != nil {
		mlog.Error("GetPokerEncode Err:", err.Error())
		return 0, output, err
	}
	i := int64(0)
	ret[0].ForEachRow(func(r db.RecordRow) {
		for _, row := range r {
			output = output + fmt.Sprintf("%v,", row)
			i++
		}
	})
	if count != i {
		mlog.Warn("GetPokerEncode Length not correct! need %d, but has %d", count, i)
	}
	return int32(i), output, nil
}

func (b *BJL_TableCardHandler) GetPokerUseCount(d *db.Database, tableNo int32, shoeNo int32, channel int32) int32 {
	sql := fmt.Sprintf("SELECT TOP (1) [CardTableName],[TotalCount],[UseCount],[CurrCount] FROM [BJL_Shoe].[Table].[Poker] where TableSN ='%d-%d' and Channel = %d", tableNo, shoeNo, channel)
	ret, err := d.ExecSql(sql)
	if err != nil {
		mlog.Error("GetPokerEncode Err:", err.Error())
		return 0
	}

	cardTableName := ret[0].GetValueByColName(0, "CardTableName").(string)
	count := ret[0].GetValueByColName(0, "UseCount").(int64)
	mlog.Debug("cardTableName = " + cardTableName)
	return int32(count)
}

func (b *BJL_TableCardHandler) parseBJLCardRet(ret db.RecordSet, retmsg *msg.Message) []*msg.Message {
	return []*msg.Message{retmsg}
}
