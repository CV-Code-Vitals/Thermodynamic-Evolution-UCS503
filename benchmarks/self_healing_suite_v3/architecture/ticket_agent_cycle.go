package architecture

type SupportTicketV3 struct {
	ID       string
	Partner  *AssignedAgentV3
	Metadata map[string]string
}

type AssignedAgentV3 struct {
	ID       string
	Parent   *SupportTicketV3
	Capacity int
}
