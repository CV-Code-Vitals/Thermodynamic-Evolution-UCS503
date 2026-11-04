package cast

import (
	"encoding/json"
	"os"
)

// Neg_Control_LoadConfigSafe checks and propagates JSON unmarshal error
func LoadJSONConfigSafe(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// INJ-EXT2-ERR-04: Critical error on json.Unmarshal discarded via blank identifier
func ParsePayloadUnchecked(data []byte) map[string]interface{} {
	var res map[string]interface{}
	_ = json.Unmarshal(data, &res) // UNHANDLED_ERROR: discarded error
	return res
}
