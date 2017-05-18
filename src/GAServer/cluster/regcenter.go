package cluster

import (
	"GAServer/log"
	"GAServer/messages"
	"GAServer/service"
	_ "encoding/json"
)

//注册到center
func RegServerToCenter(s *service.ServiceData, values []*messages.ServiceValue) bool {
	log.Info("%v reg to center...", s.Name)

	msg := messages.AddService{
		ServiceName: s.Name,
		ServiceType: s.TypeName,
		Pid:         s.GetPID(),
		Values:      values}
	_, err := GetServicePID("center").Ask(&msg)
	if err != nil {
		log.Error("%v reg to center fail,%v", s.Name, err)
		//if err.Error() == "timeout" {
		//DisconnectService("center")
		//}
		//重连
		return false
	}
	log.Info("%v reg to center OK!", s.Name)
	return true
}

func RegServerWork(s *service.ServiceData, values []*messages.ServiceValue) {
	go func() {
		for {
			if RegServerToCenter(s, values) {
				break
			}
		}
	}()

}
