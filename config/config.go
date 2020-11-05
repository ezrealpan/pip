package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"ezreal.com.cn/pip/pip"
	"ezreal.com.cn/pip/pip/input"
	"ezreal.com.cn/pip/pip/models"
	"ezreal.com.cn/pip/pip/output"
	"ezreal.com.cn/pip/pip/parsers"
	"ezreal.com.cn/pip/pip/processors"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

var (
	// Default sections
	sectionDefaults = []string{"global_tags", "agent", "outputs",
		"processors", "aggregators", "inputs"}

	// Default input plugins
	inputDefaults = []string{"cpu", "mem", "swap", "system", "kernel",
		"processes", "disk", "diskio"}

	// Default output plugins
	outputDefaults = []string{"influxdb"}

	// envVarRe is a regex to find environment variables in the config file
	envVarRe = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)

	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
	)
)

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	Tags          map[string]string
	InputFilters  []string
	OutputFilters []string

	Inputs  []*models.RunningInput
	Outputs []*models.RunningOutput
	// Processors have a slice wrapper type because they need to be sorted
	Processors    models.RunningProcessors
	AggProcessors models.RunningProcessors
}

func NewConfig() *Config {
	c := &Config{
		// Agent defaults:

		Tags:          make(map[string]string),
		Inputs:        make([]*models.RunningInput, 0),
		Outputs:       make([]*models.RunningOutput, 0),
		Processors:    make([]*models.RunningProcessor, 0),
		AggProcessors: make([]*models.RunningProcessor, 0),
		InputFilters:  make([]string, 0),
		OutputFilters: make([]string, 0),
	}
	return c
}

// LoadConfig loads the given config file and applies it to c
func (c *Config) LoadConfig(path string) error {
	var err error
	data, err := loadConfig(path)
	if err != nil {
		return fmt.Errorf("Error loading config file %s: %w", path, err)
	}

	outor, ok := output.Outputs["simpleoutput"]
	if !ok {
		//TODO
	}
	out := outor()
	c.addOutput(out)

	printer, ok := processors.Processors["printer"]
	if !ok {
		//TODO
	}
	print := printer()
	c.addProcessor(print)

	if err = c.LoadConfigData(data); err != nil {
		return fmt.Errorf("Error loading config file %s: %w", path, err)
	}
	return nil
}
func (c *Config) addOutput(output pip.Output) error {
	rp := models.NewRunningOutput(output)
	c.Outputs = append(c.Outputs, rp)
	return nil
}

func (c *Config) addProcessor(processor pip.StreamingProcessor) error {
	rp := models.NewRunningProcessor(processor)
	c.Processors = append(c.Processors, rp)
	return nil
}

func loadConfig(config string) ([]byte, error) {
	return ioutil.ReadFile(config)

}

// LoadConfigData loads TOML-formatted config data
func (c *Config) LoadConfigData(data []byte) error {
	tbl, err := parseConfig(data)
	if err != nil {
		return fmt.Errorf("Error parsing data: %s", err)
	}

	// Parse tags tables first:
	for _, tableName := range []string{"tags", "global_tags"} {
		if val, ok := tbl.Fields[tableName]; ok {
			fmt.Println("val", val)
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("invalid configuration, bad table name %q", tableName)
			}
			if err = toml.UnmarshalTable(subTable, c.Tags); err != nil {
				return fmt.Errorf("error parsing table name %q: %w", tableName, err)
			}
		}
	}

	// Parse all the rest of the plugins:
	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("invalid configuration, error parsing field %q as table", name)
		}
		fmt.Println("name", name)
		fmt.Printf("subTable%+v", subTable)
		switch name {
		case "inputs", "plugins":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				// legacy [inputs.cpu] support
				case *ast.Table:
					fmt.Println("pluginName1", pluginName)
					fmt.Printf("pluginVal1%+v", pluginVal)
					if err = c.addInput(pluginName, pluginSubTable); err != nil {
						return fmt.Errorf("Error parsing %s, %s", pluginName, err)
					}
				case []*ast.Table:
					fmt.Println("pluginName2", pluginName)
					fmt.Printf("pluginVal2%+v", pluginVal)
					fmt.Println("----------------------")
					for _, t := range pluginSubTable {
						fmt.Printf("pluginVal2-t%+v", t)
						fmt.Println("----------------------")
						if err = c.addInput(pluginName, t); err != nil {
							return fmt.Errorf("Error parsing %s, %s", pluginName, err)
						}
					}
				default:
					fmt.Println("pluginName3", pluginName)
					fmt.Printf("pluginVal3%+v", pluginVal)
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
			}
		// Assume it's an input input for legacy config file support if no other
		// identifiers are present
		default:
			if err = c.addInput(name, subTable); err != nil {
				return fmt.Errorf("Error parsing %s, %s", name, err)
			}
		}
	}

	return nil
}

func (c *Config) addInput(name string, table *ast.Table) error {

	creator, ok := input.Inputs[name]
	if !ok {
		return fmt.Errorf("Undefined but requested input: %s", name)
	}
	input := creator()

	// If the input has a SetParser function, then this means it can accept
	// arbitrary types of input, so build the parser and set it.
	switch t := input.(type) {
	case parsers.ParserInput:
		parser, err := buildParser(name, table)
		if err != nil {
			return err
		}
		t.SetParser(parser)
	}

	switch t := input.(type) {
	case parsers.ParserFuncInput:
		config, err := getParserConfig(name, table)
		if err != nil {
			return err
		}
		t.SetParserFunc(func() (parsers.Parser, error) {
			return parsers.NewParser(config)
		})
	}

	pluginConfig, err := buildInput(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, input); err != nil {
		return err
	}

	rp := models.NewRunningInput(input, pluginConfig)
	rp.SetDefaultTags(c.Tags)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

// buildInput parses input specific items from the ast.Table,
// builds the filter and returns a
// models.InputConfig to be inserted into models.RunningInput
func buildInput(name string, tbl *ast.Table) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}

	cp.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := toml.UnmarshalTable(subtbl, cp.Tags); err != nil {
				return nil, fmt.Errorf("could not parse tags for input %s\n", name)
			}
		}
	}

	delete(tbl.Fields, "tags")
	return cp, nil
}

// parseConfig loads a TOML configuration from a provided path and
// returns the AST produced from the TOML parser. When loading the file, it
// will find environment variables and replace them.
func parseConfig(contents []byte) (*ast.Table, error) {
	contents = trimBOM(contents)

	parameters := envVarRe.FindAllSubmatch(contents, -1)
	for _, parameter := range parameters {
		if len(parameter) != 3 {
			continue
		}

		var env_var []byte
		if parameter[1] != nil {
			env_var = parameter[1]
		} else if parameter[2] != nil {
			env_var = parameter[2]
		} else {
			continue
		}

		env_val, ok := os.LookupEnv(strings.TrimPrefix(string(env_var), "$"))
		if ok {
			env_val = escapeEnv(env_val)
			contents = bytes.Replace(contents, parameter[0], []byte(env_val), 1)
		}
	}

	return toml.Parse(contents)
}

// trimBOM trims the Byte-Order-Marks from the beginning of the file.
// this is for Windows compatibility only.
// see https://github.com/influxdata/telegraf/issues/1378
func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}

// escapeEnv escapes a value for inserting into a TOML string.
func escapeEnv(value string) string {
	return envVarEscaper.Replace(value)
}

// buildParser grabs the necessary entries from the ast.Table for creating
// a parsers.Parser object, and creates it, which can then be added onto
// an Input object.
func buildParser(name string, tbl *ast.Table) (parsers.Parser, error) {
	config, err := getParserConfig(name, tbl)
	if err != nil {
		return nil, err
	}
	return parsers.NewParser(config)
}

func getParserConfig(name string, tbl *ast.Table) (*parsers.Config, error) {
	c := &parsers.Config{
		JSONStrict: true,
	}

	if node, ok := tbl.Fields["data_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DataFormat = str.Value
			}
		}
	}

	// Legacy support, exec plugin originally parsed JSON by default.
	if name == "exec" && c.DataFormat == "" {
		c.DataFormat = "json"
	} else if c.DataFormat == "" {
		c.DataFormat = "influx"
	}

	if node, ok := tbl.Fields["separator"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.Separator = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["templates"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.Templates = append(c.Templates, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["tag_keys"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.TagKeys = append(c.TagKeys, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["json_string_fields"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.JSONStringFields = append(c.JSONStringFields, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["json_name_key"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.JSONNameKey = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["json_query"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.JSONQuery = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["json_time_key"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.JSONTimeKey = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["json_time_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.JSONTimeFormat = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["json_timezone"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.JSONTimezone = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["json_strict"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if b, ok := kv.Value.(*ast.Boolean); ok {
				var err error
				c.JSONStrict, err = b.Boolean()
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if node, ok := tbl.Fields["data_type"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DataType = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["collectd_auth_file"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CollectdAuthFile = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["collectd_security_level"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CollectdSecurityLevel = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["collectd_parse_multivalue"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CollectdSplit = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["collectd_typesdb"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.CollectdTypesDB = append(c.CollectdTypesDB, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["dropwizard_metric_registry_path"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DropwizardMetricRegistryPath = str.Value
			}
		}
	}
	if node, ok := tbl.Fields["dropwizard_time_path"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DropwizardTimePath = str.Value
			}
		}
	}
	if node, ok := tbl.Fields["dropwizard_time_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DropwizardTimeFormat = str.Value
			}
		}
	}
	if node, ok := tbl.Fields["dropwizard_tags_path"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DropwizardTagsPath = str.Value
			}
		}
	}
	c.DropwizardTagPathsMap = make(map[string]string)
	if node, ok := tbl.Fields["dropwizard_tag_paths"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					if str, ok := kv.Value.(*ast.String); ok {
						c.DropwizardTagPathsMap[name] = str.Value
					}
				}
			}
		}
	}

	//for grok data_format
	if node, ok := tbl.Fields["grok_named_patterns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.GrokNamedPatterns = append(c.GrokNamedPatterns, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["grok_patterns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.GrokPatterns = append(c.GrokPatterns, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["grok_custom_patterns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.GrokCustomPatterns = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["grok_custom_pattern_files"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.GrokCustomPatternFiles = append(c.GrokCustomPatternFiles, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["grok_timezone"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.GrokTimezone = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["grok_unique_timestamp"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.GrokUniqueTimestamp = str.Value
			}
		}
	}

	//for csv parser
	if node, ok := tbl.Fields["csv_column_names"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.CSVColumnNames = append(c.CSVColumnNames, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_column_types"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.CSVColumnTypes = append(c.CSVColumnTypes, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_tag_columns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.CSVTagColumns = append(c.CSVTagColumns, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_delimiter"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVDelimiter = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_comment"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVComment = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_measurement_column"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVMeasurementColumn = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_timestamp_column"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVTimestampColumn = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_timestamp_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVTimestampFormat = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_timezone"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVTimezone = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_header_row_count"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if integer, ok := kv.Value.(*ast.Integer); ok {
				v, err := integer.Int()
				if err != nil {
					return nil, err
				}
				c.CSVHeaderRowCount = int(v)
			}
		}
	}

	if node, ok := tbl.Fields["csv_skip_rows"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if integer, ok := kv.Value.(*ast.Integer); ok {
				v, err := integer.Int()
				if err != nil {
					return nil, err
				}
				c.CSVSkipRows = int(v)
			}
		}
	}

	if node, ok := tbl.Fields["csv_skip_columns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if integer, ok := kv.Value.(*ast.Integer); ok {
				v, err := integer.Int()
				if err != nil {
					return nil, err
				}
				c.CSVSkipColumns = int(v)
			}
		}
	}

	if node, ok := tbl.Fields["csv_trim_space"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.Boolean); ok {
				//for config with no quotes
				val, err := strconv.ParseBool(str.Value)
				c.CSVTrimSpace = val
				if err != nil {
					return nil, fmt.Errorf("E! parsing to bool: %v", err)
				}
			}
		}
	}

	if node, ok := tbl.Fields["form_urlencoded_tag_keys"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.FormUrlencodedTagKeys = append(c.FormUrlencodedTagKeys, str.Value)
					}
				}
			}
		}
	}

	c.MetricName = name

	delete(tbl.Fields, "data_format")
	delete(tbl.Fields, "separator")
	delete(tbl.Fields, "templates")
	delete(tbl.Fields, "tag_keys")
	delete(tbl.Fields, "json_name_key")
	delete(tbl.Fields, "json_query")
	delete(tbl.Fields, "json_string_fields")
	delete(tbl.Fields, "json_time_format")
	delete(tbl.Fields, "json_time_key")
	delete(tbl.Fields, "json_timezone")
	delete(tbl.Fields, "json_strict")
	delete(tbl.Fields, "data_type")
	delete(tbl.Fields, "collectd_auth_file")
	delete(tbl.Fields, "collectd_security_level")
	delete(tbl.Fields, "collectd_typesdb")
	delete(tbl.Fields, "collectd_parse_multivalue")
	delete(tbl.Fields, "dropwizard_metric_registry_path")
	delete(tbl.Fields, "dropwizard_time_path")
	delete(tbl.Fields, "dropwizard_time_format")
	delete(tbl.Fields, "dropwizard_tags_path")
	delete(tbl.Fields, "dropwizard_tag_paths")
	delete(tbl.Fields, "grok_named_patterns")
	delete(tbl.Fields, "grok_patterns")
	delete(tbl.Fields, "grok_custom_patterns")
	delete(tbl.Fields, "grok_custom_pattern_files")
	delete(tbl.Fields, "grok_timezone")
	delete(tbl.Fields, "grok_unique_timestamp")
	delete(tbl.Fields, "csv_column_names")
	delete(tbl.Fields, "csv_column_types")
	delete(tbl.Fields, "csv_comment")
	delete(tbl.Fields, "csv_delimiter")
	delete(tbl.Fields, "csv_field_columns")
	delete(tbl.Fields, "csv_header_row_count")
	delete(tbl.Fields, "csv_measurement_column")
	delete(tbl.Fields, "csv_skip_columns")
	delete(tbl.Fields, "csv_skip_rows")
	delete(tbl.Fields, "csv_tag_columns")
	delete(tbl.Fields, "csv_timestamp_column")
	delete(tbl.Fields, "csv_timestamp_format")
	delete(tbl.Fields, "csv_timezone")
	delete(tbl.Fields, "csv_trim_space")
	delete(tbl.Fields, "form_urlencoded_tag_keys")

	return c, nil
}
