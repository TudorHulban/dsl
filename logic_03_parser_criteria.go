package main

func (p *Parser) parseCriteria() *Criteria {
	var result Criteria

	if !p.expectIdentifier(_dslCriteria) {
		return nil
	}

	if p.tokenCurrent.kind != tokenStringLiteral {
		p.errorf(
			"expected criteria name string, got %v",
			p.tokenCurrent.kind,
		)

		return nil
	}

	result.Name = p.tokenCurrent.valueLiteral
	p.advanceToken()

	if !p.expect(tokenLeftBrace) {
		return nil
	}

	for p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF && p.tokenCurrent.kind != tokenError {
		switch {
		case p.tokenCurrent.kind == tokenIdentifier && (p.tokenCurrent.valueLiteral == "baseline" || p.tokenCurrent.valueLiteral == "increment"):
			sett := p.parseSetting()
			if sett != nil {
				result.Settings = append(result.Settings, sett)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skipToIdentifierRightBrace("baseline", "increment", "monitor")
			}

		case p.tokenCurrent.kind == tokenIdentifier && p.tokenCurrent.valueLiteral == "monitor":
			mon := p.parseMonitor()
			if mon != nil {
				result.Monitors = append(result.Monitors, mon)
			} else {
				// error recovery: skip to next setting/monitor or '}'
				p.skipToIdentifierRightBrace("baseline", "increment", "monitor")
			}

		default:
			p.errorf("unexpected token inside criteria block: %v (%s)", p.tokenCurrent.kind, p.tokenCurrent.valueLiteral)
			p.skipToIdentifierRightBrace("baseline", "increment", "monitor")
		}
	}

	if !p.expect(tokenRightBrace) {
		p.errorf("criteria block not properly closed")

		if p.tokenCurrent.kind != tokenRightBrace && p.tokenCurrent.kind != tokenEOF {
			p.advanceToken()
		}
	}

	return &result
}
