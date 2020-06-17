package config

import (
	"context"
	"encoding/json"

	"gopkg.in/yaml.v2"
)

//KeyType is the name of a config
type KeyType string

var creators = make(map[KeyType]Creator)

// Creator creates default config struct for a module
type Creator func() interface{}

// RegisterConfigCreator registers a config struct for parsing
func RegisterConfigCreator(name KeyType, creator Creator) {
	name += "_CONFIG"
	creators[name] = creator
}

func parseJSON(data []byte) (map[KeyType]interface{}, error) {
	result := make(map[KeyType]interface{})
	for name, creator := range creators {
		config := creator()
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
		result[name] = config
	}
	return result, nil
}

func parseYAML(data []byte) (map[KeyType]interface{}, error) {
	result := make(map[KeyType]interface{})
	for name, creator := range creators {
		config := creator()
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
		result[name] = config
	}
	return result, nil
}

func WithJSONConfig(ctx context.Context, data []byte) (context.Context, error) {
	var configs map[KeyType]interface{}
	var err error
	configs, err = parseJSON(data)
	if err != nil {
		return ctx, err
	}
	for name, config := range configs {
		ctx = context.WithValue(ctx, name, config)
	}
	return ctx, nil
}

func WithYAMLConfig(ctx context.Context, data []byte) (context.Context, error) {
	var configs map[KeyType]interface{}
	var err error
	configs, err = parseYAML(data)
	if err != nil {
		return ctx, err
	}
	for name, config := range configs {
		ctx = context.WithValue(ctx, name, config)
	}
	return ctx, nil
}

func WithConfig(ctx context.Context, name KeyType, cfg interface{}) context.Context {
	name += "_CONFIG"
	return context.WithValue(ctx, name, cfg)
}

// FromContext extracts config from a context
func FromContext(ctx context.Context, name KeyType) interface{} {
	return ctx.Value(name + "_CONFIG")
}
