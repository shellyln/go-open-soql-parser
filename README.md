# go-open-soql-parser
Open source implementation of the SOQL parser for Go.

[![Test](https://github.com/shellyln/go-open-soql-parser/actions/workflows/test.yml/badge.svg)](https://github.com/shellyln/go-open-soql-parser/actions/workflows/test.yml)
[![release](https://img.shields.io/github/v/release/shellyln/go-open-soql-parser)](https://github.com/shellyln/go-open-soql-parser/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/shellyln/go-open-soql-parser)](https://github.com/shellyln/go-open-soql-parser)

<img src="https://raw.githubusercontent.com/shellyln/go-open-soql-parser/master/_assets/logo-opensoql.svg" alt="logo" style="width:250px;" width="250">

---

## 🧭 Examples

* [Live demo](https://shellyln.github.io/soql/)

## 🚀 Getting started

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/shellyln/go-open-soql-parser/soql/parser"
)

func main() {
    src := `
    SELECT Id FROM Contact WHERE Name like 'a%'
    `

    ret, err := parser.Parse(src)
    if err != nil {
        fmt.Println(err)
    }

    jsonStr, err := json.Marshal(ret)
    if err != nil {
        println(err)
    }

    var buf bytes.Buffer
    json.Indent(&buf, jsonStr, "", "  ")

    fmt.Println(buf.String())
}
```

## ⚖️ License

MIT  
Copyright (c) 2023 Shellyl_N and Authors.
