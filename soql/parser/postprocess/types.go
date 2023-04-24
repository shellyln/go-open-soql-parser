package postprocess

import (
	. "github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

type soqlConditionStack struct {
	cond   SoqlCondition
	srcPos int
	minPos int
}

type soqlQueryPlace int

const (
	soqlQueryPlace_Primary soqlQueryPlace = iota + 1
	soqlQueryPlace_Select
	soqlQueryPlace_ConditionalOperand
)
