package builder

import (
	"fmt"
	"strings"
)

func getValue(value interface{}) string {
	res := ""
	switch value.(type) {
	case string:
		res = fmt.Sprintf("'%s'", value.(string))
	case bool:
		res = fmt.Sprintf("%t", value.(bool))
	case int64:
		res = fmt.Sprintf("%d", value.(int64))
	case float64:
		res = fmt.Sprintf("%f", value.(float64))
	case []int64:
		var values []string
		for _, val := range value.([]int64) {
			values = append(values, fmt.Sprintf("%d", val))
		}

		res = fmt.Sprintf("ARRAY [%s]", strings.Join(values, ", "))
	default:
		res = "''"
	}

	return res
}
