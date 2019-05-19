package module

//config 存取加载的配置
type Config struct {
	LogLevel  string        `json:"log_level"`
	LogPath   string        `json:"log_path"`
	ChanSize  int           `json:"chan_size"`
	KafkaAddr string        `json:"kafka_addr"`
	Collect   []CollectConf `json:"collect"`
}
//CollectConf 日志收集配置
type CollectConf struct {
	LogPath 	string `json:"log_path"`
	Topic 		string `json:"topic"`
}


