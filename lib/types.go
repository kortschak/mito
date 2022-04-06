package lib

import (
	"reflect"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/types/known/anypb"
)

// Types used in overloads.
var (
	typeV        = decls.NewTypeParamType("V")
	typeK        = decls.NewTypeParamType("K")
	mapKV        = decls.NewMapType(typeK, typeV)
	mapStringDyn = decls.NewMapType(decls.String, decls.Dyn)
	listV        = decls.NewListType(typeV)
)

// Types used for conversion to native.
var (
	// encodableTypes is the preferred type correspondence between CEL types
	// and Go types. This mapping must be kept in agreement with the types in
	// cel-go/common/types.
	encodableTypes = map[ref.Type]reflect.Type{
		types.BoolType:      reflectBoolType,
		types.BytesType:     reflectByteSliceType,
		types.DoubleType:    reflect.TypeOf(float64(0)),
		types.DurationType:  reflect.TypeOf(time.Duration(0)),
		types.IntType:       reflectInt64Type,
		types.ListType:      reflect.TypeOf([]interface{}(nil)),
		types.MapType:       reflectMapStringAnyType,
		types.NullType:      reflect.TypeOf((*int)(nil)), // Any pointer will do.
		types.StringType:    reflectStringType,
		types.TimestampType: reflect.TypeOf(time.Time{}),
		types.UintType:      reflect.TypeOf(uint64(0)),
		types.UnknownType:   reflect.TypeOf([]int64(types.Unknown(nil))), // Double conversion to catch type changes.
	}

	// Linear search for proto.Message mappings and others.
	protobufTypes = []reflect.Type{
		reflect.TypeOf((*structpb.Value)(nil)),
		reflect.TypeOf((*structpb.ListValue)(nil)),
		reflect.TypeOf((*structpb.Struct)(nil)),
		// Catch all.
		reflect.TypeOf((*anypb.Any)(nil)),
	}
)

// Types used for reflect conversion.
var (
	reflectBoolType                 = reflect.TypeOf(true)
	reflectByteSliceType            = reflect.TypeOf([]byte(nil))
	reflectIntType                  = reflect.TypeOf(0)
	reflectInt64Type                = reflect.TypeOf(int64(0))
	reflectMapStringAnyType         = reflect.TypeOf(map[string]interface{}(nil))
	reflectMapStringStringSliceType = reflect.TypeOf(map[string][]string(nil))
	reflectStringType               = reflect.TypeOf("")
	reflectStringSliceType          = reflect.TypeOf([]string(nil))
)
