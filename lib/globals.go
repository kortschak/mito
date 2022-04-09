package lib

import (
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Globals returns a cel.EnvOption to configure global variables for the environment.
// Each variable will be visible as the key in the vars map. Not all Go variable
// types are acceptable to the CEL environment, but Globals will make a best-effort
// to map types to their CEL equivalents. This typing is only done for the values
// in the map and does not apply recursively if those values are nested.
func Globals(vars map[string]interface{}) cel.EnvOption {
	return cel.Lib(globalsLib(vars))
}

type globalsLib map[string]interface{}

func (l globalsLib) CompileOptions() []cel.EnvOption {
	globals := make([]*expr.Decl, 0, len(l))
	for name, val := range l {
		var typ *expr.Type
		// Do times and []byte first since otherwise duration gets expressed as an
		// primitive:INT64 and []byte gets expressed as list_type:{elem_type:{dyn:{}}}.
		switch val.(type) {
		case time.Duration:
			typ = decls.Duration
		case time.Time:
			typ = decls.Timestamp
		case []byte:
			typ = decls.Bytes
		default:
			rt := reflect.TypeOf(val)
			kind := rt.Kind()
			var ok bool
			typ, ok = primativeTypeFor(kind)
			if ok {
				break
			}
			switch kind {
			case reflect.Slice, reflect.Array:
				elem, _ := primativeTypeFor(rt.Elem().Kind())
				typ = decls.NewListType(elem)
			case reflect.Map:
				key, _ := primativeTypeFor(rt.Key().Kind())
				elem, _ := primativeTypeFor(rt.Elem().Kind())
				typ = decls.NewMapType(key, elem)
			default:
				typ = decls.Dyn
			}
		}
		globals = append(globals, decls.NewVar(name, typ))
	}

	return []cel.EnvOption{cel.Declarations(globals...)}
}

func (l globalsLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.Globals(map[string]interface{}(l)),
	}
}

func primativeTypeFor(kind reflect.Kind) (typ *expr.Type, definitive bool) {
	switch kind {
	case reflect.Bool:
		return decls.Bool, true
	case reflect.String:
		return decls.String, true
	case reflect.Float32, reflect.Float64:
		return decls.Double, true
	case reflect.Int, reflect.Int64:
		return decls.Int, true
	case reflect.Uint, reflect.Uint64:
		return decls.Uint, true
	default:
		return decls.Dyn, false
	}
}
