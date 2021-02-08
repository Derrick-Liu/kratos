package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/config/source"
	"github.com/imdario/mergo"
)

// Reader is config reader.
type Reader interface {
	Merge(...*source.KeyValue) error
	Value(string) (Value, bool)
	Source() ([]byte, error)
}

type reader struct {
	opts   options
	values map[string]interface{}
}

func newReader(opts options) Reader {
	return &reader{
		opts:   opts,
		values: make(map[string]interface{}),
	}
}

func (r *reader) Merge(kvs ...*source.KeyValue) error {
	merged, err := cloneMap(r.values)
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		var next map[string]interface{}
		if err := r.opts.decoder(kv, &next); err != nil {
			return err
		}
		if err := mergo.Map(&merged, convertMap(next), mergo.WithOverride); err != nil {
			return err
		}
	}
	r.values = merged
	return nil
}

func (r *reader) Value(path string) (Value, bool) {
	var (
		next = r.values
		keys = strings.Split(path, ".")
		last = len(keys) - 1
	)
	for idx, key := range keys {
		value, ok := next[key]
		if !ok {
			return nil, false
		}
		if idx == last {
			av := &atomicValue{}
			av.Store(value)
			return av, true
		}
		switch value.(type) {
		case map[string]interface{}:
			next = value.(map[string]interface{})
		default:
			return nil, false
		}
	}
	return nil, false
}

func (r *reader) Source() ([]byte, error) {
	return json.Marshal(r.values)
}

func cloneMap(src map[string]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	dst := make(map[string]interface{})
	if err = json.Unmarshal(data, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}

func convertMap(src interface{}) interface{} {
	switch m := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{})
		for k, v := range m {
			dst[k] = convertMap(v)
		}
		return dst
	case map[interface{}]interface{}:
		dst := make(map[string]interface{})
		for k, v := range m {
			dst[fmt.Sprint(k)] = convertMap(v)
		}
		return dst
	default:
		return src
	}
}
