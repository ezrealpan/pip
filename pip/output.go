package pip

// Output ...
type Output interface {
	PluginDescriber

	// Connect to the Output
	Connect() error
	// Close any connections to the Output
	Close() error
	// Write takes in group of points to be written to the Output
	Write(metrics []Metric) error
}
