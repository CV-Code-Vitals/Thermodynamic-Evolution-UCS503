package documentation

type PayloadProtoEncoderV3 struct {
	Active bool
	Zone   string
}

func (c *PayloadProtoEncoderV3) EncodeFrameV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
