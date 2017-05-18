package center

import (
	"GAServer/service"
)

type CenterService struct {
	service.ServiceData
}

//Service 获取服务对象
func Service() service.IService {
	return new(CenterService)
}

func Type() string {
	return "center"
}

//以下为接口函数
func (s *CenterService) OnReceive(context service.Context) {

}
func (s *CenterService) OnInit() {

}
func (s *CenterService) OnStart() {

}
