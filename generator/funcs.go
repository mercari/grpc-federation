package generator

import (
	"errors"
	"fmt"
)

// CreateMap creates a map from a list of key-value pairs to pass multiple arguments to sub-templates.
func CreateMap(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.New("the number of arguments must be divisible by two")
	}

	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)

		if !ok {
			return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}
