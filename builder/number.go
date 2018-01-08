package builder

import (
	"github.com/go-json-schema/schema/common"
	"github.com/go-json-schema/validator"
	"github.com/pkg/errors"
)

type numericT interface {
	basicNumericT
	commonT
}

type basicNumericT interface {
	HasExclusiveMaximum() bool
	HasExclusiveMinimum() bool
	HasMaximum() bool
	HasMinimum() bool
	HasMultipleOf() bool
	Maximum() float64
	Minimum() float64
	MultipleOf() float64
}

type draft04NumericT interface {
	ExclusiveMaximum() bool
	ExclusiveMinimum() bool

	basicNumericT
}

type draft07NumericT interface {
	ExclusiveMaximum() float64
	ExclusiveMinimum() float64

	basicNumericT
}

func buildNumericConstraint(ctx *buildctx, nc validator.NumericConstraint, s numericT) error {
	// draft version specific
	if s1, ok := s.(draft04NumericT); ok {
		if s1.HasMinimum() {
			if s1.HasExclusiveMinimum() && s1.ExclusiveMinimum() {
				nc.ExclusiveMinimum(s1.Minimum())
			} else {
				nc.Minimum(s1.Minimum())
			}
		}

		if s1.HasMaximum() {
			if s1.HasExclusiveMaximum() && s1.ExclusiveMaximum() {
				nc.ExclusiveMaximum(s1.Maximum())
			} else {
				nc.Maximum(s1.Maximum())
			}
		}
	} else if s1, ok := s.(draft07NumericT); ok {
		if s1.HasExclusiveMinimum() {
			nc.ExclusiveMinimum(s1.ExclusiveMinimum())
		} else if s1.HasMinimum() {
			nc.Minimum(s1.Minimum())
		}

		if s1.HasExclusiveMaximum() {
			nc.ExclusiveMaximum(s1.ExclusiveMaximum())
		} else if s1.HasMaximum() {
			nc.Maximum(s1.Maximum())
		}
	}

	if s.HasMultipleOf() {
		nc.MultipleOf(s.MultipleOf())
	}

	if s.HasEnum() {
		var enums []interface{}
		for v := range s.Enum().Iterator() {
			enums = append(enums, v)
		}
		nc.Enum(enums...)
	}

	if s.HasDefault() {
		nc.Default(s.Default())
	}

	return nil
}

func buildIntegerConstraint(ctx *buildctx, nc *validator.IntegerConstraint, s numericT) error {
	if s.HasType() {
		l := s.Type()
		if len(l) <= 0 {
			return errors.New(`invalid type (empty list)`)
		}

		if !l.Contains(common.IntegerType) {
			return errors.New("schema is not for integer")
		}
	}
	return buildNumericConstraint(ctx, nc, s)
}

func buildNumberConstraint(ctx *buildctx, nc *validator.NumberConstraint, s numericT) error {
	if s.HasType() {
		l := s.Type()
		if len(l) <= 0 {
			return errors.New(`invalid type (empty list)`)
		}

		if !l.Contains(common.NumberType) {
			return errors.New("schema is not for number")
		}
	}

	return buildNumericConstraint(ctx, nc, s)
}
