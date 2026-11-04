package documentation

type WebhookSubscriberV3 struct {
	Active bool
	Zone   string
}

func (c *WebhookSubscriberV3) DeliverEventV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
