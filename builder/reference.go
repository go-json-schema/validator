package builder

import (
	"errors"

	"github.com/go-json-schema/validator"
	"github.com/lestrrat/go-pdebug"
)

type referenceT interface {
	Reference() string
	HasReference() bool
}

func hasReference(s referenceT) bool {
	return s.HasReference()
}

func buildReferenceConstraint(ctx *buildctx, c *validator.ReferenceConstraint, s referenceT) (err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("ReferenceConstraint.buildFromSchema '%s'", s.Reference).BindError(&err)
		defer g.End()
	}

	if !s.HasReference() {
		return errors.New("schema does not contain a reference")
	}
	c.RefersTo(s.Reference())
	ctx.R[s.Reference()] = struct{}{}

	return nil
}
