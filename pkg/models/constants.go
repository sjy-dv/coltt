package models

type BalancerType int

const (
	LB BalancerType = iota
	SLB
)

type Infra int

const (
	Local Infra = iota
	K8S
)
