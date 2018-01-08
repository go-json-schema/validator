package builder

/* Package builder contains structures and methods responsible for
 * generating a validator.JSVal structure from a JSON schema
 */

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"sort"

	"github.com/go-json-schema/schema"
	"github.com/go-json-schema/schema/common"
	"github.com/go-json-schema/schema/draft04"
	"github.com/go-json-schema/schema/draft07"
	"github.com/go-json-schema/validator"
	"github.com/lestrrat/go-jsref"
	"github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
)

type commonT interface {
	HasDefault() bool
	HasEnum() bool
	HasType() bool
	Default() interface{}
	Enum() common.EnumList
	Type() common.PrimitiveTypeList
}

// Builder builds Validator objects from JSON schemas
type Builder struct{}

type buildctx struct {
	V *validator.JSVal
	S schema.Schema
	R map[string]struct{}
}

// New creates a new builder object
func New() *Builder {
	return &Builder{}
}

type draft04Builder struct{}
type draft07Builder struct{}

// Build creates a new validator from the specified schema
func (b *Builder) Build(s schema.Schema) (v *validator.JSVal, err error) {
	if pdebug.Enabled {
		g := pdebug.IPrintf("START Builder.Build")
		defer func() {
			if err == nil {
				g.IRelease("END Builder.Build (OK)")
			} else {
				g.IRelease("END Builder.Build (FAIL): %s", err)
			}
		}()
	}

	if s == nil {
		return nil, errors.New("nil schema")
	}

	return b.BuildWithCtx(s, nil)
}

// BuildWithCtx creates a new validator from the specified schema, using
// the jsctx parameter as the context to resolve JSON References with.
// If you expect your schema to contain JSON references to itself,
// you will have to pass the context as a map with raw decoded JSON data
func (b *Builder) BuildWithCtx(s schema.Schema, jsctx interface{}) (v *validator.JSVal, err error) {
	if pdebug.Enabled {
		g := pdebug.Marker("Builder.BuildWithCtx").BindError(&err)
		defer g.End()
	}

	if s == nil {
		return nil, errors.New("nil schema")
	}

	v = validator.New()
	ctx := buildctx{
		V: v,
		S: s,
		R: map[string]struct{}{}, // names of references used
	}

	var c validator.Constraint
	c, err = buildFromSchema(&ctx, s)

	if _, ok := ctx.R["#"]; ok {
		v.SetReference("#", c)
		delete(ctx.R, "#")
	}

	// Now, resolve references that were used in the schema
	if len(ctx.R) > 0 {
		if pdebug.Enabled {
			pdebug.Printf("Checking references now")
		}
		if jsctx == nil {
			jsctx = s
		}

		r := jsref.New()
		for ref := range ctx.R {
			if err := compileReferences(&ctx, r, v, ref, jsctx); err != nil {
				return nil, err
			}
		}
	}
	v.SetRoot(c)
	return v, nil
}

func compileReferences(ctx *buildctx, r *jsref.Resolver, v *validator.JSVal, ref string, jsctx interface{}) error {
	if _, err := v.GetReference(ref); err == nil {
		if pdebug.Enabled {
			pdebug.Printf("Already resolved constraints for reference '%s'", ref)
		}
		return nil
	}

	if pdebug.Enabled {
		pdebug.Printf("Building constraints for reference '%s'", ref)
	}

	thing, err := r.Resolve(jsctx, ref)
	if err != nil {
		return err
	}

	if pdebug.Enabled {
		pdebug.Printf("'%s' resolves to the main schema", ref)
	}

	var s1 schema.Schema
	switch thing.(type) {
	case schema.Schema:
		s1 = thing.(schema.Schema)
	case map[string]interface{}:
		s1 = &draft04.Schema{} // schema.New()
		// XXX Very inefficient, should probably fix
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(thing); err != nil {
			return errors.Wrap(err, `failed to encode resolved schema`)
		}
		if err := json.NewDecoder(&buf).Decode(s1); err != nil {
			return errors.Wrap(err, `failed to decode resolved schema`)
		}
	}

	c1, err := buildFromSchema(ctx, s1)
	if err != nil {
		return err
	}

	v.SetReference(ref, c1)
	for ref := range ctx.R {
		if err := compileReferences(ctx, r, v, ref, jsctx); err != nil {
			return err
		}
	}
	return nil
}

func buildFromSchema(ctx *buildctx, s schema.Schema) (validator.Constraint, error) {
	var c validator.Constraint
	var err error
	switch v := s.(type) {
	case *draft04.Schema:
		c, err = buildFromDraft04Schema(ctx, v)
		if err != nil {
			return nil, errors.Wrap(err, `failed to build draft-04 validator`)
		}
	case *draft07.Schema:
		c, err = buildFromDraft07Schema(ctx, v)
		if err != nil {
			return nil, errors.Wrap(err, `failed to build draft-07 validator`)
		}
	default:
		return nil, errors.Errorf(`invalid schema %T`, s)
	}
	return c, nil
}

func buildFromDraft07Schema(ctx *buildctx, s *draft07.Schema) (validator.Constraint, error) {
	if hasReference(s) {
		c := validator.Reference(ctx.V)
		if err := buildReferenceConstraint(ctx, c, s); err != nil {
			return nil, errors.Wrap(err, `failed to build reference constraint`)
		}
		return c, nil
	}

	ct := validator.All()

	return ct, nil
}

func buildFromDraft04Schema(ctx *buildctx, s *draft04.Schema) (validator.Constraint, error) {
	if hasReference(s) {
		c := validator.Reference(ctx.V)
		if err := buildReferenceConstraint(ctx, c, s); err != nil {
			return nil, errors.Wrap(err, `failed to build reference constraint`)
		}
		return c, nil
	}

	ct := validator.All()

	switch {
	case s.HasNot():
		if pdebug.Enabled {
			pdebug.Printf("Not constraint")
		}
		c1, err := buildFromSchema(ctx, s.Not())
		if err != nil {
			return nil, err
		}
		ct.Add(validator.Not(c1))
	case s.HasAllOf():
		if pdebug.Enabled {
			pdebug.Printf("AllOf constraint")
		}
		ac := validator.All()
		for s1 := range s.AllOf().Iterator() {
			c1, err := buildFromSchema(ctx, s1)
			if err != nil {
				return nil, err
			}
			ac.Add(c1)
		}
		ct.Add(ac.Reduce())
	case s.HasAnyOf():
		if pdebug.Enabled {
			pdebug.Printf("AnyOf constraint")
		}
		ac := validator.Any()
		for s1 := range s.AnyOf().Iterator() {
			c1, err := buildFromSchema(ctx, s1)
			if err != nil {
				return nil, err
			}
			ac.Add(c1)
		}
		ct.Add(ac.Reduce())
	case s.HasOneOf():
		if pdebug.Enabled {
			pdebug.Printf("OneOf constraint")
		}
		oc := validator.OneOf()
		for s1 := range s.OneOf().Iterator() {
			c1, err := buildFromSchema(ctx, s1)
			if err != nil {
				return nil, err
			}
			oc.Add(c1)
		}
		ct.Add(oc.Reduce())
	}

	var sts common.PrimitiveTypeList
	if s.HasType() {
		l := s.Type()
		sts = make(common.PrimitiveTypeList, len(l))
		copy(sts, l)
	} else {
		if pdebug.Enabled {
			pdebug.Printf("Schema doesn't seem to contain a 'type' field. Now guessing...")
		}
		sts = guessSchemaType(s)
	}
	sort.Sort(sts)

	if len(sts) > 0 {
		tct := validator.Any()
		for _, st := range sts {
			var c validator.Constraint
			switch st {
			case common.StringType:
				sc := validator.String()
				if err := buildStringConstraint(ctx, sc, s); err != nil {
					return nil, err
				}
				c = sc
			case common.NumberType:
				nc := validator.Number()
				if err := buildNumberConstraint(ctx, nc, s); err != nil {
					return nil, err
				}
				c = nc
			case common.IntegerType:
				ic := validator.Integer()
				if err := buildIntegerConstraint(ctx, ic, s); err != nil {
					return nil, err
				}
				c = ic
			case common.BooleanType:
				bc := validator.Boolean()
				if err := buildBooleanConstraint(ctx, bc, s); err != nil {
					return nil, err
				}
				c = bc
			case common.ArrayType:
				ac := validator.Array()
				if err := buildArrayConstraint(ctx, ac, s); err != nil {
					return nil, err
				}
				c = ac
			case common.ObjectType:
				oc := validator.Object()
				if err := buildObjectConstraint(ctx, oc, s); err != nil {
					return nil, err
				}
				c = oc
			case common.NullType:
				c = validator.NullConstraint
			default:
				return nil, errors.New("unknown type: " + st.String())
			}
			tct.Add(c)
		}
		ct.Add(tct.Reduce())
	} else {
		// All else failed, check if we have some enumeration?
		if s.HasEnum() {
			ec := validator.Enum()
			for e := range s.Enum().Iterator() {
				ec.Append(e)
			}
			ct.Add(ec)
		}
	}

	return ct.Reduce(), nil
}

func guessSchemaType(s schema.Schema) common.PrimitiveTypeList {
	if pdebug.Enabled {
		g := pdebug.Marker("guessSchemaType")
		defer g.End()
	}

	var sts common.PrimitiveTypeList
	if schemaLooksLikeObject(s) {
		if pdebug.Enabled {
			pdebug.Printf("Looks like it could be an object...")
		}
		sts = append(sts, common.ObjectType)
	}

	if schemaLooksLikeArray(s) {
		if pdebug.Enabled {
			pdebug.Printf("Looks like it could be an array...")
		}
		sts = append(sts, common.ArrayType)
	}

	if schemaLooksLikeString(s) {
		if pdebug.Enabled {
			pdebug.Printf("Looks like it could be a string...")
		}
		sts = append(sts, common.StringType)
	}

	if ok, typ := schemaLooksLikeNumber(s); ok {
		if pdebug.Enabled {
			pdebug.Printf("Looks like it could be a number...")
		}
		sts = append(sts, typ)
	}

	if schemaLooksLikeBool(s) {
		if pdebug.Enabled {
			pdebug.Printf("Looks like it could be a bool...")
		}
		sts = append(sts, common.BooleanType)
	}

	if pdebug.Enabled {
		pdebug.Printf("Guessed types: %#v", sts)
	}

	return sts
}

func numberLooksLikeInteger(n float64) bool {
	return math.Floor(n) == n
}

func schemaLooksLikeNumber(s schema.Schema) (bool, common.PrimitiveType) {
	n, ok := s.(basicNumericT)
	if !ok {
		return false, common.UnspecifiedType
	}

	if n.HasMultipleOf() {
		if numberLooksLikeInteger(n.MultipleOf()) {
			return true, common.IntegerType
		}
		return true, common.NumberType
	}

	if n.HasMinimum() {
		if numberLooksLikeInteger(n.Minimum()) {
			return true, common.IntegerType
		}
		return true, common.NumberType
	}

	if n.HasMaximum() {
		if numberLooksLikeInteger(n.Maximum()) {
			return true, common.IntegerType
		}
		return true, common.NumberType
	}

	if n.HasExclusiveMinimum() {
		if n1, ok := n.(draft07NumericT); ok && numberLooksLikeInteger(n1.ExclusiveMinimum()) {
			return true, common.IntegerType
		}
		return true, common.NumberType
	}

	if n.HasExclusiveMaximum() {
		if n1, ok := n.(draft07NumericT); ok && numberLooksLikeInteger(n1.ExclusiveMaximum()) {
			return true, common.IntegerType
		}
		return true, common.NumberType
	}

	/*
		for _, v := range s.Enum {
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return true, common.IntegerType
			case reflect.Float32, reflect.Float64:
				return true, common.NumberType
			}
		}
	*/

	return false, common.UnspecifiedType
}

func schemaLooksLikeBool(s schema.Schema) bool {
	e, ok := s.(interface {
		HasEnum() bool
		Enum() common.EnumList
	})
	if !ok {
		return false
	}

	for v := range e.Enum().Iterator() {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Bool:
			return true
		}
	}

	return false
}
