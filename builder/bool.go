package builder

import (
	"github.com/go-json-schema/validator"
	"github.com/pkg/errors"
)

type boolT interface {
	HasDefault() bool
	Default() interface{}
}

func buildBooleanConstraint(_ *buildctx, c *validator.BooleanConstraint, s boolT) error {
	if !s.HasDefault() {
		return nil
	}

	switch v := s.Default(); v.(type) {
	case bool:
		c.Default(v.(bool))
	default:
		return errors.Errorf(`invalid default value for boolean type: %v`, v)
	}
	return nil
}
