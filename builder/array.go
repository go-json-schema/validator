package builder

import (
	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/draft04"
	"github.com/go-json-schema/schema/draft07"
	"github.com/go-json-schema/validator"
	"github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
)

type arrayT interface {
	HasAdditionalItems() bool
	HasItems() bool
	HasMinItems() bool
	HasMaxItems() bool
	HasUniqueItems() bool
	MinItems() int64
	MaxItems() int64
	UniqueItems() bool
}

type draft04AdditionalItems interface {
	AdditionalItems() *draft04.Schema
}

type draft07AdditionalItems interface {
	AdditionalItems() *draft07.Schema
}

type draft04Items interface {
	Items() *draft04.SchemaList
}

type draft07Items interface {
	Items() *draft07.SchemaList
}

func arrayAdditionalItems(s interface{}) (schema.Schema, error) {
	if v, ok := s.(draft04AdditionalItems); ok {
		return v.AdditionalItems(), nil
	}
	if v, ok := s.(draft07AdditionalItems); ok {
		return v.AdditionalItems(), nil
	}
	return nil, errors.New(`could not fetch additional items from schema`)
}

func arrayItems(s interface{}) ([]schema.Schema, error) {
	var schemas []schema.Schema
	if v, ok := s.(draft04Items); ok {
		for item := range v.Items().Iterator() {
			schemas = append(schemas, item)
		}
		return schemas, nil
	}
	if v, ok := s.(draft07Items); ok {
		for item := range v.Items().Iterator() {
			schemas = append(schemas, item)
		}
		return schemas, nil
	}
	return nil, errors.New(`could not fetch items from schema`)
}

func buildArrayConstraint(ctx *buildctx, c *validator.ArrayConstraint, s arrayT) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("buildArrayConstraint").BindError(&err)
		defer g.End()
	}

	if s.HasItems() {
		schemas, err := arrayItems(s)
		if err != nil {
			return errors.Wrap(err, `failed to extract items from schema`)
		}
		switch l := len(schemas); l {
		case 0:
			// WTF
			return errors.New(`invalid number of defintions in items field: 0`)
		case 1:
			specs, err := buildFromSchema(ctx, schemas[0])
			if err != nil {
				return errors.Wrap(err, `failed to build schemas for items`)
			}
			c.Items(specs)
		default:
			specs := make([]validator.Constraint, l)
			for i, espec := range schemas {
				item, err := buildFromSchema(ctx, espec)
				if err != nil {
					return errors.Wrap(err, `failed to build constraints for item elements`)
				}
				specs[i] = item
			}
			c.PositionalItems(specs)

			if !s.HasAdditionalItems() {
				if pdebug.Enabled {
					pdebug.Printf("Disabling additional items")
				}
				// No additional items
				c.AdditionalItems(nil)
			} else {
				as, err := arrayAdditionalItems(s)
				if err != nil {
					return errors.Wrap(err, `invalid additional item spec`)
				}
				spec, err := buildFromSchema(ctx, as)
				if err != nil {
					return errors.Wrap(err, `failed to build constraints for additionalItems`)
				}
				if pdebug.Enabled {
					pdebug.Printf("Using constraint for additional items ")
				}
				c.AdditionalItems(spec)
			}
		}
	}

	if s.HasMinItems() {
		c.MinItems(int(s.MinItems())) // TODO: do away with type conversion
	}

	if s.HasMaxItems() {
		c.MinItems(int(s.MaxItems())) // TODO: do away with type conversion
	}

	if s.HasUniqueItems() {
		c.UniqueItems(s.UniqueItems())
	}

	return nil
}

func schemaLooksLikeArray(s schema.Schema) bool {
	if v, ok := s.(interface {
		HasItems() bool
	}); ok && v.HasItems() {
		return true
	}

	if v, ok := s.(interface {
		HasAdditionalItems() bool
	}); ok && v.HasAdditionalItems() {
		return true
	}

	if v, ok := s.(interface {
		HasMinItems() bool
	}); ok && v.HasMinItems() {
		return true
	}

	if v, ok := s.(interface {
		HasMaxItems() bool
	}); ok && v.HasMaxItems() {
		return true
	}

	if v, ok := s.(interface {
		HasUniqueItems() bool
	}); ok && v.HasUniqueItems() {
		return true
	}

	/*
		for _, v := range s.Enum {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Slice:
				return true
			}
		}
	*/

	return false
}
