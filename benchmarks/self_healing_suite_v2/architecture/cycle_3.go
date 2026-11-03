package architecture

type ServiceAlpha3 struct {
	Peer *ServiceBeta3
}

type ServiceBeta3 struct {
	Peer *ServiceAlpha3
}
