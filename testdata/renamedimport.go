package testdata

import (
	ref "reflect"
)

//handleall:ref.Kind
func HandleKinds(k ref.Kind) bool {
	switch k {
	case ref.Invalid:
	case ref.Bool:
	case ref.Int:
	case ref.Int8:
	case ref.Int16:
	case ref.Int32:
	case ref.Int64:
	case ref.Uint:
	case ref.Uint8:
	case ref.Uint16:
	case ref.Uint32:
	case ref.Uint64:
	case ref.Uintptr:
	case ref.Float32:
	case ref.Float64:
	case ref.Complex64:
	case ref.Complex128:
	case ref.Array:
	case ref.Chan:
	case ref.Func:
	case ref.Interface:
	case ref.Map:
	case ref.Ptr:
	case ref.Slice:
	case ref.String:
	case ref.Struct:
	case ref.UnsafePointer:
	default:
		return false
	}
	return true
}

// MATCH:46 "Unhandled const: Invalid (0)"

//handleall:ref.Kind
func HandleKindsMinusInvalid(k ref.Kind) bool {
	switch k {
	case ref.Bool:
	case ref.Int:
	case ref.Int8:
	case ref.Int16:
	case ref.Int32:
	case ref.Int64:
	case ref.Uint:
	case ref.Uint8:
	case ref.Uint16:
	case ref.Uint32:
	case ref.Uint64:
	case ref.Uintptr:
	case ref.Float32:
	case ref.Float64:
	case ref.Complex64:
	case ref.Complex128:
	case ref.Array:
	case ref.Chan:
	case ref.Func:
	case ref.Interface:
	case ref.Map:
	case ref.Ptr:
	case ref.Slice:
	case ref.String:
	case ref.Struct:
	case ref.UnsafePointer:
	default:
		return false
	}
	return true
}
