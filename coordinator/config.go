package coordinator

type Config struct {
	RaftNodeId  uint64
	DataDir     string
	Port        string
	JoinNodes   []string
	IsJoin      bool
	TlsCertFile string
	TlsKeyFile  string
}

func NewConfig() *Config {
	return &Config{
		Port:    "50051",
		DataDir: "/vemoo_pid",
		IsJoin:  false,
	}
}
