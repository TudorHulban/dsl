package dslalert

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
	if !p.expectNoTokenAdvance(
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
			KindExpected: tokenLeftBrace,
		},
	) {
		return nil
	}

	// 4. Check for empty block
	if p.currentTokenIs(tokenRightBrace) {
		p.errorf(
			"Caller:%s\nEmpty criteria block not allowed",
			"parseCriteria - 4",
		)

		p.advanceToken() // Consume the '}' to avoid hanging

		return nil
	}

	// 5. Body parsing (improved keyword detection)
	for !p.currentTokenIs(tokenRightBrace) {
		switch p.tokenCurrent.kind { // Switch on kind, not valueLiteral
		case tokenIdentifier:
			switch p.tokenCurrent.valueLiteral {
			default:
				p.errorf(
					"Caller:%s\nUnexpected identifier: %s",
					"parseCriteria - 4",
					p.tokenCurrent.valueLiteral,
				)
			}

		case tokenMonitor:
			if monitor := p.parseMonitor(); monitor != nil {
				result.Monitors = append(result.Monitors, monitor)

				continue
			}

		default:
			p.errorf(
				"Caller:%s\nUnexpected token: %v",
				"parseCriteria - 5",
				p.tokenCurrent.kind,
			)
		}

		p.skipToIdentifierRightBrace("baseline", "increment", _dslMonitor)
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
