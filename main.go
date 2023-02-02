package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
)

func main() {
	decode := false
	rust := false

	flag.Usage = func() {
		fmt.Print(`usage: bytes [-d|--decode] [--rust] <file>?

bytes formats binary input as a []byte{} array for use in go code, or a vec![] for rust.
If no file name is provided, bytes reads from stdin
If -d or --decode is provided bf converts from the source-code representation back to binary

bytes can also be used to extract the input from a go fuzz file if the fuzz
test took bytes as input.
`)
		flag.PrintDefaults()
	}
	flag.BoolVar(&decode, "decode", false, "output formatted bytes as binary")
	flag.BoolVar(&decode, "d", false, "")
	flag.BoolVar(&rust, "rust", false, "")
	flag.Parse()

	var input []byte
	var err error

	if len(flag.Args()) == 0 {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			fmt.Fprintln(os.Stderr, "Reading from stdin... (ctrl+d when done)")
		}
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(flag.Args()[0])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if decode {
		doDecode(input)
		return
	}

	if rust {
		fmt.Printf("vec![")
		for i, b := range input {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Print(b)
		}
		fmt.Printf("]\n")
		return
	}
	fmt.Printf("%#v\n", input)
}

// Based on code from https://tip.golang.org/src/internal/fuzz/encoding.go
func doDecode(input []byte) {
	input = bytes.Trim(input, " \t\n;")
	input = bytes.TrimPrefix(input, []byte("go test fuzz v1\n"))
	if bytes.HasPrefix(input, []byte("vec![")) {
		input = bytes.Replace(input, []byte("]"), []byte("}"), -1)
		input = bytes.Replace(input, []byte("vec!["), []byte("[]byte{"), -1)
	}
	fs := token.NewFileSet()
	expr, err := parser.ParseExprFrom(fs, "(test)", input, 0)
	if err != nil {
		parseErr()
	}
	composite, ok := expr.(*ast.CompositeLit)
	var output []byte
	if ok {
		if !isByteSlice(composite.Type) {
			parseErr()
		}

		output = make([]byte, len(composite.Elts))
		for i, elt := range composite.Elts {
			bl, ok := elt.(*ast.BasicLit)
			if !ok {
				parseErr()
			}
			b, err := strconv.ParseInt(bl.Value, 0, 8)
			if err != nil {
				sb, err := strconv.ParseUint(bl.Value, 0, 8)
				if err != nil {
					parseErr()
				}
				output[i] = byte(sb)
			} else {
				output[i] = byte(b)
			}
		}

	} else {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			parseErr()
		}
		if len(call.Args) != 1 {
			parseErr()
		}

		if !isByteSlice(call.Fun) {
			parseErr()
		}

		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			parseErr()
		}
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			parseErr()
		}
		output = []byte(s)
	}

	os.Stdout.Write(output)
}

func isByteSlice(t ast.Expr) bool {
	arrayType, ok := t.(*ast.ArrayType)
	if !ok {
		return false
	}
	if arrayType.Len != nil {
		return false
	}
	elt, ok := arrayType.Elt.(*ast.Ident)
	if !ok || elt.Name != "byte" {
		return false
	}
	return true
}

func parseErr() {
	fmt.Fprintln(os.Stderr, "error: expected input to match []byte{1,2,3} or vec![1,2,3]")
	os.Exit(1)
}
