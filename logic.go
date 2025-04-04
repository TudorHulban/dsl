package dslalert

import (
	"bytes"
	"fmt"
	"io"
)

func Parse(input io.Reader) (*AlertConfiguration, []string) {
	buf := make([]byte, 1)
	n, err := input.Read(buf)
	if n == 0 || err == io.EOF {
		return nil,
			[]string{"input is empty"}
	}

	l := newLexer(
		io.MultiReader(bytes.NewReader(buf), input),
	)

	p := NewParser(
		&ParamsNewParser{
			Lexer: l,
		},
	)

	programAST, errParser := p.parserEntrypoint()

	if l.errorParsing != nil {
		p.errors = append(
			p.errors,
			fmt.Sprintf("lexer error: %v", l.errorParsing),
		)
	}

	if len(p.errors) > 0 {
		return nil,
			p.errors
	}

	if errParser != nil {
		return nil,
			p.errors
	}

	return programAST,
		nil
}
