package main

import (
	"context"
	"fmt"
	"reflect"
)

type Tests struct{}

func (m *Tests) All(ctx context.Context) error {
	t := []func(ctx context.Context) error{
		m.TestGenerateCfgSchema,
		m.TestConfigParsing,
	}

	for _, t1 := range t {
		if err := t1(ctx); err != nil {
			return fmt.Errorf("ERROR: %s: %w", reflect.TypeOf(t1).Name(), err)
		}

	}
	return nil
}

func (m *Tests) TestGenerateCfgSchema(ctx context.Context) error {
	_, err := dag.Ci().ConfigJSONSchema(ctx)
	return err
}

func (m *Tests) TestConfigParsing(ctx context.Context) error {
	c := dag.Ci().WithConfig()
	if json, err := c.ConfigJSONSchema(ctx); err != nil {
		return err
	} else if json == "" {
		return fmt.Errorf("Output is empty")
	}
	return nil
}
