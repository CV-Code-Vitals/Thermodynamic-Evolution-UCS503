package architecture

type OrderModelV3 struct {
	ID       string
	Partner  *CustomerModelV3
	Metadata map[string]string
}

type CustomerModelV3 struct {
	ID       string
	Parent   *OrderModelV3
	Capacity int
}
