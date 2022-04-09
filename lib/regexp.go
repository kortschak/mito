package lib

import (
	"regexp"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Regexp returns a cel.EnvOption to configure extended functions for
// using regular expressions on strings and bytes. It takes a mapping of
// names to Go regular expressions. The names are used to specify the pattern
// in the CEL regexp call.
//
// Each function corresponds to methods on regexp.Regexp in the Go standard
// library.
//
// For the examples below assume an input patterns map:
//
//     map[string]*regexp.Regexp{
//         "foo":     regexp.MustCompile("foo(.)"),
//         "foo_rep": regexp.MustCompile("(f)oo([ld])"),
//     }
//
// RE Match
//
// Returns whether the named pattern matches the receiver:
//
//     <bytes>.re_match(<string>) -> <bool>
//     <string>.re_match(<string>) -> <bool>
//
// Examples:
//
//     'food'.re_match('foo')    // return true
//     b'food'.re_match(b'foo')  // return true
//
//
// RE Find
//
// Returns a string or bytes of the named pattern's match:
//
//     <bytes>.re_find(<string>) -> <bytes>
//     <string>.re_find(<string>) -> <string>
//
// Examples:
//
//     'food'.re_find('foo')    // return "foo"
//     b'food'.re_find(b'foo')  // return "Zm9v"
//
//
// RE Find All
//
// Returns a list of strings or bytes of all the named pattern's matches:
//
//     <bytes>.re_find_all(<string>) -> <list<bytes>>
//     <string>.re_find_all(<string>) -> <list<string>>
//
// Examples:
//
//     'food fool'.re_find_all('foo')  // return ["food", "fool"]
//     b'food fool'.re_find_all(b'foo')  // return ["Zm9vZA==", "Zm9vZA=="]
//
//
// RE Find Submatch
//
// Returns a list of strings or bytes of the named pattern's submatches:
//
//     <bytes>.re_find_submatch(<string>) -> <list<bytes>>
//     <string>.re_find_submatch(<string>) -> <list<string>>
//
// Examples:
//
//     'food fool'.re_find_submatch('foo')   // return ["food", "d"]
//     b'food fool'.re_find_submatch('foo')  // return ["Zm9vZA==", "ZA=="]
//
//
// RE Find All Submatch
//
// Returns a list of lists of strings or bytes of all the named pattern's submatches:
//
//     <bytes>.re_find_all_submatch(<string>) -> <list<list<bytes>>>
//     <string>.re_find_all_submatch(<string>) -> <list<list<string>>>
//
// Examples:
//
//     'food fool'.re_find_all_submatch('foo')  // return [["food", "d"], ["fool", "l"]]
//
//
// RE Replace All
//
// Returns a strings or bytes applying a replacement to all matches of the named
// pattern:
//
//     <bytes>.re_replace_all(<string>, <bytes>) -> <bytes>
//     <string>.re_replace_all(<string>, <string>) -> <string>
//
// Examples:
//
//     'food fool'.re_replace_all('foo_rep', '${1}u${2}')    // return "fud ful"
//     b'food fool'.re_replace_all('foo_rep', b'${1}u${2}')  // return "ZnVkIGZ1bA=="
//
func Regexp(patterns map[string]*regexp.Regexp) cel.EnvOption {
	return cel.Lib(regexpLib(patterns))
}

type regexpLib map[string]*regexp.Regexp

func (l regexpLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Declarations(
			decls.NewFunction("re_match",
				decls.NewInstanceOverload(
					"bytes_re_match_string",
					[]*expr.Type{decls.Bytes, decls.String},
					decls.Bool,
				),
				decls.NewInstanceOverload(
					"string_re_match_string",
					[]*expr.Type{decls.String, decls.String},
					decls.Bool,
				),
			),
			decls.NewFunction("re_find",
				decls.NewInstanceOverload(
					"bytes_re_find_string",
					[]*expr.Type{decls.Bytes, decls.String},
					decls.Bytes,
				),
				decls.NewInstanceOverload(
					"string_re_find_string",
					[]*expr.Type{decls.String, decls.String},
					decls.String,
				),
			),
			decls.NewFunction("re_find_all",
				decls.NewInstanceOverload(
					"bytes_re_find_all_string",
					[]*expr.Type{decls.Bytes, decls.String},
					decls.NewListType(decls.Bytes),
				),
				decls.NewInstanceOverload(
					"string_re_find_all_string",
					[]*expr.Type{decls.String, decls.String},
					decls.NewListType(decls.String),
				),
			),
			decls.NewFunction("re_find_submatch",
				decls.NewInstanceOverload(
					"bytes_re_find_submatch_string",
					[]*expr.Type{decls.Bytes, decls.String},
					decls.NewListType(decls.Bytes),
				),
				decls.NewInstanceOverload(
					"string_re_find_submatch_string",
					[]*expr.Type{decls.String, decls.String},
					decls.NewListType(decls.String),
				),
			),
			decls.NewFunction("re_find_all_submatch",
				decls.NewInstanceOverload(
					"bytes_re_find_all_submatch_string",
					[]*expr.Type{decls.Bytes, decls.String},
					decls.NewListType(decls.NewListType(decls.Bytes)),
				),
				decls.NewInstanceOverload(
					"string_re_find_all_submatch_string",
					[]*expr.Type{decls.String, decls.String},
					decls.NewListType(decls.NewListType(decls.String)),
				),
			),
			decls.NewFunction("re_replace_all",
				decls.NewInstanceOverload(
					"bytes_re_replace_all_string_bytes",
					[]*expr.Type{decls.Bytes, decls.String, decls.Bytes},
					decls.Bytes,
				),
				decls.NewInstanceOverload(
					"string_re_replace_all_string_string",
					[]*expr.Type{decls.String, decls.String, decls.String},
					decls.String,
				),
			),
		),
	}
}

func (l regexpLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_re_match_string",
				Binary:   l.matchBytes,
			},
			&functions.Overload{
				Operator: "string_re_match_string",
				Binary:   l.matchString,
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_re_find_string",
				Binary:   l.findBytes,
			},
			&functions.Overload{
				Operator: "string_re_find_string",
				Binary:   l.findString,
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_re_find_all_string",
				Binary:   l.findAllBytes,
			},
			&functions.Overload{
				Operator: "string_re_find_all_string",
				Binary:   l.findAllString,
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_re_find_submatch_string",
				Binary:   l.findSubmatchBytes,
			},
			&functions.Overload{
				Operator: "string_re_find_submatch_string",
				Binary:   l.findSubmatchString,
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_re_find_all_submatch_string",
				Binary:   l.findAllSubmatchBytes,
			},
			&functions.Overload{
				Operator: "string_re_find_all_submatch_string",
				Binary:   l.findAllSubmatchString,
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_re_replace_all_string_bytes",
				Function: l.replaceAllBytes,
			},
			&functions.Overload{
				Operator: "string_re_replace_all_string_string",
				Function: l.replaceAllString,
			},
		),
	}
}

func (l regexpLib) matchBytes(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.Bytes)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.Bool(l[string(patName)].Match(src))
}

func (l regexpLib) matchString(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.String)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.Bool(l[string(patName)].MatchString(string(src)))
}

func (l regexpLib) findBytes(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.Bytes)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.Bytes(l[string(patName)].Find(src))
}

func (l regexpLib) findString(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.String)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.String(l[string(patName)].FindString(string(src)))
}

func (l regexpLib) findAllBytes(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.Bytes)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.DefaultTypeAdapter.NativeToValue(l[string(patName)].FindAll(src, -1))
}

func (l regexpLib) findAllString(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.String)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.DefaultTypeAdapter.NativeToValue(l[string(patName)].FindAllString(string(src), -1))
}

func (l regexpLib) findSubmatchBytes(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.Bytes)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.DefaultTypeAdapter.NativeToValue(l[string(patName)].FindSubmatch(src))
}

func (l regexpLib) findSubmatchString(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.String)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.DefaultTypeAdapter.NativeToValue(l[string(patName)].FindStringSubmatch(string(src)))
}

func (l regexpLib) findAllSubmatchBytes(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.Bytes)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.DefaultTypeAdapter.NativeToValue(l[string(patName)].FindAllSubmatch(src, -1))
}

func (l regexpLib) findAllSubmatchString(arg1, arg2 ref.Val) ref.Val {
	src, ok := arg1.(types.String)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := arg2.(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	return types.DefaultTypeAdapter.NativeToValue(l[string(patName)].FindAllStringSubmatch(string(src), -1))
}

func (l regexpLib) replaceAllBytes(args ...ref.Val) ref.Val {
	if len(args) != 3 {
		return types.NoSuchOverloadErr()
	}
	src, ok := args[0].(types.Bytes)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := args[1].(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	repl, ok := args[2].(types.Bytes)
	if !ok {
		return types.ValOrErr(repl, "no such overload")
	}
	return types.Bytes(l[string(patName)].ReplaceAll(src, repl))
}

func (l regexpLib) replaceAllString(args ...ref.Val) ref.Val {
	if len(args) != 3 {
		return types.NoSuchOverloadErr()
	}
	src, ok := args[0].(types.String)
	if !ok {
		return types.ValOrErr(src, "no such overload")
	}
	patName, ok := args[1].(types.String)
	if !ok {
		return types.ValOrErr(patName, "no such overload")
	}
	repl, ok := args[2].(types.String)
	if !ok {
		return types.ValOrErr(repl, "no such overload")
	}
	return types.String(l[string(patName)].ReplaceAllString(string(src), string(repl)))
}
