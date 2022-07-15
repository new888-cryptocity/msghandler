package msghandler

import (
	"666.com/gameserver/framework/db"
	"666.com/gameserver/framework/mlog"
	"666.com/gameserver/framework/msg"
	"github.com/xinholib/netproto"
	"github.com/golang/protobuf/proto"
)

//数据库连接字符串
var ConnString string
var DealLogConnString string
var PoolSize int = 100
var database *db.Database
var dealLogDataBase *db.Database

// 為了黑傑克增加的
//var DbUtil db.DatabaseEngine
//var DealLogDbUtil db.DatabaseEngine
//var ClubDbUtil db.DatabaseEngine
//var WildCatDbUtil db.DatabaseEngine
//var LogDbUtil db.DatabaseEngine
/*
type DBUtilD struct {
}

func (DBUtilD) GetDbutil(args int) db.DatabaseEngine {
	return DbUtil
}

func (DBUtilD) GetDealLogDbUtil() db.DatabaseEngine {
	return DealLogDbUtil
}

func (DBUtilD) GetClubDbUtil() db.DatabaseEngine {
	return ClubDbUtil
}

func (DBUtilD) GetLogDbUtil() db.DatabaseEngine {
	return LogDbUtil
}
*/
func InitDatabase() {
	database = db.NewDatabase(ConnString, PoolSize)
	dealLogDataBase = db.NewDatabase(DealLogConnString, PoolSize)

	//DbUtil = db.NewDatabaseEngine(database)
	//DealLogDbUtil = db.NewDatabaseEngine(dealLogDataBase)
	//ClubDbUtil = db.NewDatabaseEngine(clubDataBase)
	//WildCatDbUtil = db.NewDatabaseEngine(wildCatDataBase)
	//LogDbUtil = db.NewDatabaseEngine(logDbDataBase)

	//command.AddCommand(NewDBUtilCmd())
}

func GetDatabase() *db.Database {
	return database
}

func GetDealLogDatabase() *db.Database {
	return dealLogDataBase
}

//获取提示消息
func GetErrMsgValue(set db.RecordSet) string {
	rlen := len(set)

	return set[rlen-2].GetValueByColName(0, "chvErrMsg").(string)
}

//获取数据库执行结果消息
func GetDBRetMsg(retCode int32) *msg.Message {
	ret := new(netproto.DBServerRet)
	ret.Code = proto.Int32(retCode)

	msg := new(msg.Message)
	msg.BClassID = int32(netproto.MessageBClassID_DBServer)
	msg.SClassID = int32(netproto.DBServerClassID_DBRetID)
	msg.MsgData = ret

	return msg
}

//解析用户登录大廳结果
func parseUserLogonRet(ret db.RecordSet) *netproto.UserLoginRet {
	/*
		code := 0
		if ret.GetReturnValue() == 1 {
			code = 1
		}
	*/
	// nero 直接塞errorcode.
	code := ret.GetReturnValue()

	logret := &netproto.UserLoginRet{}
	logret.Code = proto.Int32(int32(code))
	logret.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))

	if code == 1 {
		logret.UserID = proto.Int32(int32(ret.GetOutputParamValue("intUserID").(int64)))
		logret.Cer = proto.String(ret.GetOutputParamValue("chvCertification").(string))

		userdata := new(netproto.UserHallLogonData)
		userdata.LoginName = proto.String(ret[0].GetSingleValue("LoginName").(string))
		userdata.NickName = proto.String(ret[0].GetSingleValue("NickName").(string))
		userdata.UserType = proto.Int32(ret[0].GetSingleValueInt32("UserType"))
		userdata.Sex = proto.Int32(ret[0].GetSingleValueInt32("Sex"))
		userdata.HeadID = proto.Int32(ret[0].GetSingleValueInt32("HeadID"))
		userdata.IsTopWindow = proto.Bool(ret[0].GetSingleValueInt32("IsTopWindow") == 1)
		userdata.CashAmount = proto.Int64(ret[0].GetSingleValueInt64("CashAmount"))
		userdata.BankAmount = proto.Int64(ret[0].GetSingleValueInt64("BankAmount"))
		userdata.IsGaming = proto.Bool(ret[0].GetSingleValue("IsGaming").(bool))
		userdata.ServerAddr = proto.String(ret[0].GetSingleValue("ServerAddr").(string))
		userdata.ServerName = proto.String(ret[0].GetSingleValue("ServerName").(string))
		userdata.IsKick = proto.Bool(ret[0].GetSingleValue("IsKick").(bool))
		userdata.IsBindGuest = proto.Bool(ret[0].GetSingleValueInt32("IsBindGuest") == 1)
		userdata.IsBindZhifubao = proto.Bool(ret[0].GetSingleValueInt32("IsBindZhifubao") == 1)
		userdata.Zhifubao = proto.String(ret[0].GetSingleValue("AlipayAccount").(string))
		userdata.RealName = proto.String(ret[0].GetSingleValue("RealName").(string))
		userdata.UserLevel = proto.Int32(ret[0].GetSingleValueInt32("UserLevel"))
		userdata.LevelKey = proto.String(ret[0].GetSingleValue("LevelKey").(string))
		userdata.AnnMsg = proto.String(ret[0].GetSingleValue("AnnMsg").(string))
		userdata.ConvertRateTipMsg = proto.String(ret[0].GetSingleValue("ConvertRateTipMsg").(string))
		userdata.BankPwdTipMsg = proto.String(ret[0].GetSingleValue("BankPwdTipMsg").(string))
		userdata.UIFlag = proto.String(ret[0].GetSingleValue("UIFlag").(string))
		userdata.PaySort = proto.String(ret[0].GetSingleValue("PaySort").(string))
		userdata.PayAmountConfig = proto.String(ret[0].GetSingleValue("PayAmountConfig").(string))
		userdata.PayNotifyMsg = proto.String(ret[0].GetSingleValue("PayNotifyMsg").(string))
		userdata.PayTips = proto.String(ret[0].GetSingleValue("PayTips").(string))
		userdata.NotifyFlag = proto.String(ret[0].GetSingleValue("NotifyFlag").(string))
		userdata.UpGradeMsg = proto.String(ret[0].GetSingleValue("UpGradeMsg").(string))
		userdata.LockGameID = proto.Int32(ret[0].GetSingleValueInt32("LockGameID"))
		userdata.IsBindBankCard = proto.Bool(ret[0].GetSingleValueInt32("IsBindBankAccount") == 1)
		userdata.BankCardNumber = proto.String(ret[0].GetSingleValue("BankAccountNumber").(string))
		userdata.BankCardName = proto.String(ret[0].GetSingleValue("BankAccountName").(string))
		userdata.BankName = proto.String(ret[0].GetSingleValue("BankName").(string))
		userdata.BankConvertRateTipMsg = proto.String(ret[0].GetSingleValue("BankConvertRateTipMsg").(string))
		userdata.InVGameID = proto.Int32(ret[0].GetSingleValueInt32("VGameID"))
		userdata.XiuXianAmount = proto.Int64(ret[0].GetSingleValueInt64("XiuXianScore"))
		userdata.XiuXianTotalCharge = proto.Int64(ret[0].GetSingleValueInt64("XiuxianTotalCharge"))
		userdata.LianyunID = proto.Int32(ret[0].GetSingleValueInt32("GroupID"))
		userdata.VipLv = proto.Int32(ret[0].GetSingleValueInt32("VipLv"))
		userdata.HeadFrameID = proto.Int32(ret[0].GetSingleValueInt32("HeadFrameID"))
		userdata.Language = proto.String(ret[0].GetSingleValue("Language").(string))
		userdata.Currency = proto.String(ret[0].GetSingleValue("Currency").(string))
		userdata.Denomination = proto.Int32(ret[0].GetSingleValueInt32("Denomination"))
		userdata.ChangeNickNameCost = proto.Int64(ret[0].GetSingleValueInt64("NickNameChangeCost"))
		userdata.FaceUrl = proto.String(ret[0].GetSingleValue("FaceUrl").(string))
		userdata.ChangeFaceCost = proto.Int64(ret[0].GetSingleValueInt64("FaceChangeCost"))
		userdata.InviteCode = proto.String(ret[0].GetSingleValue("InviteCode").(string))
		userdata.NickNameChangeCount = proto.Int32(ret[0].GetSingleValueInt32("NickNameChangeCount"))
		userdata.FaceUrlChangeCount = proto.Int32(ret[0].GetSingleValueInt32("FaceUrlChangeCount"))
		mlog.Debug("parseUserLogonRet [解析用户登录大廳结果] UserID:%d, LoginName:%s, Currency:%s, Denomination:%d, CashAmount:%d, BankAmount:%d",
			*logret.UserID, *userdata.LoginName, *userdata.Currency, *userdata.Denomination, *userdata.CashAmount, *userdata.BankAmount)

		//游戏
		if ret.GetRetTableCount() >= 3 {
			grt := ret[1]

			grt.ForEachRow(func(r db.RecordRow) {
				gamesortcate := new(netproto.GameSortCateInfo)
				gamesortcate.GameID = proto.Int32(int32(r[0].(int64)))
				gamesortcate.CategoryID = proto.String(r[1].(string))
				gamesortcate.Label = proto.Int32(int32(r[2].(int64)))

				userdata.GameList = append(userdata.GameList, gamesortcate)
			})
		}

		//分类列表
		if ret.GetRetTableCount() >= 4 {
			grt := ret[2]

			grt.ForEachRow(func(r db.RecordRow) {
				gamecate := new(netproto.GameCategoryInfo)
				gamecate.CategoryID = proto.Int32(int32(r[0].(int64)))
				gamecate.Name = proto.String(r[1].(string))

				userdata.GameCategoryList = append(userdata.GameCategoryList, gamecate)
			})
		}

		//已注册的视讯类游戏ID列表获取版本信息
		if ret.GetRetTableCount() >= 5 {
			grt := ret[3]

			grt.ForEachRow(func(r db.RecordRow) {
				gid := r[0].(int64)

				userdata.VGameIDS = append(userdata.VGameIDS, int32(gid))
			})

		}

		//版本列表
		if ret.GetRetTableCount() >= 6 {
			grt := ret[4]

			grt.ForEachRow(func(r db.RecordRow) {
				ver := new(netproto.SkinVersionInfo)
				ver.ID = proto.Int32(int32(r[0].(int64)))
				ver.BunldID = proto.String(r[1].(string))
				ver.Ver = proto.String(r[2].(string))
				ver.Path = proto.String(r[3].(string))
				ver.Ver1 = proto.String(r[4].(string))
				ver.Platform = proto.String(r[5].(string))
				ver.Channel = proto.String(r[6].(string))
				ver.SkinVer = proto.String(r[7].(string))
				ver.LimitIP = proto.String(r[8].(string))

				userdata.VersionList = append(userdata.VersionList, ver)
			})
		}

		logret.UserData = userdata
	} else {
		mlog.Warn("parseUserLogonRet [解析用户登录大廳结果] 失敗　Code:%v, Message:%v", *logret.Code, *logret.Message)
	}

	return logret
}

/*
func updateProtectNodeStatus(userid int32, token string, status int32) {
	ps := db.NewSqlParameters()
	ps.AddVarcharInput("chvToken", token, 10)
	ps.AddIntInput("intLoginUserID", userid)
	ps.AddIntInput("intStatus", status)
	proc := db.NewProcedure("PrPs_UpdateProtectNodeStatus", ps)
	ret, err := DbUtil.ExecProc(proc)
	defer func() {
		if err := recover(); err != nil {
			errorStr := util.GetStackInfo()
			mlog.NewDefaultLogger().SetData(ret).Error("发生错误%s, 執行存儲過程%s", errorStr, proc.GetSqlString())
			panic(err)
		}
	}()
	//fmt.Println(proc.GetSqlString())
	if err != nil {
		panic(err)
	}
}
*/

func getRetMsg(code int32, msgContent string) *netproto.RetMessage {
	retmsg := new(netproto.RetMessage)
	retmsg.Code = proto.Int32(code)
	retmsg.Message = proto.String(msgContent)

	return retmsg
}

func getRetMsgAddInvite(code int32, msgContent string, invitecode string) *netproto.RetMessage {
	retmsg := new(netproto.RetMessage)
	retmsg.Code = proto.Int32(code)
	retmsg.Message = proto.String(msgContent)
	retmsg.InviteCode = proto.String(invitecode)
	return retmsg
}

func parseActivityInfoRet(ret db.RecordSet) *netproto.ActivityInfoRet {
	/*
		code := 0
		if ret.GetReturnValue() == 1 {
			code = 1
		}
	*/
	// nero 直接塞errorcode.
	code := ret.GetReturnValue()

	activityInfoRet := &netproto.ActivityInfoRet{}
	activityInfoRet.Code = proto.Int32(int32(code))
	activityInfoRet.Message = proto.String(ret.GetOutputParamValue("chvErrMsg").(string))

	if code == 1 {
		userActivityInfo := new(netproto.UserActivityInfo)
		userActivityInfo.UserID = proto.Int32(ret[0].GetSingleValueInt32("UserID"))
		userActivityInfo.TotalEnroll = proto.Int64(ret[0].GetSingleValueInt64("TotalEnroll"))
		userActivityInfo.CurrentAmount = proto.Int64(ret[0].GetSingleValueInt64("CurrentAmount"))
		userActivityInfo.TotalAmount = proto.Int64(ret[0].GetSingleValueInt64("TotalAmount"))
		userActivityInfo.Status = proto.Int32(ret[0].GetSingleValueInt32("Status"))

		activityInfoRet.UserActivityInfo = userActivityInfo

		//游戏基础信息
		if ret.GetRetTableCount() >= 3 {
			baseInfoRet := ret[1]
			activityBaseInfo := new(netproto.ActivityBaseInfo)
			activityBaseInfo.ActivityID = proto.Int32(baseInfoRet.GetSingleValueInt32("ActivityID"))
			activityBaseInfo.ActivityName = proto.String(baseInfoRet.GetSingleValue("ActivityName").(string))
			activityBaseInfo.ActivityContent = proto.String(baseInfoRet.GetSingleValue("ActivityContent").(string))
			activityBaseInfo.ActivityStatus = proto.Int32(baseInfoRet.GetSingleValueInt32("Status"))
			activityBaseInfo.ActivityBeginTime = proto.String(baseInfoRet.GetSingleValue("BeginTime").(string))
			activityBaseInfo.ActivityEndTime = proto.String(baseInfoRet.GetSingleValue("EndTime").(string))
			activityBaseInfo.ActivityTipContent = proto.String(baseInfoRet.GetSingleValue("ActivityTipContent").(string))
			activityBaseInfo.Extend = proto.String(baseInfoRet.GetSingleValue("Extend").(string))

			if ret.GetRetTableCount() >= 4 {
				configInfoRet := ret[2]

				configInfoRet.ForEachRow(func(r db.RecordRow) {
					activityConfig := new(netproto.ActivityConfig)
					activityConfig.ParamName = proto.String(r[1].(string))
					activityConfig.ParamValue = proto.String(r[2].(string))
					activityConfig.ParamDesc = proto.String(r[3].(string))

					activityBaseInfo.ActivityConfig = append(activityBaseInfo.ActivityConfig, activityConfig)
				})
			}

			activityInfoRet.ActivityBaseInfo = activityBaseInfo
		}
	}
	return activityInfoRet
}
