package validator

import (
	"errors"
	"math"
	"reflect"

	"github.com/lestrrat/go-pdebug"
)

// Enum specifies the values that this constraint can have
func (nc *NumberConstraint) Enum(l ...interface{}) NumericConstraint {
	if nc.enums == nil {
		nc.enums = Enum()
	}
	nc.enums.Enum(l...)
	return nc
}

// Default specifies the default value for the given value
func (nc *NumberConstraint) Default(v interface{}) NumericConstraint {
	nc.defaultValue.initialized = true
	nc.defaultValue.value = v
	return nc
}

// Maximum sepcifies the maximum value that the constraint can allow
func (nc *NumberConstraint) Maximum(n float64) NumericConstraint {
	nc.applyMaximum = applyLimitInclusive
	nc.maximum = n
	return nc
}

// Minimum sepcifies the minimum value that the constraint can allow
func (nc *NumberConstraint) Minimum(n float64) NumericConstraint {
	nc.applyMinimum = applyLimitInclusive
	nc.minimum = n
	return nc
}

// MultipleOf specifies the number that the given value must be
// divisible by. That is, the constraint will return an error unless
// the given value satisfies `math.Mod(v, n) == 0`
func (nc *NumberConstraint) MultipleOf(n float64) NumericConstraint {
	nc.applyMultipleOf = true
	nc.multipleOf = n
	return nc
}

// ExclusiveMinimum specifies the minimum valid value excluding the specified value
func (nc *NumberConstraint) ExclusiveMinimum(v float64) NumericConstraint {
	nc.applyMinimum = applyLimitExclusive
	nc.minimum = v
	return nc
}

// ExclusiveMaximum specifies the maximum valid value excluding the specified value
func (nc *NumberConstraint) ExclusiveMaximum(v float64) NumericConstraint {
	nc.applyMaximum = applyLimitExclusive
	nc.maximum = v
	return nc
}

// Validate validates the value against this constraint
func (nc *NumberConstraint) Validate(v interface{}) (err error) {
	if pdebug.Enabled {
		g := pdebug.IPrintf("START NumberConstraint.Validate")
		defer func() {
			if err == nil {
				g.IRelease("END NumberConstraint.Validate (PASS)")
			} else {
				g.IRelease("END NumberConstraint.Validate (FAIL): %s", err)
			}
		}()
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
	default:
		return errors.New("value is not a float")
	}

	f := rv.Float()

	switch nc.applyMinimum {
	case applyLimitNone:
	case applyLimitInclusive:
		if pdebug.Enabled {
			pdebug.Printf("Checking inclusive minimum (%f)", nc.minimum)
		}
		if nc.minimum > f {
			return errors.New("numeric value is less than the minimum")
		}
	case applyLimitExclusive:
		if pdebug.Enabled {
			pdebug.Printf("Checking exclusive minimum (%f)", nc.minimum)
		}
		if nc.minimum >= f {
			return errors.New("numeric value is less than the minimum")
		}
	}

	switch nc.applyMaximum {
	case applyLimitNone:
	case applyLimitInclusive:
		if pdebug.Enabled {
			pdebug.Printf("Checking inclusive maximum (%f)", nc.maximum)
		}
		if nc.maximum > f {
			return errors.New("numeric value is less than the maximum")
		}
	case applyLimitExclusive:
		if pdebug.Enabled {
			pdebug.Printf("Checking exclusive maximum (%f)", nc.maximum)
		}
		if nc.maximum >= f {
			return errors.New("numeric value is less than the maximum")
		}
	}

	if nc.applyMultipleOf {
		if pdebug.Enabled {
			pdebug.Printf("Checking MultipleOf (%f)", nc.multipleOf)
		}

		if nc.multipleOf != 0 {
	    if math.Mod(f, nc.multipleOf) != 0 {
				return errors.New("numeric value is fails multipleOf validation")
			}
		}
	}

	if enum := nc.enums; enum != nil {
		if err := enum.Validate(f); err != nil {
			return err
		}
	}

	return nil
}

// Number creates a new NumberConstraint
func Number() *NumberConstraint {
	return &NumberConstraint{}
}

// Integer creates a new IntegerrConstraint
func Integer() *IntegerConstraint {
	c := &IntegerConstraint{}
	return c
}

// Default specifies the default value for the given value
func (ic *IntegerConstraint) Default(v interface{}) NumericConstraint {
	ic.NumberConstraint.Default(v)
	return ic
}

// Maximum sepcifies the maximum value that the constraint can allow
func (ic *IntegerConstraint) Maximum(n float64) NumericConstraint {
	ic.NumberConstraint.Maximum(n)
	return ic
}

// Minimum sepcifies the minimum value that the constraint can allow
func (ic *IntegerConstraint) Minimum(n float64) NumericConstraint {
	ic.NumberConstraint.Minimum(n)
	return ic
}

// ExclusiveMinimum specifies the minimum valid value excluding the specified value
func (ic *IntegerConstraint) ExclusiveMinimum(v float64) NumericConstraint {
	ic.NumberConstraint.ExclusiveMinimum(v)
	return ic
}

// ExclusiveMaximum specifies the maximum valid value excluding the specified value
func (ic *IntegerConstraint) ExclusiveMaximum(v float64) NumericConstraint {
	ic.NumberConstraint.ExclusiveMaximum(v)
	return ic
}

// Validate validates the value against integer validation rules.
// Note that because when Go decodes JSON it FORCES float64 on numbers,
// this method will return true even if the *type* of the value is
// float32/64. We just check that `math.Floor(v) == v`
func (ic *IntegerConstraint) Validate(v interface{}) (err error) {
	if pdebug.Enabled {
		g := pdebug.IPrintf("START IntegerConstraint.Validate")
		defer func() {
			if err == nil {
				g.IRelease("END IntegerConstraint.Validate (PASS)")
			} else {
				g.IRelease("END IntegerConstraint.Validate (FAIL): %s", err)
			}
		}()
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Interface, reflect.Ptr:
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ic.NumberConstraint.Validate(float64(rv.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return ic.NumberConstraint.Validate(float64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		fv := rv.Float()
		if math.Floor(fv) != fv {
			return errors.New("value is not an int/uint")
		}
		return ic.NumberConstraint.Validate(fv)
	default:
		return errors.New("value is not numeric")
	}
}
