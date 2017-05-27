package login

import (
	"GAServer/config"
	"GAServer/log"
	"GAServer/service"
	"Server/cluster"
	"fmt"
	"gameproto/msgs"
	"net/http"

	_ "Server/db"
	"gameproto"

	"strconv"
	"strings"
	_ "time"

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
		log.Error("a,p is empty:", req.Form)
		return
	}
	//账号
	acc := ""
	if al, ok := req.Form["a"]; ok {
		acc = al[0]
	}
	//pwd := ""
	//if pl, ok := req.Form["i"]; ok {
	//	pwd = pl[0]
	//}

	//验证 here...
	log.Println("login account:", acc)
	strs := strings.Split(acc, "_")
	id, _ := strconv.Atoi(strs[1])
	//now := time.Now().Unix()
	//user := &db.User{PlatformId: acc, LastLoginTime: now}
	//var id uint64
	//norow, e := db.GetGameDB().Read(user, "PlatformId")
	//if e != nil && !norow {
	//	loginBackError(w, e)
	//	return
	//}

	//user.LastLoginTime = now
	//if norow {
	//新用户ErrNoRows
	//	user.RegisterTime = now
	//	db.GetGameDB().Insert(user)
	//} else {
	//老用户
	//	db.GetGameDB().Update(user, "LastLoginTime")
	//}

	//id = user.Id

	resp, err := cluster.GetServicePID("session").Ask(&msgs.UserLogin{acc, uint64(id)})
	if err == nil {
		var s, _ = resp.(*gameproto.UserLoginResult).Marshal()
		//var s, _ = json.Marshal(resp)
		w.Write(s)
		log.Info("login ok:msg=%v", resp)
	} else {
		loginBackError(w, err)
		log.Println("login error:", acc, err)
	}
}

func loginBackError(w http.ResponseWriter, e error) {
	log.Error("create user db :%v", e)
	d, _ := proto.Marshal(&gameproto.UserLoginResult{Result: int32(msgs.Error)})
	w.Write(d)
}
