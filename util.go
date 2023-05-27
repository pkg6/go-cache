package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
)

var (
	ErrIncrementOverflow = errors.New("this incr invocation will overflow")
	ErrDecrementOverflow = errors.New("this decr invocation will overflow")
	ErrNotIntegerType    = errors.New("item val is not (u)int (u)int32 (u)int64")
)

// GobEncode Gob encodes a file cache item.
func GobEncode(data any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, errors.New("could not encode this data")
	}
	return buf.Bytes(), nil
}

// GobDecode Gob decodes a file cache item.
func GobDecode(data []byte, to *CacheItem) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&to)
	if err != nil {
		return errors.New("could not decode this data to FileCacheItem. Make sure that the data is encoded by GOB")
	}
	return nil
}

func UnwrapF(format string, a ...any) error {
	return errors.Unwrap(fmt.Errorf(format, a...))
}
func WrapF(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

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
