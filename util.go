package cache

import (
	"errors"
	"math"
)

var (
	ErrIncrementOverflow = errors.New("this incr invocation will overflow")
	ErrDecrementOverflow = errors.New("this decr invocation will overflow")
	ErrNotIntegerType    = errors.New("item val is not (u)int (u)int32 (u)int64")
)

// Decrement Self decrement
func Decrement(originVal any, step int) (any, error) {
	switch val := originVal.(type) {
	case int:
		tmp := val - step
		if val < 0 && tmp > 0 {
			return nil, ErrDecrementOverflow
		}
		return tmp, nil
	case int32:
		if val == math.MinInt32 {
			return nil, ErrDecrementOverflow
		}
		return val - int32(step), nil
	case int64:
		if val == math.MinInt64 {
			return nil, ErrDecrementOverflow
		}
		return val - int64(step), nil
	case uint:
		if val == 0 {
			return nil, ErrDecrementOverflow
		}
		return val - uint(step), nil
	case uint32:
		if val == MinUint32 {
			return nil, ErrDecrementOverflow
		}
		return val - uint32(step), nil
	case uint64:
		if val == MinUint64 {
			return nil, ErrDecrementOverflow
		}
		return val - uint64(step), nil
	default:
		return nil, ErrNotIntegerType
	}
}

// Increment Autoincrement
func Increment(originVal any, step int) (any, error) {
	switch val := originVal.(type) {
	case int:
		tmp := val + step
		if val > 0 && tmp < 0 {
			return nil, ErrIncrementOverflow
		}
		return tmp, nil
	case int32:
		if val == math.MaxInt32 {
			return nil, ErrIncrementOverflow
		}
		return val + int32(step), nil
	case int64:
		if val == math.MaxInt64 {
			return nil, ErrIncrementOverflow
		}
		return val + int64(step), nil
	case uint:
		tmp := val + 1
		if tmp < val {
			return nil, ErrIncrementOverflow
		}
		return tmp, nil
	case uint32:
		if val == math.MaxUint32 {
			return nil, ErrIncrementOverflow
		}
		return val + uint32(step), nil
	case uint64:
		if val == math.MaxUint64 {
			return nil, ErrIncrementOverflow
		}
		return val + uint64(step), nil
	default:
		return nil, ErrNotIntegerType
	}
}
