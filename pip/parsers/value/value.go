package value

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ezreal.com.cn/pip/pip"
	"ezreal.com.cn/pip/pip/metric"
)

// ValueParser ...
type ValueParser struct {
	MetricName  string
	DataType    string
	DefaultTags map[string]string
}

// Parse ...
func (v *ValueParser) Parse(buf []byte) ([]pip.Metric, error) {
	vStr := string(bytes.TrimSpace(bytes.Trim(buf, "\x00")))

	// unless it's a string, separate out any fields in the buffer,
	// ignore anything but the last.
	if v.DataType != "string" {
		values := strings.Fields(vStr)
		if len(values) < 1 {
			return []pip.Metric{}, nil
		}
		vStr = string(values[len(values)-1])
	}

	var value interface{}
	var err error
	switch v.DataType {
	case "", "int", "integer":
		value, err = strconv.Atoi(vStr)
	case "float", "long":
		value, err = strconv.ParseFloat(vStr, 64)
	case "str", "string":
		value = vStr
	case "bool", "boolean":
		value, err = strconv.ParseBool(vStr)
	}
	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{"value": value}
	metric, err := metric.New(v.MetricName, v.DefaultTags,
		fields, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return []pip.Metric{metric}, nil
}

// ParseLine ...
func (v *ValueParser) ParseLine(line string) (pip.Metric, error) {
	metrics, err := v.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

func (v *ValueParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}
