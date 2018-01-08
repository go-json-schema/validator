package builder

import (
	"regexp"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/draft04"
	"github.com/go-json-schema/schema/draft07"
	"github.com/go-json-schema/validator"
	"github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
)

type objectT interface {
	HasMaxProperties() bool
	HasMinProperties() bool
	HasRequired() bool
	MaxProperties() int64
	MinProperties() int64
	Required() []string
}

func buildObjectConstraint(ctx *buildctx, c *validator.ObjectConstraint, s objectT) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("ObjectConstraint.FromSchema").BindError(&err)
		defer g.End()
	}

	if s.HasRequired() {
		c.Required(s.Required()...)
	}

	if s.HasMinProperties() {
		c.MinProperties(s.MinProperties())
	}

	if s.HasMaxProperties() {
		c.MaxProperties(s.MaxProperties())
	}

	switch v := s.(type) {
	case *draft04.Schema:
		return buildDraft04ObjectConstraint(ctx, c, v)
	case *draft07.Schema:
		return buildDraft07ObjectConstraint(ctx, c, v)
	default:
		return errors.New(`invalid schema type`)
	}
}

func buildDraft07ObjectConstraint(ctx *buildctx, c *validator.ObjectConstraint, s *draft07.Schema) error {
	return nil
}

func buildDraft04ObjectConstraint(ctx *buildctx, c *validator.ObjectConstraint, s *draft04.Schema) error {
	if s.HasProperties() {
		for prop := range s.Properties().Iterator() {
			cprop, err := buildFromSchema(ctx, prop.Definition())
			if err != nil {
				return err
			}

			c.AddProp(prop.Name(), cprop)
		}
	}

	if s.HasPatternProperties() {
		for prop := range s.PatternProperties().Iterator() {
			cprop, err := buildFromSchema(ctx, prop.Definition())
			if err != nil {
				return err
			}
			rx, err := regexp.Compile(prop.Name())
			if err != nil {
				return errors.Wrap(err, `failed to compile regular expression`)
			}

			c.PatternProperties(rx, cprop)
		}
	}

	if !s.HasAdditionalProperties() {
		c.AdditionalProperties(validator.EmptyConstraint)
	} else {
		aitem, err := buildFromSchema(ctx, s.AdditionalProperties())
		if err != nil {
			return errors.Wrap(err, `failed to build additional proerties schema`)
		}
		c.AdditionalProperties(aitem)
	}

	if s.HasDependencies() {
	for from, to := range s.Dependencies().Names() {
		c.PropDependency(from, to...)
	}

	for prop := range s.Dependencies().Schemas().Iterator() {
			depc, err := buildFromSchema(ctx, prop.Definition())
			if err != nil {
				return errors.Wrapf(err, `failed to build dependency %s`, prop.Name())
			}

			c.SchemaDependency(prop.Name(), depc)
		}
	}

	return nil
}

func schemaLooksLikeObject(s schema.Schema) bool {
	if v, ok := s.(interface {
		HasProperties() bool
	}); ok && v.HasProperties() {
		return true
	}

	if v, ok := s.(interface {
		HasAdditionalProperties() bool
	}); ok && v.HasAdditionalProperties() {
		return true
	}

	if v, ok := s.(interface {
		HasMinProperties() bool
	}); ok && v.HasMinProperties() {
		return true
	}

	if v, ok := s.(interface {
		HasMaxProperties() bool
	}); ok && v.HasMaxProperties() {
		return true
	}

	if v, ok := s.(interface {
		HasRequired() bool
	}); ok && v.HasRequired() {
		return true
	}

	if v, ok := s.(interface {
		HasPatternProperties() bool
	}); ok && v.HasPatternProperties() {
		return true
	}

	/*
		for _, v := range s.Enum {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Map, reflect.Struct:
				return true
			}
		}
	*/

	return false
}
