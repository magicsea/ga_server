package game

import (
	"GAServer/cluster"
	"GAServer/log"
	"GAServer/messages"
	"GAServer/service"
	"reflect"
	"strconv"

	_ "github.com/AsynkronIT/protoactor-go/actor"
)

type GameService struct {
	service.ServiceData
}

//Service 获取服务对象
func Service() service.IService {
	return new(GameService)
}

func Type() string {
	return "game"
}

//以下为接口函数
func (s *GameService) OnReceive(context service.Context) {
	log.Println("game.OnReceive:", context.Message())
}

func (s *GameService) OnInit() {

}

func (s *GameService) OnStart(as *service.ActorService) {
	as.RegisterMsg(reflect.TypeOf(&messages.CreatePlayer{}), s.OnCreatePlayer) //登录

	//注册到center
	cluster.RegServerWork(&s.ServiceData, nil)
}

func (s *GameService) OnCreatePlayer(context service.Context) {
	log.Info("GameService.OnCreatePlayer:", context.Message())
	msg := context.Message().(*messages.CreatePlayer)

	//todo:从db里load基本数据（比如player表）...
	baseInfo := &messages.UserBaseInfo{Uid: msg.Uid, Name: "玩家" + strconv.Itoa(int(msg.Uid))}
	//创建玩家对象actor(异步)
	player, err := NewPlayer(baseInfo, msg.AgentPID, context)
	errCode := messages.OK
	if err != nil {
		log.Error("NewPlayer error:%v,%v", msg.Uid, err)
		errCode = messages.Error
	}
	result := &messages.CreatePlayerResult{errCode, baseInfo, player.selfPID}
	context.Tell(context.Sender(), result)

	log.Println("GameService.OnCreatePlayer  ok:", baseInfo.Uid, " playerpid=", player.selfPID)
}
