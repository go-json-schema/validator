package builder

import (
	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/validator"
)

type enumT interface {
	Enum() schema.EnumList
}

func buildEnumConstraint(_ *buildctx, c *validator.EnumConstraint, s enumT) error {
	for e := range s.Enum().Iterator() {
		c.Append(e)
	}
	return nil
}
