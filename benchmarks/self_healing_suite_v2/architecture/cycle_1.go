package architecture

type ServiceAlpha1 struct {
	Peer *ServiceBeta1
}

type ServiceBeta1 struct {
	Peer *ServiceAlpha1
}
