/*
**服务类型，可作为独立进程或线程使用
**
 */
package service

import (
	"fmt"

	"reflect"

	"GAServer/log"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/remote"
)

type Context actor.Context

type IService interface {
	IServiceData
	//Receive(context actor.Context)
	OnReceive(context Context)
	OnInit()
	OnStart(as *ActorService)
	OnDestory()
}

//interface
//func (s *BaseServer) OnReceive(context Context)            {}
//func (s *BaseServer) OnInit()                              {}
//func (s *BaseServer) OnStart()                             {}

type MessageFunc func(context Context)

//服务的代理
type ActorService struct {
	serviceIns IService
	rounter    map[reflect.Type]MessageFunc
}

func (s *ActorService) Receive(context actor.Context) {
	//switch msg := context.Message().(type) {
	//case *hello:
	//	fmt.Printf("Hello %v\n", msg.Who)
	//}
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Stopped:
		fmt.Println("Stopped, actor and its children are stopped")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	default:
		fmt.Println("recv defalult:", msg)
		s.serviceIns.OnReceive(context.(Context))
		fun := s.rounter[reflect.TypeOf(msg)]
		if fun != nil {
			fun(context.(Context))
		}
	}
}

func (s *ActorService) RegisterMsg(t reflect.Type, f MessageFunc) {
	s.rounter[t] = f
}

func StartService(s IService) {
	ac := &ActorService{s, make(map[reflect.Type]MessageFunc)}
	props := actor.FromInstance(ac)
	if s.GetAddress() != "" {
		remote.Start(s.GetAddress())
	}
	pid, err := actor.SpawnNamed(props, s.GetName())
	if err == nil {
		s.SetPID(pid)
		s.OnStart(ac)
	} else {
		log.Error("#############actor.SpawnNamed error:%v", err)
	}

}

func DestoryService(s *ActorService) {
	s.serviceIns.OnDestory()
}
