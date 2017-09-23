package parser

import (
	"fmt"

	_config "github.com/k4yl3x/logql/config"
)

// LineParser is ...
type LineParser struct {
	stringGroupingRuleIndex      int
	stringGroupingRuleStartRunes []rune
	stringGroupingRuleEndRunes   []rune
	skipBolWiths                 [][]rune
	delimiters                   []rune
	skipDelimiterRepeat          bool
	trimBolDelimiters            bool
	columnNum                    int
}

// NewLineParser is ...
func NewLineParser(cfg *_config.StringParseConfig) (parser LineParser, err error) {
	var drs []rune
	var bsrs []rune
	var bers []rune
	var sbws [][]rune

	for i := 0; i < len(cfg.Delims); i++ {
		dr := []rune(cfg.Delims[i])
		if len(dr) != 1 {
			err = fmt.Errorf("Invalid delimiter string '%s'", cfg.Delims[i])
			return
		}
		drs = append(drs, dr[0])
	}
	// fmt.Printf("%v\n", drs)

	for i := 0; i < len(cfg.StringGroupingRules); i++ {
		bsws := []rune(cfg.StringGroupingRules[i].StartWith)
		bews := []rune(cfg.StringGroupingRules[i].EndWith)

		if len(bsws) != 1 || len(bews) != 1 {
			err = fmt.Errorf("Invalid line block rule")
		}
		bsrs = append(bsrs, bsws[0])
		bers = append(bers, bews[0])
	}
	// fmt.Printf("%v\n", brs)

	for i := 0; i < len(cfg.SkipBolWiths); i++ {
		sbws = append(sbws, []rune(cfg.SkipBolWiths[i]))
	}

	parser = LineParser{
		stringGroupingRuleIndex:      -1,
		stringGroupingRuleStartRunes: bsrs,
		stringGroupingRuleEndRunes:   bers,
		skipBolWiths:                 sbws,
		delimiters:                   drs,
		skipDelimiterRepeat:          cfg.SkipDelimiterRepeat,
		trimBolDelimiters:            cfg.TrimBolDelimiters,
		columnNum:                    len(cfg.Columns),
	}

	return
}

// Parse is ...
func (p LineParser) Parse(src string) (attrs []string, err error) {
	r := []rune(src + string(p.delimiters[0]))
	s := 0

	for i := 0; i < len(r); i++ {
		if p.trimBolDelimiters && s == 0 {
			for arrayIn(r[i], p.delimiters) >= 0 {
				i++
			}
			s = i
			continue
		}
		if p.stringGroupingRuleIndex >= 0 {
			for r[i] != p.stringGroupingRuleEndRunes[p.stringGroupingRuleIndex] {
				i++
			}
			p.stringGroupingRuleIndex = -1
			continue
		}

		if arrayIn(r[i], p.delimiters) >= 0 {
			if s >= 0 {
				// fmt.Println("> " + string(r[s:i]))
				attrs = append(attrs, string(r[s:i]))

				if len(attrs)+1 >= p.columnNum {
					attrs = append(attrs, string(r[i+1:len(r)-1]))
					break
				}
			}

			if len(r) <= i+1 {
				break
			}

			if p.skipDelimiterRepeat {
				for arrayIn(r[i+1], p.delimiters) >= 0 {
					i++
				}
			}
			s = i + 1

			if n := arrayIn(r[s], p.stringGroupingRuleStartRunes); n >= 0 {
				p.stringGroupingRuleIndex = n
				i++
			}
		}
	}

	return
}

func arrayIn(r rune, l []rune) int {
	for i := 0; i < len(l); i++ {
		if r == l[i] {
			return i
		}
	}
	return -1
}
