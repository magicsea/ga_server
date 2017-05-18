package session

import (
	"GAServer/messages"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type PlayerSession struct {
	userInfo      *messages.UserBaseInfo
	gatePid       *actor.PID //gate服地址
	agentPid      *actor.PID //agent对象地址
	gamePlayerPid *actor.PID //player对象地址

	key string //动态生成密码
}

//踢下线
func (p *PlayerSession) Kick() {
	//msg := &messages.Kick{p.userInfo.Uid}
	//p.gatePid.Tell(msg)
	//p.gamePlayerPid.Tell(msg)
}
