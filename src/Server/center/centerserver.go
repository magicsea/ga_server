package center

import (
	"GAServer/log"
	"GAServer/messages"
	"GAServer/service"

	"reflect"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type CenterService struct {
	service.ServiceData
	serviceGroups map[string]*ServiceGroup //所有服务 map[type]group
	serviceAll    map[string]*ServiceNode  //索引服务路径引用 map[addr+id]group
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
	log.Println("center.OnReceive:", context.Message())

}

func (s *CenterService) OnInit() {
	s.serviceGroups = make(map[string]*ServiceGroup)
	s.serviceAll = make(map[string]*ServiceNode)
}

func (s *CenterService) OnStart(as *service.ActorService) {
	//as.RegisterMsg(reflect.TypeOf(&messages.RemoveService{}), s.OnRemoveService)  //解注册服务器
	as.RegisterMsg(reflect.TypeOf(&messages.AddService{}), s.OnAddService)          //注册服务器
	as.RegisterMsg(reflect.TypeOf(&actor.Terminated{}), s.OnChildServiceTerminated) //被动断开服务器
	as.RegisterMsg(reflect.TypeOf(&messages.UploadService{}), s.OnUpdateService)    //更新服务器
	as.RegisterMsg(reflect.TypeOf(&messages.ApplyService{}), s.OnApplyService)      //获取一个服务器
	as.RegisterMsg(reflect.TypeOf(&messages.GetTypeServices{}), s.GetTypeServices)  //获取一类服务器
}

//注册服务器
func (s *CenterService) OnAddService(context service.Context) {
	log.Println("center.OnAddService:", context.Message())
	msg := context.Message().(*messages.AddService)
	var group *ServiceGroup
	if g, ok := s.serviceGroups[msg.ServiceType]; !ok {
		group = new(ServiceGroup)
		group.services = make(map[string]*ServiceNode)
		s.serviceGroups[msg.ServiceType] = group
		log.Println("new service Type:", msg.ServiceType)
	} else {
		group = g
	}

	var node = &ServiceNode{pid: msg.Pid,
		serviceType: msg.ServiceType,
		serviceName: msg.ServiceName,
		load:        0,
		state:       messages.ServiceStateFree,
		values:      msg.Values}

	s.serviceAll[node.pid.String()] = node //加入索引
	group.AddService(node)                 //加入group
	context.Watch(node.pid)                //监控

	context.Tell(context.Sender(), &messages.SendOK{})
	log.Println("center.OnAddService  OK,", msg.ServiceName)
}

//解注册服务器
func (s *CenterService) __OnRemoveService(context service.Context) {
	log.Println("center.OnRemoveService:", context.Message())
	msg := context.Message().(*messages.RemoveService)

	var group *ServiceGroup
	if g, ok := s.serviceGroups[msg.ServiceType]; !ok {
		group = g
		log.Error("no found service Type:%v", msg.ServiceType)
		return
	}

	group.RemoveService(msg.ServiceName)
}

//被动断开服务器
func (s *CenterService) OnChildServiceTerminated(context service.Context) {
	log.Println("center.OnChildServiceTerminated:", context.Message())

	msg := context.Message().(*actor.Terminated)
	//context.Unwatch(msg.Who)//需要主动unwatch???
	path := msg.Who.String()

	sv := s.serviceAll[path]
	if sv == nil {
		log.Error("OnChildServiceTerminated,no found service:%v", path)
		return
	}
	delete(s.serviceAll, path) //移除索引

	var group *ServiceGroup
	if g, ok := s.serviceGroups[sv.serviceType]; !ok {
		group = g
		log.Error("OnChildServiceTerminated,no found service Type:%v", sv.serviceType)
		return
	}

	group.RemoveService(sv.serviceName) //移除group
}

//更新服务器
func (s *CenterService) OnUpdateService(context service.Context) {
	log.Println("center.OnUpdateService:", context.Message())
	msg := context.Message().(*messages.UploadService)

	if sv, ok := s.serviceAll[msg.ServiceName]; ok {
		sv.load = msg.Load
		sv.state = msg.State
	}
}

//获取一个服务器
func (s *CenterService) OnApplyService(context service.Context) {
	log.Println("center.OnApplyService:", context.Message())
	msg := context.Message().(*messages.ApplyService)
	var group *ServiceGroup
	if g, ok := s.serviceGroups[msg.ServiceType]; !ok {
		log.Error("OnApplyService,no found service Type:%v", msg.ServiceType)
		context.Sender().Tell(&messages.ApplyServiceResult{Result: messages.Error})
		return
	} else {
		group = g
	}
	sv := group.GetBestService()
	resultMsg := &messages.ApplyServiceResult{ServiceType: msg.ServiceType}
	if sv != nil {
		resultMsg.Result = messages.OK
		resultMsg.Pid = sv.pid
		resultMsg.ServiceName = sv.serviceName
		resultMsg.Values = sv.values
	} else {
		resultMsg.Result = messages.Fail
		log.Error("OnApplyService have no service:%v", msg.ServiceType)
	}
	context.Sender().Tell(resultMsg)
}

//获取一类服务器
func (s *CenterService) GetTypeServices(context service.Context) {
	log.Println("center.GetTypeServices:", context.Message())
	msg := context.Message().(*messages.GetTypeServices)
	var group *ServiceGroup
	if g, ok := s.serviceGroups[msg.ServiceType]; !ok {
		log.Error("GetTypeServices,no found service Type:%v", msg.ServiceType)
		context.Sender().Tell(&messages.GetTypeServicesResult{})
		return
	} else {
		group = g
	}

	resultMsg := &messages.GetTypeServicesResult{}
	for _, v := range group.services {
		resultMsg.Pids = append(resultMsg.Pids, v.pid)
	}
	context.Sender().Tell(resultMsg)
}
