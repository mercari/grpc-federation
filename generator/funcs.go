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

// ParentCtx creates parent context name from the given level.
func ParentCtx(level int) (string, error) {
	if level <= 0 {
		return "", errors.New("level cannot be less than or equal to 0")
	}

	// the level 0 (root) context should be ctx, not ctx0
	if level == 1 {
		return "ctx", nil
	}

	return fmt.Sprintf("ctx%d", level-1), nil
}

// Add adds two numbers.
func Add(n1, n2 int) int {
	return n1 + n2
}
