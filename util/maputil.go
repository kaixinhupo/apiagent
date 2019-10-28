package util

func CopyMap(src map[string]interface{}, target map[string]interface{}) {
	for k, v := range src {
		target[k] = v
	}
}
