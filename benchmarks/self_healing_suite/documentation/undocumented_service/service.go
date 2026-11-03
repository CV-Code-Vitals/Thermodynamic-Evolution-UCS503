package undocumented_service

// Ground-truth defect: FIX-DOC-01 (Documentation: Missing required documentation on exported API)

type PaymentGatewayClient struct {
	Endpoint string
	APIKey   string
}

func NewPaymentGatewayClient(endpoint, apiKey string) *PaymentGatewayClient {
	return &PaymentGatewayClient{
		Endpoint: endpoint,
		APIKey:   apiKey,
	}
}

func (c *PaymentGatewayClient) Charge(cents int64, currency string) (string, error) {
	return "txn_12345", nil
}

