//go:build wasm
// +build wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/shellyln/go-open-soql-parser/soql/parser"
)

func doPanic(this js.Value, args []js.Value) (retval interface{}) {
	defer func() {
		if r := recover(); r != nil {
			retval = fmt.Sprintf("%v", r)
		}
	}()
	panic("error!")
}

func parseSoql(this js.Value, args []js.Value) interface{} {
	src := ""
	if 0 < len(args) {
		src = args[0].String()
	}

	var jsonStr []byte

	parsedQuery, err := parser.Parse(src)
	if err != nil {
		return js.ValueOf(err.Error())
	}

	jsonStr, err = json.Marshal(parsedQuery)
	if err != nil {
		return js.ValueOf(err.Error())
	}

	var buf bytes.Buffer
	json.Indent(&buf, jsonStr, "", "  ")
	return js.ValueOf(buf.String())
}

func main() {
	println("Go WebAssembly Initialized")

	js.Global().Set("parseSoql", js.FuncOf(parseSoql))

	select {}
}
