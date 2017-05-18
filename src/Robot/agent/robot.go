package agent

import (
	"GAServer/messages"
	"GAServer/network"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
)

type Robot struct {
	account string
	pwd     string

	gateAddr string
	uid      uint64
	key      string

	client      *network.TCPClient
	agent       *Agent
	wg          sync.WaitGroup
	taskCounter int
	taskCount   int
	actionTime  time.Duration
	result      chan string
}

func NewRobot(account, pwd string, actionTime time.Duration) *Robot {
	return &Robot{account: account, pwd: pwd, actionTime: actionTime, result: make(chan string, 1)}
}

func (robot *Robot) Start(taskCount int) string {
	robot.taskCount = taskCount
	if !robot.Login() {
		return "Login fail"
	}
	//robot.wg.Add(1)
	robot.ConnectGate()
	//robot.wg.Wait()
	r := <-robot.result
	return r
}

func (robot *Robot) Login() bool {
	fmt.Println("login...")

	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:8080/login?a=%s&p=111", robot.account))
	if err != nil {
		log.Println("login http.get fail:", err)
		return false
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	result := messages.UserLoginResult{}
	umErr := proto.Unmarshal(body, &result)
	fmt.Println("err:", umErr, "  result:", result)
	robot.uid = result.Uid
	robot.key = result.Key
	robot.gateAddr = result.GateAddr
	return result.GetResult() == messages.OK
}

func (robot *Robot) newAgent(conn *network.TCPConn) network.Agent {
	robot.agent = new(Agent)
	robot.agent.conn = conn
	robot.agent.msgHandle = robot.OnMsgRecv
	robot.agent.errorFun = robot.OnErr
	robot.agent.closeFun = robot.OnDisconnected
	robot.OnConnected()
	return robot.agent
}

func (robot *Robot) ConnectGate() {
	fmt.Println("ConnectGate...")
	robot.client = new(network.TCPClient)
	robot.client.Addr = robot.gateAddr
	robot.client.NewAgent = robot.newAgent
	//robot.client.AutoReconnect = true
	robot.client.Start()

}

func (robot *Robot) OnConnected() {
	fmt.Println("OnConnected...")
	robot.SendMsg(messages.Login, 0, &messages.CheckLogin{robot.uid, robot.key})
}

func (robot *Robot) OnDisconnected() {
	fmt.Println("OnDisconnected...")
	robot.result <- "close"
}

func (robot *Robot) EnterGame() {
	fmt.Println("EnterGame...")

	robot.SendMsg(messages.Shop, byte(messages.C2S_ShopBuy), &messages.C2S_ShopBuyMsg{1})
	//robot.SendMsg(messages.Chat, byte(messages.C2S_PrivateChat), &messages.C2S_PrivateChatMsg{"玩家11", "hello"})
	//robot.SendMsg(messages.Chat, byte(messages.C2S_WorldChat), &messages.C2S_WorldChatMsg{"world"})
}

func (robot *Robot) OnErr(err string) {
	robot.result <- err
	//robot.Finish(err)
	//robot.SendMsg(messages.Shop, byte(messages.C2S_ShopBuy), &messages.C2S_ShopBuyMsg{1})
}
func (robot *Robot) OnMsgRecv(channel byte, msgId byte, data []byte) {
	c := messages.ChannelType(channel)
	//fmt.Println("OnMsgRecv:", c, " msg:", msgId, " data:", len(data))
	if c == messages.Login {
		msg := messages.CheckLoginResult{}
		proto.Unmarshal(data, &msg)
		fmt.Println("login result:", msg)
		if msg.Result == messages.OK {
			robot.EnterGame()
		}
	} else if c == messages.Shop {
		tmsgId := messages.ShopMsgType(msgId)
		switch tmsgId {
		case messages.S2C_ShopBuy:
			msg := messages.S2C_ShopBuyMsg{}
			proto.Unmarshal(data, &msg)
			//fmt.Println("shop result:", msg)
			if robot.taskCounter > robot.taskCount {
				robot.Finish("OK")
			}
			if robot.actionTime > 0 {
				time.Sleep(robot.actionTime)
			}

			robot.taskCounter++
			robot.SendMsg(messages.Shop, byte(messages.C2S_ShopBuy), &messages.C2S_ShopBuyMsg{1})
		}
	} else if c == messages.Chat {
		tmsgId := messages.ChatMsgType(msgId)
		switch tmsgId {
		case messages.S2C_PrivateChat:
			msg := messages.S2C_PrivateChatMsg{}
			proto.Unmarshal(data, &msg)
			fmt.Println("chat back result:", msg)
		case messages.S2C_PrivateOtherChat:
			msg := messages.S2C_PrivateOtherChatMsg{}
			proto.Unmarshal(data, &msg)
			fmt.Println("otherchat:", msg)
		case messages.S2C_WorldChat:
			msg := messages.S2C_WorldChatMsg{}
			proto.Unmarshal(data, &msg)
			fmt.Println("worldchat :", msg)
		}
	}
}

func (robot *Robot) SendMsg(channel messages.ChannelType, msgId byte, pb proto.Message) {
	data, err := proto.Marshal(pb)
	if err != nil {
		fmt.Println("###EncodeMsg error:", err)
		return
	}
	robot.agent.WriteMsg(byte(channel), msgId, data)
}

func (robot *Robot) Finish(result string) {
	//robot.result = result
	robot.client.AutoReconnect = false
	robot.agent.Close()
	//robot.wg.Done()
	robot.result <- result
}
