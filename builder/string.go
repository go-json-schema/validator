package builder

import (
	"reflect"
	"regexp"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/common"
	"github.com/go-json-schema/validator"
	"github.com/pkg/errors"
)

type stringT interface {
	HasFormat() bool
	HasMaxLength() bool
	HasMinLength() bool
	HasPattern() bool
	Format() common.Format
	MaxLength() int64
	MinLength() int64
	Pattern() string

	commonT
}

func buildStringConstraint(ctx *buildctx, c *validator.StringConstraint, s stringT) error {
	if s.HasType() {
		if !s.Type().Contains(common.StringType) {
			return errors.New("schema is not for string")
		}
	}

	if s.HasMaxLength() {
		c.MaxLength(s.MaxLength())
	}

	if s.HasMinLength() {
		c.MinLength(s.MinLength())
	}

	if s.HasPattern() {
		rx, err := regexp.Compile(s.Pattern())
		if err != nil {
			return errors.Wrap(err, `failed to compile pattern`)
		}
		c.Regexp(rx)
	}

	if s.HasFormat() {
		c.Format(string(s.Format()))
	}

	if s.HasEnum() {
		var enums []interface{}
		for v := range s.Enum().Iterator() {
			enums = append(enums, v)
		}
		c.Enum(enums...)
	}

	if s.HasDefault() {
		c.Default(s.Default())
	}

	return nil
}

func schemaLooksLikeString(s schema.Schema) bool {
	if v, ok := s.(interface {
		HasMinLength() bool
	}); ok && v.HasMinLength() {
		return true
	}

	if v, ok := s.(interface {
		HasMaxLength() bool
	}); ok && v.HasMaxLength() {
		return true
	}

	if v, ok := s.(interface {
		HasPattern() bool
	}); ok && v.HasPattern() {
		return true
	}

	if v, ok := s.(interface {
		HasFormat() bool
	}); ok && v.HasFormat() {
		return true
	}

	enumContainer, ok := s.(interface {
		HasEnum() bool
		Enum() common.EnumList
	})
	if ok && enumContainer.HasEnum() {
		for v := range enumContainer.Enum().Iterator() {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.String:
				return true
			}
		}
	}

	return false
}
