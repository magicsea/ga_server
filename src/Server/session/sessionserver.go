package session

import (
	"GAServer/cluster"
	"GAServer/log"
	"GAServer/messages"
	"GAServer/service"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type SessionService struct {
	service.ServiceData
	sessionMgr *SessionManager
}

//Service 获取服务对象
func Service() service.IService {
	return new(SessionService)
}

func Type() string {
	return "session"
}

//以下为接口函数
func (s *SessionService) OnReceive(context service.Context) {
	fmt.Println("session.OnReceive:", context.Message())
}

func (s *SessionService) OnInit() {
	s.sessionMgr = NewSessionManager()
}

func (s *SessionService) OnStart(as *service.ActorService) {
	as.RegisterMsg(reflect.TypeOf(&messages.UserLogin{}), s.OnUserLogin)                     //注册登录
	as.RegisterMsg(reflect.TypeOf(&messages.ServerCheckLogin{}), s.OnUserCheckLogin)         //二次验证
	as.RegisterMsg(reflect.TypeOf(&messages.GetSessionInfo{}), s.GetSessionInfo)             //查询玩家信息
	as.RegisterMsg(reflect.TypeOf(&messages.GetSessionInfoByName{}), s.GetSessionInfoByName) //查询玩家信息通过名字
}

//查询玩家信息
func (s *SessionService) GetSessionInfo(context service.Context) {
	fmt.Println("SessionService.GetSessionInfo:", context.Message())
	msg := context.Message().(*messages.GetSessionInfo)
	ss := s.sessionMgr.GetSession(msg.Uid)
	if ss != nil {
		context.Tell(context.Sender(), &messages.GetSessionInfoResult{Result: messages.OK, UserInfo: ss.userInfo, AgentPID: ss.agentPid})
	} else {
		context.Tell(context.Sender(), &messages.GetSessionInfoResult{Result: messages.Fail})
	}
}

//查询玩家信息 by name
func (s *SessionService) GetSessionInfoByName(context service.Context) {
	fmt.Println("SessionService.GetSessionInfoByName:", context.Message())
	msg := context.Message().(*messages.GetSessionInfoByName)
	ss := s.sessionMgr.GetSessionByName(msg.Name)
	if ss != nil {
		context.Tell(context.Sender(), &messages.GetSessionInfoResult{Result: messages.OK, UserInfo: ss.userInfo, AgentPID: ss.agentPid})
	} else {
		context.Tell(context.Sender(), &messages.GetSessionInfoResult{Result: messages.Fail})
	}
	fmt.Println("GetSessionInfoByName end")
}

//玩家登陆
func (s *SessionService) OnUserLogin(context service.Context) {
	fmt.Println("SessionService.OnUserLogin:", context.Message())
	msg := context.Message().(*messages.UserLogin)

	//踢掉老玩家
	oldSession := s.sessionMgr.GetSession(msg.Uid)
	if oldSession != nil {
		oldSession.Kick()
		s.sessionMgr.RemoveSession(msg.Uid)
	}

	//请求gate
	result, err := cluster.GetServicePID("center").Ask(&messages.ApplyService{"gate"})
	if err != nil {
		log.Error("get gate server,%v", err)
		context.Tell(context.Sender(), &messages.UserLoginResult{Result: messages.Error})
		return
	}

	sr := result.(*messages.ApplyServiceResult)
	if sr.Result != messages.OK {
		context.Tell(context.Sender(), &messages.UserLoginResult{Result: sr.Result})
		return
	}

	//加入数据
	uInfo := &messages.UserBaseInfo{msg.Account, "玩家" + strconv.Itoa(int(msg.Uid)), msg.Uid}
	ss := &PlayerSession{userInfo: uInfo, gatePid: sr.Pid, key: "1111"}
	s.sessionMgr.AddSession(ss)

	gateAddr := GetServiceValue("TcpAddr", sr.Values)
	context.Tell(context.Sender(), &messages.UserLoginResult{msg.Uid, gateAddr, ss.key, messages.OK})
}

func GetServiceValue(key string, values []*messages.ServiceValue) string {
	for _, v := range values {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

//玩家验证
func (s *SessionService) OnUserCheckLogin(context service.Context) {
	fmt.Println("SessionService.OnUserCheckLogin:", context.Message())
	msg := context.Message().(*messages.ServerCheckLogin)

	oldSession := s.sessionMgr.GetSession(msg.Uid)
	//人不在
	if oldSession == nil || oldSession.key != msg.Key {
		log.Error("OnUserCheckLogin,no found player,id=%v", msg.Uid)
		context.Tell(context.Sender(), &messages.CheckLoginResult{Result: messages.Fail})
		return
	}
	//密码错
	if oldSession.key != msg.Key {
		log.Error("OnUserCheckLogin,key error,id=%v,key=%v:%v", msg.Uid, oldSession.key, msg.Key)
		context.Tell(context.Sender(), &messages.CheckLoginResult{Result: messages.KeyError})
		return
	}
	//请求gameserver
	result, err := cluster.GetServicePID("center").Ask(&messages.ApplyService{"game"})
	if err != nil {
		log.Error("get gameserver error:%v", err)
		context.Tell(context.Sender(), &messages.CheckLoginResult{Result: messages.Error})
		return
	}

	sr := result.(*messages.ApplyServiceResult)
	if sr.Result != messages.OK {
		context.Tell(context.Sender(), &messages.CheckLoginResult{Result: sr.Result})
		return
	}
	//安装
	//todo:这里可能比较耗时，以后改成异步
	gsPid := sr.Pid
	r, err := context.RequestFuture(gsPid, &messages.CreatePlayer{msg.Uid, msg.AgentPID}, time.Second*3).Result()
	if err != nil {
		log.Error("OnUserCheckLogin,create player error,id=%v,%v", msg.Uid, err)
		context.Tell(context.Sender(), &messages.CheckLoginResult{Result: messages.Error})
		return
	}
	cresult, _ := r.(*messages.CreatePlayerResult)
	if cresult.Result != messages.OK {
		log.Error("OnUserCheckLogin,create player fail,id=%v,%v", msg.Uid, cresult.Result)
		context.Tell(context.Sender(), &messages.CheckLoginResult{Result: messages.Error})
		return
	}
	//完成
	oldSession.agentPid = msg.AgentPID
	oldSession.gamePlayerPid = cresult.PlayerPID
	gsValue := messages.UserBindServer{messages.GameServer, cresult.GetPlayerPID()}
	context.Tell(context.Sender(), &messages.CheckLoginResult{
		Result:      messages.OK,
		BaseInfo:    oldSession.userInfo,
		BindServers: []*messages.UserBindServer{&gsValue}})
	log.Println("SessionService.OnUserCheckLogin ok:", msg.Uid)
}
