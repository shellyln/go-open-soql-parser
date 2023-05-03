# go-open-soql-parser
Open source implementation of the SOQL parser for Go.

[![Test](https://github.com/shellyln/go-open-soql-parser/actions/workflows/test.yml/badge.svg)](https://github.com/shellyln/go-open-soql-parser/actions/workflows/test.yml)
[![release](https://img.shields.io/github/v/release/shellyln/go-open-soql-parser)](https://github.com/shellyln/go-open-soql-parser/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/shellyln/go-open-soql-parser)](https://github.com/shellyln/go-open-soql-parser)

<img src="https://raw.githubusercontent.com/shellyln/go-open-soql-parser/master/_assets/logo-opensoql.svg" alt="logo" style="width:250px;" width="250">

---

## 🧭 Examples

* [Live demo](https://shellyln.github.io/soql/)
* [SOQL query visualizer](https://shellyln.github.io/soql-visualizer/)

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

## 🚧 TODO
* Unit tests
* `GROUP BY ROLLUP` and `GROUP BY CUBE` clause, `GROUPING()` function
* `WITH` clause
* `USING SCOPE` clause
* Formula in fieldExpression at conditionExpression (where / having)
* Polymorphic Fields
* "null Values in Lookup Relationships and Outer Joins" - If an object has a conditional expression whose right hand side is null, it is not a condition for inner join.
    * cf. "Using Relationship Queries" - If the condition is complete within the parent object (no "or" across relationships), it is inner joined.

## 🔗 Related projects
* [GraphDT](https://github.com/shellyln/go-graphdt) - Datatable
* [go-open-soql-visualizer](https://github.com/shellyln/go-open-soql-visualizer) - Visualizer
* [Open SOQL (JavaScript)](https://github.com/shellyln/open-soql) - JavaScript implementation

## 🙋 FAQ
* What does `SOQL` stand for?
  * 👉 In `Open SOQL`, `SOQL` stands for `SOQL is Object Query Language`.
  * 👉 In [original SOQL](https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql.htm), `SOQL` stands for `Salesforce Object Query Language`.

## ⚖️ License

MIT  
Copyright (c) 2023 Shellyl_N and Authors.
