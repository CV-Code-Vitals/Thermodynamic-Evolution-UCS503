package architecture

type ClusterCoordinatorV3 struct {
	ID       string
	Partner  *MemberPeerV3
	Metadata map[string]string
}

type MemberPeerV3 struct {
	ID       string
	Parent   *ClusterCoordinatorV3
	Capacity int
}
