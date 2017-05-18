package app

import (
	"GAServer/cluster"
	. "GAServer/config"
	"GAServer/log"
	"GAServer/service"
	"os"
	"os/signal"
)

type MakeServiceFunc func() service.IService

var (
	serviceTypeMap map[string]MakeServiceFunc
	services       []service.IService
)

func init() {
	serviceTypeMap = make(map[string]MakeServiceFunc)
}

func RegisterService(serviceType string, f MakeServiceFunc) {
	serviceTypeMap[serviceType] = f
}

func Run(conf *ServiceConfig) {
	SetGlobleConfig(conf)
	//init log
	if conf.LogConf.LogLevel != "" {
		err := log.NewLogGroup(conf.LogConf.LogLevel, conf.LogConf.LogPath, true, conf.LogConf.LogFlag)
		if err != nil {
			panic(err)
		}
		//log.Export(logger)
		defer log.Close()
	}
	log.Info("log started.")
	cluster.InitCluster()
	//生成服务对象
	for _, sc := range conf.Services {
		makefunc := serviceTypeMap[sc.ServiceType]
		if makefunc != nil {
			ser := makefunc()
			log.Info("生成服务:", sc.ServiceName)
			ser.Init(sc.RemoteAddr, sc.ServiceName, sc.ServiceType)
			services = append(services, ser)
		} else {
			log.Fatal("未注册的服务类型:", sc)
		}
	}

	//init
	for _, ser := range services {
		log.Println("init服务:", ser.GetName())
		ser.OnInit()
	}

	//start
	for _, ser := range services {
		log.Println("start服务:", ser.GetName())
		service.StartService(ser)
	}

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Println("closing down (signal: %v)", sig)
	OnDestory()
}

func OnDestory() {
	for _, ser := range services {
		log.Println("destory服务:", ser.GetName())
		ser.OnDestory()
	}

}
