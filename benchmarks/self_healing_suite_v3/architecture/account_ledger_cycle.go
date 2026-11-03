package architecture

type AccountNodeV3 struct {
	ID       string
	Partner  *LedgerNodeV3
	Metadata map[string]string
}

type LedgerNodeV3 struct {
	ID       string
	Parent   *AccountNodeV3
	Capacity int
}
