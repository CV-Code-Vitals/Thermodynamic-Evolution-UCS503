package architecture

type ServiceAlpha2 struct {
	Peer *ServiceBeta2
}

type ServiceBeta2 struct {
	Peer *ServiceAlpha2
}
