# ga_server

基于protoactor框架的actor游戏服务器。

## 设计动机
- 一套面向actor的分布式游戏服务器
- 实现可伸缩设计，缩可以放在一个进程，伸可以扩展多台机器均衡负载

## 目录结构
- GAServer：基本库代码，主要是gate模块和service类型的封装
- Robot:机器人测试代码，robotMachine是压力测试，robotTest是简单功能测试
- Server：里面是各种服务的实现


## 登录流程
红色为单点，其他都是多点。
login使用http协议和客户端沟通，其他请求通过gate转发
![image](http://www.processon.com/chart_image/58f6d36be4b02e95ec64c368.png)


## TODO
- battleserver实现
- gate加密
- ...