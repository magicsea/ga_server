package login

import (
	"GAServer/cluster"
	"GAServer/config"
	"GAServer/messages"
	"GAServer/service"
	"fmt"
	"log"
	"net/http"
	"strings"

	"strconv"

	"github.com/gogo/protobuf/proto"
)

type LoginService struct {
	service.ServiceData
}

//Service 获取服务对象
func Service() service.IService {
	return new(LoginService)
}

func Type() string {
	return "login"
}

//以下为接口函数
func (s *LoginService) OnReceive(context service.Context) {
	fmt.Println("center.OnReceive:", context.Message())
}
func (s *LoginService) OnInit() {

}
func (s *LoginService) OnStart(as *service.ActorService) {
	//as.RegisterMsg(reflect.TypeOf(&messages.UserLogin{}), s.OnUserLogin) //注册登录

	//开启rpc,任意端口
	//remote.Start("127.0.0.1:0")
	//cluster.Start(&cluster.ClusterConfig{"127.0.0.1:8090", "127.0.0.1:8091"})

	go func() {
		//开启http服务
		http.HandleFunc("/login", login)

		httpAddr := config.GetServiceConfigString(s.Name, "httpAddr")
		log.Println("login listen http:", s.Name, "  ", httpAddr)
		http.ListenAndServe(httpAddr, nil)
	}()

}

func login(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Form["a"] == nil || req.Form["p"] == nil {
		log.Println("a,p is empty:", req.Form)
		return
	}
	acc := req.Form["a"][0]
	//_ := req.Form["p"][0]

	//验证 here...
	log.Println("login account:", acc)
	strs := strings.Split(acc, "_")
	id, _ := strconv.Atoi(strs[1])
	resp, err := cluster.GetServicePID("session").Ask(&messages.UserLogin{acc, uint64(id)})
	if err == nil {
		var s, _ = proto.Marshal(resp.(*messages.UserLoginResult))
		//var s, _ = json.Marshal(resp)
		w.Write(s)
		log.Println("login ok:", resp)
	} else {
		result := messages.UserLoginResult{Result: messages.Error}
		var s, _ = proto.Marshal(&result)
		w.Write(s)
		log.Println("login error:", acc, err)
	}
}
