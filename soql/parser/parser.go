// Open source implementation of the SOQL parser.
package parser

import (
	"errors"
	"strconv"
	"time"

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
	meta := &types.SoqlQueryMeta{
		Version: "0.9",
		Date:    time.Now().UTC(),
		Source:  s,
	}

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

	q.Meta = meta

	if err := postprocess.Normalize(&q); err != nil {
		return nil, err
	}

	endDate := time.Now()
	q.Meta.ElapsedTime = endDate.Sub(q.Meta.Date)

	return &q, nil
}
