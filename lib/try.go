package lib

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Try returns a cel.EnvOption to configure extended functions for allowing
// errors to be weakened to strings or objects.
//
// Try
//
// try returns either passes a value through unaltered if it is valid and
// not an error, or it returns a string or object describing the error:
//
//     try(<error>) -> <map<string,sting>>
//     try(<dyn>) -> <dyn>
//     try(<error>, <string>) -> <map<string,sting>>
//     try(<dyn>, <string>) -> <dyn>
//
// Examples:
//
//     try(0/1)            // return 0
//     try(0/0)            // return "division by zero"
//     try(0/0, "error")   // return {"error": "division by zero"}
//
//
// Is Error
//
// is_error returns a bool indicating whether the argument is an error:
//
//     is_error(<dyn>) -> <bool>
//
// Examples:
//
//     is_error(0/1)            // return false
//     is_error(0/0)            // return true
//
// Depends on https://github.com/google/cel-go/issues/525.
func Try() cel.EnvOption {
	return cel.Lib(tryLib{})
}

type tryLib struct{}

func (tryLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Declarations(
			decls.NewFunction("try",
				decls.NewOverload(
					"try_dyn",
					[]*expr.Type{decls.Dyn},
					decls.Dyn,
				),
				decls.NewOverload(
					"try_dyn_string",
					[]*expr.Type{decls.Dyn, decls.String},
					decls.Dyn,
				),
			),
			decls.NewFunction("is_error",
				decls.NewOverload(
					"is_error_dyn",
					[]*expr.Type{decls.Dyn},
					decls.Bool,
				),
			),
		),
	}
}

func (tryLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator:   "try_dyn",
				Unary:      try,
				AllowError: true,
			},
			&functions.Overload{
				Operator:   "try_dyn_string",
				Binary:     tryMessage,
				AllowError: true,
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator:   "is_error_dyn",
				Unary:      isError,
				AllowError: true,
			},
		),
	}
}

func try(arg ref.Val) ref.Val {
	if types.IsError(arg) {
		return types.String(fmt.Sprint(arg))
	}
	return arg
}

func tryMessage(arg, msg ref.Val) ref.Val {
	str, ok := msg.(types.String)
	if !ok {
		return types.NoSuchOverloadErr()
	}
	if types.IsError(arg) {
		return types.NewRefValMap(types.DefaultTypeAdapter, map[ref.Val]ref.Val{
			str: types.String(fmt.Sprint(arg)),
		})
	}
	return arg
}

func isError(arg ref.Val) ref.Val {
	return types.Bool(types.IsError(arg))
}
