package main

import "fmt"

func (p *Parser) parseCriteria() *Criteria {
	var result Criteria

	// 1. Criteria keyword
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseCriteria - 1",
			KindExpected: tokenCriteria,
		},
	) {
		return nil
	}

	// 2. Criteria name (string)
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseCriteria - 2",
			KindExpected: tokenStringLiteral,
		},
	) {
		return nil
	}

	result.Name = p.tokenCurrent.valueLiteral
	p.advanceToken()

	// 3. Opening brace
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseCriteria - 3",
			KindExpected: tokenMonitor,
		},
	) {
		return nil
	}

	// 4. Body parsing (improved keyword detection)
	for !p.currentTokenIs(tokenRightBrace) {
		switch {
		case p.currentTokenIs(tokenIdentifier):
			switch p.tokenCurrent.valueLiteral {
			case "baseline", "increment":
				if setting := p.parseSetting(); setting != nil {
					result.Settings = append(
						result.Settings,
						setting,
					)

					continue
				}

			case _dslRightBrace:
				continue

			case _dslMonitor:
				if monitor := p.parseMonitor(); monitor != nil {
					result.Monitors = append(
						result.Monitors,
						monitor,
					)

					continue
				}
			}

			p.errorf(
				"Caller:%s\nUnexpected identifier: %s",

				"parseCriteria - 4",
				p.tokenCurrent.valueLiteral,
			)

		default:
			fmt.Println(
				"xxxxxxxxxxxxx",
				p.tokenCurrent.valueLiteral,
			)

			p.errorf(
				"Caller:%s\nUnexpected token: %v",

				"parseCriteria - 5",
				p.tokenCurrent.kind,
			)
		}

		p.skipToIdentifierRightBrace(
			"baseline",
			"increment",
			_dslMonitor,
		)
	}

	// 5. Closing brace
	if !p.expectWTokenAdvance(
		&paramsExpect{
			Caller:       "parseCriteria - 6",
			KindExpected: tokenRightBrace,
		},
	) {
		p.tryRecoverAtBlockEnd()

		return nil
	}

	return &result
}
