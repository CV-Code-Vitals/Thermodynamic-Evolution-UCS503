package architecture

type SessionContextV3 struct {
	ID       string
	Partner  *DeviceRegistryV3
	Metadata map[string]string
}

type DeviceRegistryV3 struct {
	ID       string
	Parent   *SessionContextV3
	Capacity int
}
