#停止维护。移步新的项目 https://github.com/magicsea/ganet

# ga_server

基于protoactor框架的actor游戏服务器。

## 设计动机
- 一套面向actor的分布式游戏服务器
- 实现可伸缩设计，缩可以放在一个进程，伸可以扩展多台机器均衡负载

## 目录结构
- cofig：游戏协议，gameproto存放c2s/s2c协议，msgs存放s2s协议。打包将生成到src的gameproto目录
- src/GAServer：基本库代码，主要是gate模块和service类型的封装
- src/Robot:机器人测试代码，robotMachine是压力测试，robotTest是简单功能测试
> 目前数据：robotCount= 500 time=ms 5457 all_qps= 91625.44
- src/Server：里面是各种服务的实现。服务器的具体实现目录
## 启动
- win编译出server执行文件
- 可以直接执行server，默认读取config.json配置，所有服务在一个进程
- 或者执行StartMultiServer.bat，启多个进程服务器,服务分开部署

## 登录流程
红色为单点，其他都是多点。
login使用http协议和客户端沟通，其他请求通过gate转发
![image](http://on-img.com/chart_image/58f6d36be4b02e95ec64c368.png)

## 依赖
主要依赖protoactor里的库，具体参考protoactor的readme。简单的直接使用LiteIde执行go get一下就自动下载。google的几个库需要科学上网，没条件的下载我网盘里的[ google库](http://pan.baidu.com/s/1qYjUHJY)
## TODO
- battleserver实现
- gate加密
- ...
## QQ群：285728047
