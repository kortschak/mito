// Package mito provides the logic for a main function and test infrastructure
// for a CEL-based message stream processor.
//
// This repository is a design sketch. The majority of the logic resides in the
// the lib package.
package mito

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/interpreter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/kortschak/mito/lib"
)

const root = "data"

func Main() int {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of %s:

  %[1]s [opts] <src.cel>

`, os.Args[0])
		flag.PrintDefaults()
	}
	use := flag.String("use", "all", "libraries to use")
	data := flag.String("data", "", "path to a JSON object holding input (exposed as the label "+root+")")
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		return 2
	}

	var libs []cel.EnvOption
	if *use == "all" {
		for _, l := range libMap {
			libs = append(libs, l)
		}
	} else {
		for _, u := range strings.Split(*use, ",") {
			libs = append(libs, libMap[u])
		}
	}
	b, err := os.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	var input interface{}
	if *data != "" {
		b, err := os.ReadFile(*data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		err = json.Unmarshal(b, &input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		input = map[string]interface{}{root: input}
	}

	res, err := eval(string(b), root, input, libs...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(res)
	return 0
}

var (
	libMap = map[string]cel.EnvOption{
		"collections": lib.Collections(),
		"crypto":      lib.Crypto(),
		"json":        lib.JSON(nil),
		"time":        lib.Time(),
		"try":         lib.Try(),
		"file":        lib.File(mimetypes),
		"mime":        lib.MIME(mimetypes),
		"http":        lib.HTTP(),
	}

	mimetypes = map[string]interface{}{
		"text/rot13":           func(r io.Reader) io.Reader { return rot13{r} },
		"text/upper":           toUpper,
		"application/gzip":     func(r io.Reader) (io.Reader, error) { return gzip.NewReader(r) },
		"application/x-ndjson": lib.NDJSON,
	}
)

func eval(src, root string, input interface{}, libs ...cel.EnvOption) (string, error) {
	opts := append([]cel.EnvOption{
		cel.Declarations(decls.NewVar(root, decls.Dyn)),
	}, libs...)
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create env: %v", err)
	}

	ast, iss := env.Compile(src)
	if iss.Err() != nil {
		return "", fmt.Errorf("failed compilation: %v", iss.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return "", fmt.Errorf("failed program instantiation: %v", err)
	}

	if input == nil {
		input = interpreter.EmptyActivation()
	}
	out, _, err := prg.Eval(input)
	if err != nil {
		return "", fmt.Errorf("failed eval: %v", err)
	}

	v, err := out.ConvertToNative(reflect.TypeOf(&structpb.Value{}))
	if err != nil {
		return "", fmt.Errorf("failed proto conversion: %v", err)
	}
	b, err := protojson.MarshalOptions{Indent: "\t"}.Marshal(v.(proto.Message))
	if err != nil {
		return "", fmt.Errorf("failed native conversion: %v", err)
	}
	var res interface{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return "", fmt.Errorf("failed json conversion: %v", err)
	}
	b, err = json.MarshalIndent(res, "", "\t")
	return string(b), err
}

// rot13 is provided for testing purposes.
type rot13 struct {
	r io.Reader
}

func (r rot13) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	for i, b := range p[:n] {
		var base byte
		switch {
		case 'A' <= b && b <= 'Z':
			base = 'A'
		case 'a' <= b && b <= 'z':
			base = 'a'
		default:
			continue
		}
		p[i] = ((b - base + 13) % 26) + base
	}
	return n, err
}

func toUpper(p []byte) {
	for i, b := range p {
		if 'a' <= b && b <= 'z' {
			p[i] &^= 'a' - 'A'
		}
	}
}
