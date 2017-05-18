package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	_ "github.com/AsynkronIT/protoactor-go/remote"
)

type IServiceData interface {
	Init(addr string, name string, typename string)
	GetName() string
	GetType() string
	GetAddress() string
	SetPID(pid *actor.PID)
	GetPID() *actor.PID
}

type ServiceData struct {
	Address  string
	Name     string
	TypeName string
	Pid      *actor.PID
}

func (s *ServiceData) Init(addr string, name string, typename string) {
	s.Address = addr
	s.Name = name
	s.TypeName = typename
}

func (s *ServiceData) GetType() string {
	return s.TypeName
}
func (s *ServiceData) GetName() string {
	return s.Name
}
func (s *ServiceData) GetAddress() string {
	return s.Address
}
func (s *ServiceData) SetPID(pid *actor.PID) {
	s.Pid = pid
}
func (s *ServiceData) GetPID() *actor.PID {
	return s.Pid
}
func (s *ServiceData) OnDestory() {

}
