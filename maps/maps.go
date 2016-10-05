package maps

import (
	"errors"
	"strconv"
	"strings"

	"github.com/influx6/fractals"
)

// Key takes a string of period delimited values and returns a slice of interface which contains each
// piece. It converts numbers into integers ensuring to keep keys aligned.
func Keys(m string) []interface{} {
	var bkeys []interface{}

	for _, item := range strings.Split(m, ".") {
		numb, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			bkeys = append(bkeys, item)
			continue
		}

		bkeys = append(bkeys, int(numb))
	}

	return bkeys
}

// Find runs down the providded map attempting to retrieve the giving value and
// root else returning an error as failure to retrieve the giving path.
func Find(path string) fractals.Handler {

	var finders []fractals.Handler

	keys := Keys(path)
	for _, key := range keys {
		switch ikey := key.(type) {
		case int:
			finders = append(finders, FindInList(ikey))
		case string:
			finders = append(finders, FindInMap(ikey))
		}
	}

	return fractals.Lift(finders...)(nil)
}

// Save runs down the providded map attempting to retrieve the giving value and
// root else returning an error as failure to retrieve the giving path.
func Save(path string, val interface{}) fractals.Handler {

	var finders []fractals.Handler

	keys := Keys(path)

	head := keys[len(keys)-1]
	keys = keys[:len(keys)-1]

	for _, key := range keys {
		switch ikey := key.(type) {
		case int:
			finders = append(finders, FindInList(ikey))
		case string:
			finders = append(finders, FindInMap(ikey))
		}
	}

	if rh, ok := head.(string); ok {
		finders = append(finders, AddInToMap(rh, val))
	}

	if rh, ok := head.(int); ok {
		finders = append(finders, AddInToList(rh, val))
	}

	return fractals.Lift(finders...)(nil)
}

// ErrKeyNotFound is returned when the key desired to be retrieved is not found.
var ErrKeyNotFound = errors.New("Key not found")

// FindInMap finds the provided key in incoming maps returning the value if found,
// else returning an error instead.
func FindInMap(key string) fractals.Handler {
	return fractals.MustWrap(func(target interface{}) (interface{}, error) {
		return getValue(target, key)
	})
}

// FindKeyInMap finds the incoming key in provided map, returning the value if found,
// else returning an error instead.
func FindKeyInMap(target interface{}) fractals.Handler {
	return fractals.MustWrap(func(key string) (interface{}, error) {
		return getValue(target, key)
	})
}

// AddInToMap finds the provided key in incoming maps returning the value if found,
// else returning an error instead.
func AddInToMap(key string, val interface{}) fractals.Handler {
	return fractals.MustWrap(func(target interface{}) (interface{}, error) {
		if err := setValue(target, key, val); err != nil {
			return nil, err
		}

		return val, nil
	})
}

// AddIntoKeyInMap finds the incoming key in provided map, returning the value if found,
// else returning an error instead.
func AddIntoKeyInMap(target interface{}, val interface{}) fractals.Handler {
	return fractals.MustWrap(func(key string) (interface{}, error) {
		if err := setValue(target, key, val); err != nil {
			return nil, err
		}

		return val, nil
	})
}

func getValue(target interface{}, key interface{}) (interface{}, error) {
	switch to := target.(type) {
	case map[interface{}]interface{}:
		if item, ok := to[key]; ok {
			return item, nil
		}

	case map[interface{}]string:
		if item, ok := to[key]; ok {
			return item, nil
		}
	case map[string]string:
		if item, ok := to[key.(string)]; ok {
			return item, nil
		}

	case map[string]interface{}:
		if item, ok := to[key.(string)]; ok {
			return item, nil
		}
	}

	return nil, ErrKeyNotFound
}

// ErrTypeNotFound is returned when the giving type is either unknown or does not
// match their lists.
var ErrTypeNotFound = errors.New("Type not found or unknown")

func setValue(target interface{}, key interface{}, val interface{}) error {
	switch to := target.(type) {
	case map[interface{}]interface{}:
		to[key] = val
		return nil

	case map[interface{}]string:
		to[key] = val.(string)
		return nil

	case map[string]string:
		to[key.(string)] = val.(string)
		return nil

	case map[string]interface{}:
		to[key.(string)] = val
		return nil
	}

	return ErrTypeNotFound
}

// ErrIndexOutOfBound returns this when  hte provided index is out of bounds/
var ErrIndexOutOfBound = errors.New("Index is out of bound")

// FindInList finds the provided index in incoming maps returning the value if found,
// else returning an error instead.
func FindInList(index int) fractals.Handler {
	return fractals.MustWrap(func(target interface{}) (interface{}, error) {
		return getIndex(target, index)
	})
}

// FindIndexInList finds the incoming key in provided map, returning the value if found,
// else returning an error instead.
func FindIndexInList(target interface{}) fractals.Handler {
	return fractals.MustWrap(func(index int) (interface{}, error) {
		return getIndex(target, index)
	})
}

// AddInToList finds the provided index in incoming maps returning the value if found,
// else returning an error instead.
func AddInToList(index int, val interface{}) fractals.Handler {
	return fractals.MustWrap(func(target interface{}) (interface{}, error) {
		if err := setIndex(target, index, val); err != nil {
			return nil, err
		}

		return val, nil
	})
}

// AddToIndexInList finds the incoming key in provided map, returning the value if found,
// else returning an error instead.
func AddToIndexInList(target interface{}, val interface{}) fractals.Handler {
	return fractals.MustWrap(func(index int) (interface{}, error) {
		if err := setIndex(target, index, val); err != nil {
			return nil, err
		}

		return val, nil
	})
}

func getIndex(target interface{}, index int) (interface{}, error) {
	switch mo := target.(type) {
	case []map[uint]string:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil

	case []map[string]uint:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []map[string]string:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []map[string]interface{}:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []interface{}:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []rune:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []byte:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []string:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []int:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []float64:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []float32:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []uint:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []uint16:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []uint32:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	case []uint64:
		if len(mo) <= index {
			return nil, ErrIndexOutOfBound
		}

		return mo[index], nil
	}

	return nil, ErrKeyNotFound
}

func setIndex(target interface{}, index int, val interface{}) error {
	switch mo := target.(type) {
	case []map[uint]string:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(map[uint]string)
		return nil

	case []map[string]uint:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(map[string]uint)
		return nil

	case []map[string]string:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(map[string]string)
		return nil

	case []map[string]interface{}:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(map[string]interface{})
		return nil

	case []interface{}:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val
		return nil
	case []rune:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(rune)
		return nil
	case []byte:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(byte)
		return nil
	case []string:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(string)
		return nil

	case []int:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(int)
		return nil
	case []float64:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(float64)
		return nil
	case []float32:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(float32)
		return nil
	case []uint:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(uint)
		return nil
	case []uint16:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(uint16)
		return nil
	case []uint32:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(uint32)
		return nil
	case []uint64:
		if len(mo) <= index {
			return ErrIndexOutOfBound
		}

		mo[index] = val.(uint64)
		return nil
	}

	return ErrTypeNotFound
}
