package gateway

type GateWay struct {
	Host        string
	Port        string
	ServerAddrs []string

	Balancer BalancerType
	Infra    Infra

	// when k8s
	ServiceName string
}
