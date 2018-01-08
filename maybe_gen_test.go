package validator_test

import (
	"testing"

	"github.com/go-json-schema/validator"
)

func TestSanity(t *testing.T) {
	t.Run("MaybeBool", func(t *testing.T) {
		var v validator.Maybe
		v = &validator.MaybeBool{}
		_ = v
	})
	t.Run("MaybeFloat", func(t *testing.T) {
		var v validator.Maybe
		v = &validator.MaybeFloat{}
		_ = v
	})
	t.Run("MaybeInt", func(t *testing.T) {
		var v validator.Maybe
		v = &validator.MaybeInt{}
		_ = v
	})
	t.Run("MaybeString", func(t *testing.T) {
		var v validator.Maybe
		v = &validator.MaybeString{}
		_ = v
	})
	t.Run("MaybeTime", func(t *testing.T) {
		var v validator.Maybe
		v = &validator.MaybeTime{}
		_ = v
	})
	t.Run("MaybeUint", func(t *testing.T) {
		var v validator.Maybe
		v = &validator.MaybeUint{}
		_ = v
	})
}
