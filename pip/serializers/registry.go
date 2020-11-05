package serializers

import (
	"fmt"

	"ezreal.com.cn/pip/pip"
)

// SerializerOutput is an interface for output plugins that are able to
// serialize telegraf metrics into arbitrary data formats.
type SerializerOutput interface {
	// SetSerializer sets the serializer function for the interface.
	SetSerializer(serializer Serializer)
}

// Serializer is an interface defining functions that a serializer plugin must
// satisfy.
//
// Implementations of this interface should be reentrant but are not required
// to be thread-safe.
type Serializer interface {
	// Serialize takes a single telegraf metric and turns it into a byte buffer.
	// separate metrics should be separated by a newline, and there should be
	// a newline at the end of the buffer.
	//
	// New plugins should use SerializeBatch instead to allow for non-line
	// delimited metrics.
	Serialize(metric pip.Metric) ([]byte, error)

	// SerializeBatch takes an array of telegraf metric and serializes it into
	// a byte buffer.  This method is not required to be suitable for use with
	// line oriented framing.
	SerializeBatch(metrics []pip.Metric) ([]byte, error)
}

// Config is a struct that covers the data types needed for all serializer types,
// and can be used to instantiate _any_ of the serializers.
type Config struct {
	// Dataformat can be one of the serializer types listed in NewSerializer.
	DataFormat string `toml:"data_format"`
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	// case "influx":
	// 	serializer, err = NewInfluxSerializerConfig(config)
	// case "graphite":
	// 	serializer, err = NewGraphiteSerializer(config.Prefix, config.Template, config.GraphiteTagSupport, config.GraphiteSeparator, config.Templates)
	// case "json":
	// 	serializer, err = NewJsonSerializer(config.TimestampUnits)
	// case "splunkmetric":
	// 	serializer, err = NewSplunkmetricSerializer(config.HecRouting, config.SplunkmetricMultiMetric)
	// case "nowmetric":
	// 	serializer, err = NewNowSerializer()
	// case "carbon2":
	// 	serializer, err = NewCarbon2Serializer()
	// case "wavefront":
	// 	serializer, err = NewWavefrontSerializer(config.Prefix, config.WavefrontUseStrict, config.WavefrontSourceOverride)
	// case "prometheus":
	// 	serializer, err = NewPrometheusSerializer(config)
	default:
		err = fmt.Errorf("Invalid data format: %s", config.DataFormat)
	}
	return serializer, err
}
