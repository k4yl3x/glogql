package input

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/k4yl3x/logql/config"
	_parser "github.com/k4yl3x/logql/parser"
)

// YaInput is ...
type YaInput struct {
	options              *YaInputOptions
	header               []string
	name                 string
	timezones            map[int]*time.Location
	parser               _parser.Parser
	divideColumns        map[int]config.DivideConfig
	columnBuilders       []columnBuilder
	timeColumnConverters map[int]timeColumnConverter
}

// YaInputOptions is ...
type YaInputOptions struct {
	EnableRawLine bool
	Config        *config.StringParseConfig
	ReadFrom      *bufio.Scanner
	Timezone      *time.Location
}

// NewYaInput is ...
func NewYaInput(yio *YaInputOptions) (yi *YaInput, err error) {
	timezones := make(map[int]*time.Location, len(yio.Config.TimeColumns))

	for _, tc := range yio.Config.TimeColumns {
		tz := time.FixedZone(tc.Timezone.Name, tc.Timezone.Offset)
		timezones[yio.Config.Columns.IndexOf(tc.ColumnName)] = tz
	}

	p, err := _parser.NewLineParser(yio.Config)
	if err != nil {
		panic(err)
	}

	yi = &YaInput{
		options:   yio,
		name:      "stdin",
		timezones: timezones,
		parser:    p,
	}
	yi.init()
	return
}

type columnBuilder interface {
	Build(columns []string) string
}
type standardColumnBuilder struct {
	sourceColumnIndex int
}

func (b standardColumnBuilder) Build(columns []string) string {
	return columns[b.sourceColumnIndex]
}

type joinColumnBuilder struct {
	sourceColumnIndexes []int
	delimiter           string
}

func (b joinColumnBuilder) Build(columns []string) string {
	t := make([]string, 0, len(b.sourceColumnIndexes))
	for _, i := range b.sourceColumnIndexes {
		t = append(t, columns[i])
	}
	return strings.Join(t, b.delimiter)
}

type timeColumnConverter struct {
	sourceFormat          string
	sourceTimezone        *time.Location
	outputFormat          string
	outputTimezone        *time.Location
	sourceTimezoneConvert bool
}

func (c timeColumnConverter) Convert(column string) string {
	var t time.Time
	var err error
	if c.sourceTimezoneConvert {
		t, err = time.ParseInLocation(c.sourceFormat, column, c.sourceTimezone)
	} else {
		t, err = time.Parse(c.sourceFormat, column)
	}
	if err == nil {
		return t.In(c.outputTimezone).Format(c.outputFormat)
	}
	return column
}

// Name is ...
func (yi *YaInput) Name() string {
	return yi.name
}

// SetName sets dataset name
func (yi *YaInput) SetName(name string) {
	yi.name = name
}

// ReadRecord is
func (yi *YaInput) ReadRecord() (rows []string) {
	if yi.options.ReadFrom.Scan() {
		line := strings.TrimRight(yi.options.ReadFrom.Text(), "\n")

		for _, p := range yi.options.Config.SkipBolWiths {
			if strings.HasPrefix(line, p) {
				return yi.ReadRecord()
			}
		}

		r, err := yi.parser.Parse(line)
		if err != nil {
			log.Fatalf("yi.parser.Parse faild %v", err)
		}

		for i, v := range yi.divideColumns {
				s := strings.SplitN(r[i], v.Delimiter, len(v.Columns))
				for j := 0; j < len(v.Columns); j++ {
					if len(s) > j {
						r = append(r, s[j])
					} else {
						r = append(r, "")
					}
				}
			}
		}

		for i, cb := range yi.columnBuilders {
			if tc, ok := yi.timeColumnConverters[i]; ok {
				rows = append(rows, tc.Convert(cb.Build(r)))
			} else {
				rows = append(rows, cb.Build(r))
			}
		}

	}
	return
}

// Header is ...
func (yi *YaInput) Header() (header []string) {
	return yi.header

}

func (yi *YaInput) init() {
	header := make([]string, 0, len(yi.options.Config.Columns))
	ht := make([]string, 0, len(yi.options.Config.Columns))
	divideColumns := make(map[int]config.DivideConfig, 0)

	// parsed and devided column name to index map
	workingColumnNames := make(map[string]int, len(yi.options.Config.Columns))
	columnBuilders := make(map[string]columnBuilder)

	// column name to column index
	for i := 0; i < len(yi.options.Config.Columns); i++ {
		workingColumnNames[yi.options.Config.Columns[i]] = i
	}

	// drop original columns when marked as dropping or using for join
	excludeColumnIndexes := make(map[int]struct{}, 0)

	// the joined new column index to column name map
	newColumns := make(map[int]string, 0)

	for _, v := range yi.options.Config.DivideColumns {
		for i := 0; i < len(v.Columns); i++ {
			// append new names and indexes about divided columns
			workingColumnNames[v.Columns[i]] = len(workingColumnNames)
		}
	}

	for k, v := range yi.options.Config.JoinColumns {
		// append new names and indexes about joined columns
		workingColumnNames[k] = len(workingColumnNames)

		// join source columns indexes
		cs := make([]int, 0, len(v.Columns))

		// minimam join column index(for new joined column index)
		m := workingColumnNames[v.Columns[0]]
		for _, k2 := range v.Columns {

			if m > workingColumnNames[k2] {
				m = workingColumnNames[k2]
			}

			// drop the original join columns
			var s struct{}
			excludeColumnIndexes[workingColumnNames[k2]] = s
			cs = append(cs, workingColumnNames[k2])
		}
		newColumns[m] = k
		columnBuilders[k] = joinColumnBuilder{
			sourceColumnIndexes: cs,
			delimiter:           v.Delimiter,
		}
	}

	for _, k := range yi.options.Config.DropColumns {
		var s struct{}
		excludeColumnIndexes[workingColumnNames[k]] = s
	}

	for i := 0; i < len(yi.options.Config.Columns); i++ {
		if _, ok := newColumns[i]; ok {
			ht = append(ht, newColumns[i])
		} else if _, ok := excludeColumnIndexes[i]; ok {
			continue
		} else {
			ht = append(ht, yi.options.Config.Columns[i])
		}
	}

	for k, i := range workingColumnNames {
		if _, ok := excludeColumnIndexes[i]; !ok {
			if _, ok := columnBuilders[k]; !ok {
				columnBuilders[k] = standardColumnBuilder{
					sourceColumnIndex: i,
				}
			}
		}
	}

	cb := make([]columnBuilder, 0, len(columnBuilders))
	for i := range ht {
		if v, ok := yi.options.Config.DivideColumns[ht[i]]; ok {
			divideColumns[workingColumnNames[ht[i]]] = v
			for j := 0; j < len(v.Columns); j++ {
				header = append(header, v.Columns[j])
				cb = append(cb, columnBuilders[v.Columns[j]])
			}
		} else {
			header = append(header, ht[i])
			cb = append(cb, columnBuilders[ht[i]])
		}
	}

	yi.header = header
	yi.divideColumns = divideColumns
	yi.columnBuilders = cb

	tcs := make(map[int]timeColumnConverter)
	for _, tc := range yi.options.Config.TimeColumns {
		i := 0
		for j, n := range header {
			if tc.ColumnName == n {
				i = j
				break
			}
		}
		tcs[i] = timeColumnConverter{
			sourceFormat:          tc.Format,
			sourceTimezone:        time.FixedZone(tc.Timezone.Name, tc.Timezone.Offset),
			outputFormat:          time.RFC3339Nano,
			outputTimezone:        yi.options.Timezone,
			sourceTimezoneConvert: tc.Timezone.Name != "" && tc.Timezone.Offset != 0,
		}
	}

	yi.timeColumnConverters = tcs

	return
}

func min(a []int) (min int, err error) {
	if len(a) == 0 {
		err = fmt.Errorf("given array length is 0")
	}
	min = a[0]
	for i := 1; i < len(a); i++ {
		if min > a[i] {
			min = a[i]
		}
	}
	return
}
