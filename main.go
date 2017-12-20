package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/dinedal/textql/outputs"
	"github.com/dinedal/textql/storage"
	"github.com/k4yl3x/logql/config"
	myInput "github.com/k4yl3x/logql/inputs"
	myOutput "github.com/k4yl3x/logql/outputs"
	"github.com/mitchellh/go-homedir"
)

var cfg config.Config

// main
func main() {
	cfgPath, err := homedir.Expand("~/.logql.yml")
	if err != nil {
		panic(err)
	}
	configYaml, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		panic(err)
	}

	cfg = config.Config{}
	if err := yaml.Unmarshal(configYaml, &cfg); err != nil {
		log.Fatal(".logql.yml parse failed")
	}

	app := cli.NewApp()
	app.Name = "logql"
	app.Usage = ""
	app.Version = "1.0.0"
	app.Author = "k4yl3x"
	cli.AppHelpTemplate = `USAGE:
   {{.Name}} [options] [arguments...]

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:
  {{.Author}}{{end}}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

	keys := make([]string, len(cfg.Parse.Types))

	i := 0
	for k := range cfg.Parse.Types {
		keys[i] = k
		i++
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "enable_raw_line, r",
			Usage: "includes original line",
		},
		cli.BoolFlag{
			Name:  "output_tsv, T",
			Usage: "outputs with tsv format",
		},
		cli.BoolFlag{
			Name:  "repeat_table_header",
			Usage: "show table header each 30 rows",
		},
		cli.StringFlag{
			Name:  "query, q",
			Usage: "SQL query",
		},
		cli.StringFlag{
			Name:  "type,t",
			Usage: "log types(" + strings.Join(keys, ", ") + ")",
		},
		cli.StringFlag{
			Name:  "analysis, a",
			Usage: "pair of query and log type defined at .logql.yml.",
		},
	}

	app.Action = action
	app.Run(os.Args)
}

func action(c *cli.Context) error {
	analysis := c.GlobalString("analysis")
	var logType string
	var sqlQuery string

	if analysis == "" {
		logType = c.GlobalString("type")
		sqlQuery = c.GlobalString("query")
	} else {
		for _, v := range cfg.Query.AnalysisSets {
			a := strings.Split(analysis, ",")
			if v.Label == a[0] {
				logType = v.Type
				sqlQuery = v.Query
			}
			for i := 1; i < len(a); i++ {
				pat := fmt.Sprintf("%%arg%d%%", i)
				sqlQuery = regexp.MustCompile(pat).ReplaceAllString(sqlQuery, a[i])
			}
			sqlQuery = regexp.MustCompile("%arg[0-9]+%").ReplaceAllString(sqlQuery, "")
		}
		if logType == "" || sqlQuery == "" {
			log.Fatalf("invalid analysis => '%s'\n", analysis)
		}

	}

	parserConfig, ok := cfg.Parse.Types[logType]
	if !ok {
		log.Fatalf("invalid log type => '%s'\n", logType)
	}

	yio := myInput.YaInputOptions{
		EnableRawLine: false,
		Config:        &parserConfig,
		ReadFrom:      bufio.NewScanner(os.Stdin),
		Timezone:      time.FixedZone(cfg.Parse.Global.Timezone.Name, cfg.Parse.Global.Timezone.Offset),
	}
	input, err := myInput.NewYaInput(&yio)
	if err != nil {
		log.Fatalf("myInput.NewYaInput failed: %v", err)
	}

	s := storage.NewSQLite3StorageWithDefaults()
	defer func() {
		s.Close()
	}()
	s.LoadInput(input)

	var output outputs.Output
	if c.GlobalBool("output_tsv") {
		displayOpts := &outputs.CSVOutputOptions{
			WriteHeader: true,
			Separator:   rune('\t'),
			WriteTo:     os.Stdout,
		}

		output = outputs.NewCSVOutput(displayOpts)
	} else {
		displayOpts := &myOutput.PrettyTableOutputOptions{
			WriteHeader:  true,
			WriteTo:      os.Stdout,
			RepeatHeader: c.GlobalBool("repeat_table_header"),
		}
		output = myOutput.NewPrettyTableOutput(displayOpts)
	}

	if sqlQuery == "" {
		sqlQuery = "select * from stdin"
	}

	for _, alias := range cfg.Query.Aliases {
		sqlQuery = regexp.MustCompile(alias.Regexp).ReplaceAllString(sqlQuery, alias.Replacement)
	}

	queryResults, queryErr := s.ExecuteSQLString(sqlQuery)

	if queryErr != nil {
		log.Fatalf("query error: %v", queryErr)
	}

	if queryResults != nil {
		if c.GlobalBool("output_tsv") {
			fmt.Print("#")
		}
		output.Show(queryResults)
	}

	return nil
}
