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

// ErrKeyNotFound is returned when the key desired to be retrieved is not found.
var ErrKeyNotFound = errors.New("Key not found")

// FindInMap finds the provided key in incoming maps returning the value if found,
// else returning an error instead.
func FindInMap(key string) fractals.Handler {
	return fractals.MustWrap(func(target interface{}) (interface{}, error) {
		switch to := target.(type) {
		case map[string]string:
			if item, ok := to[key]; ok {
				return item, nil
			}

		case map[string]interface{}:
			if item, ok := to[key]; ok {
				return item, nil
			}
		}

		return nil, ErrKeyNotFound
	})
}

// FindKeyInMap finds the incoming key in provided map, returning the value if found,
// else returning an error instead.
func FindKeyInMap(target interface{}) fractals.Handler {
	return fractals.MustWrap(func(key string) (interface{}, error) {
		switch to := target.(type) {
		case map[string]string:
			if item, ok := to[key]; ok {
				return item, nil
			}

		case map[string]interface{}:
			if item, ok := to[key]; ok {
				return item, nil
			}
		}

		return nil, ErrKeyNotFound
	})
}

// ErrIndexOutOfBound returns this when  hte provided index is out of bounds/
var ErrIndexOutOfBound = errors.New("Index is out of bound")

// FindInList finds the provided index in incoming maps returning the value if found,
// else returning an error instead.
func FindInList(index int) fractals.Handler {
	return fractals.MustWrap(func(target interface{}) (interface{}, error) {
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
	})
}

// FindKeyInMap finds the incoming key in provided map, returning the value if found,
// else returning an error instead.
func FindIndexInList(target interface{}) fractals.Handler {
	return fractals.MustWrap(func(index int) (interface{}, error) {
		switch mo := target.(type) {
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
	})
}
