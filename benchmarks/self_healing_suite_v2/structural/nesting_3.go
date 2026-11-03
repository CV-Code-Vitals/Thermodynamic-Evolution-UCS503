package structural

func ProcessDeepNesting3(data map[string]interface{}) int {
	res := 0
	if data != nil {
		if val, ok := data["key"]; ok {
			if list, isList := val.([]int); isList {
				for _, item := range list {
					if item > 10 {
						if item%2 == 0 {
							res += item
						}
					}
				}
			}
		}
	}
	return res
}
