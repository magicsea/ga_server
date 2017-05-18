package module

import (
	"runtime"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type Module interface {
	OnInit()
	OnDestroy()
	Run(closeSig chan bool)
}

type module struct {
	mi       Module
	closeSig chan bool
	wg       sync.WaitGroup
}

type IActorObject interface {
	//通知一条消息，立刻返回
	Tell(args interface{})
	//通知一条消息，阻塞等待结果
	Ask(args interface{}) interface{},error
	//通知一条信息，立刻返回，结果会放回recv通道
	AskCB(args interface{})
}


type defaultActor interface {
	pid *actor.PID
}


