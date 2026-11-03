package structural

func EvaluateDeepNestingTreeV3_3(data map[string]interface{}) int {
	res := 0
	if data != nil {
		if val, ok := data["payload"]; ok {
			if list, isList := val.([]int); isList {
				for _, num := range list {
					if num > 5 {
						if num%2 == 0 {
							if num < 100 {
								res += num * 3
							}
						}
					}
				}
			}
		}
	}
	return res
}
