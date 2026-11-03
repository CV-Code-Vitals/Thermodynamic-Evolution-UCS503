package architecture

type ServiceAlpha4 struct {
	Peer *ServiceBeta4
}

type ServiceBeta4 struct {
	Peer *ServiceAlpha4
}
