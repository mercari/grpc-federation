package federation

import "fmt"

func Int64ToInt32(v int64) (int32, error) {
	if v < -1*(1<<31) || (1<<31) <= v {
		return 0, fmt.Errorf("failed to convert int64(%d) to int32: %w", v, ErrOverflowTypeConversion)
	}
	return int32(v), nil
}

func Int64ToUint32(v int64) (uint32, error) {
	if v < 0 || (1<<32) <= v {
		return 0, fmt.Errorf("failed to convert int64(%d) to uint32: %w", v, ErrOverflowTypeConversion)
	}
	return uint32(v), nil
}

func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("failed to convert int64(%d) to uint64: %w", v, ErrOverflowTypeConversion)
	}
	return uint64(v), nil
}

func Int32ToUint32(v int32) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("failed to convert int32(%d) to uint32: %w", v, ErrOverflowTypeConversion)
	}
	return uint32(v), nil
}

func Int32ToUint64(v int32) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("failed to convert int32(%d) to uint64: %w", v, ErrOverflowTypeConversion)
	}
	return uint64(v), nil
}

func Uint64ToInt32(v uint64) (int32, error) {
	if (1 << 31) <= v {
		return 0, fmt.Errorf("failed to convert uint64(%d) to int32: %w", v, ErrOverflowTypeConversion)
	}
	return int32(v), nil
}

func Uint64ToInt64(v uint64) (int64, error) {
	if (1 << 63) <= v {
		return 0, fmt.Errorf("failed to convert uint64(%d) to int64: %w", v, ErrOverflowTypeConversion)
	}
	return int64(v), nil
}

func Uint64ToUint32(v uint64) (uint32, error) {
	if (1 << 32) <= v {
		return 0, fmt.Errorf("failed to convert uint64(%d) to uint32: %w", v, ErrOverflowTypeConversion)
	}
	return uint32(v), nil
}

func Uint32ToInt32(v uint32) (int32, error) {
	if (1 << 31) <= v {
		return 0, fmt.Errorf("failed to convert uint32(%d) to int32: %w", v, ErrOverflowTypeConversion)
	}
	return int32(v), nil
}
