package parser

import (
	"errors"
	"strconv"

	"github.com/shellyln/go-open-soql-parser/soql/parser/core"
	"github.com/shellyln/go-open-soql-parser/soql/parser/postprocess"
	"github.com/shellyln/go-open-soql-parser/soql/parser/types"
	. "github.com/shellyln/takenoco/base"
	. "github.com/shellyln/takenoco/string"
)

var (
	queryParser ParserFn
)

func init() {
	queryParser = core.Query()
}

func Parse(s string) (*types.SoqlQuery, error) {
	out, err := queryParser(*NewStringParserContext(s))
	if err != nil {
		pos := GetLineAndColPosition(s, out.SourcePosition, 4)
		return nil, errors.New(
			err.Error() +
				"\n --> Line " + strconv.Itoa(pos.Line) +
				", Col " + strconv.Itoa(pos.Col) + "\n" +
				pos.ErrSource)
	}

	if out.MatchStatus != MatchStatus_Matched {
		pos := GetLineAndColPosition(s, out.SourcePosition, 4)
		return nil, errors.New(
			"Parse failed" +
				"\n --> Line " + strconv.Itoa(pos.Line) +
				", Col " + strconv.Itoa(pos.Col) + "\n" +
				pos.ErrSource)
	}

	q := out.AstStack[0].Value.(types.SoqlQuery)

	if err := postprocess.Normalize(&q); err != nil {
		return nil, err
	}

	return &q, nil
}
