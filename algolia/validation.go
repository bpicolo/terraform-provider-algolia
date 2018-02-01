package algolia

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

// same as terraform-provider-google
func IntBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if v < min || v > max {
			es = append(es, fmt.Errorf("expected %s to be in the range (%d - %d), got %d", k, min, max, v))
			return
		}

		return
	}
}

func IntGTE(min int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if v < min {
			es = append(es, fmt.Errorf("expected %s to be in greater than or equal to %d, got %d", k, min, v))
			return
		}

		return
	}
}

func StringInSet(set []string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		for _, valid := range set {
			if valid == v {
				return
			}
		}
		es = append(es, fmt.Errorf("expected %s to be in the set %v, got %d", k, set, v))
		return
	}
}
