package config

type ServiceNodeConfig struct {
	ServiceName string                 `json:"serviceName"` //服务名 (gameserver)
	ServiceType string                 `json:"serviceType"` //服务类型 (gameserver1)
	RemoteAddr  string                 `json:"remoteAddr"`  //远程地址 (127.0.0.1:80 or "",空字符表示本地启动)
	Conf        map[string]interface{} `json:"conf"`        //单体配置
}
type LogConfig struct {
	LogLevel string `json:"level"`
	LogPath  string `json:"path"`
	LogFlag  int    `json:"flag"`
}

type ServiceConfig struct {
	Services    map[string]*ServiceNodeConfig `json:"local"`
	RemoteAddrs map[string]string             `json:"remote"`
	LogConf     *LogConfig                    `json:"log"`
}

var globleConfig ServiceConfig

func SetGlobleConfig(conf *ServiceConfig) {
	globleConfig = *conf
}

func GetGlobleConfig() *ServiceConfig {
	return &globleConfig
}

func GetServiceConfig(ser string, key string) interface{} {
	return globleConfig.Services[ser].Conf[key]
}

func GetServiceConfigString(ser string, key string) string {
	return GetServiceConfig(ser, key).(string)
}

func GetServiceConfigInt(ser string, key string) int {
	f := GetServiceConfig(ser, key).(float64)
	return int(f)
}

//GetServiceAddress 获取服务地址，先去remote找，没有就到本地找
func GetServiceAddress(serviceName string) string {
	if globleConfig.RemoteAddrs != nil {
		return globleConfig.RemoteAddrs[serviceName]
	} else {
		return globleConfig.Services[serviceName].RemoteAddr
	}
}
