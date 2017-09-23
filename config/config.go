package config

import ()

// Config is ..
type Config struct {
	Query QueryConfig `yaml:"query"`
	Parse ParseConfig `yaml:"parse"`
}

// QueryConfig is ...
type QueryConfig struct {
	Aliases      []QueryAliasConfig `yaml:"aliases"`
	AnalysisSets []AnalysisSet      `yaml:"analysis_set"`
}

// QueryAliasConfig is ...
type QueryAliasConfig struct {
	Regexp      string `yaml:"regexp"`
	Replacement string `yaml:"replacement"`
}

// AnalysisSet is ...
type AnalysisSet struct {
	Label string `yaml:"label"`
	Type  string `yaml:"type"`
	Query string `yaml:"query"`
}

// ParseConfig is ...
type ParseConfig struct {
	Global GlobalConfig                   `yaml:"global"`
	Types  map[string](StringParseConfig) `yaml:"log_types"`
}

// GlobalConfig is ...
type GlobalConfig struct {
	Timezone TimezoneConfig `yaml:"timezone"`
}

// StringParseConfig is ...
type StringParseConfig struct {
	StringGroupingRules []LineBlockRule         `yaml:"string_grouping_rules"`
	Delims              []string                `yaml:"delimiters"`
	SkipDelimiterRepeat bool                    `yaml:"skip_delimiter_repeat"`
	TrimBolDelimiters   bool                    `yaml:"trim_bol_delimiters"`
	SkipBolWiths        []string                `yaml:"skip_bol_withs"`
	Columns             ColumnSlice             `yaml:"columns"`
	DivideColumns       map[string]DivideConfig `yaml:"divide_columns"`
	JoinColumns         map[string]JoinConfig   `yaml:"join_columns"`
	DropColumns         []string                `yaml:"drop_columns"`
	TimeColumns         []TimeColumnConfig      `yaml:"time_columns"`
}

// DivideConfig is ...
type DivideConfig struct {
	Columns   []string `yaml:"columns"`
	Delimiter string   `yaml:"delimiter"`
}

// JoinConfig is ...
type JoinConfig struct {
	Columns   []string `yaml:"columns"`
	Delimiter string   `yaml:"delimiter"`
}

// LineBlockRule is ...
type LineBlockRule struct {
	StartWith string `yaml:"start_with"`
	EndWith   string `yaml:"end_with"`
}

// TimeColumnConfig is ...
type TimeColumnConfig struct {
	ColumnName string         `yaml:"column_name"`
	Format     string         `yaml:"format"`
	Timezone   TimezoneConfig `yaml:"timezone"`
}

// TimezoneConfig is ...
type TimezoneConfig struct {
	Name   string `yaml:"name"`
	Offset int    `yaml:"offset"`
}

// ColumnSlice is ...
type ColumnSlice []string

// IndexOf is ...
func (c ColumnSlice) IndexOf(s string) int {
	for i := 0; i < len(c); i++ {
		if s == c[i] {
			return i
		}
	}
	return -1
}
