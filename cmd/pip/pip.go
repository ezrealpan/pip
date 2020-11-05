package command

import (
	"context"

	"ezreal.com.cn/pip/config"
	"ezreal.com.cn/pip/pip/agent"
	_ "ezreal.com.cn/pip/pip/all"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	pipCfgFile string
)

// NewPipCmd ...
func NewPipCmd() *cobra.Command {
	var pipCmd = &cobra.Command{
		Use:   "pip [string to echo]",
		Short: "pip ",
		Long:  `https://github.com/influxdata/telegraf`,
		Run:   runPip,
	}

	pipCmd.PersistentFlags().StringVar(&pipCfgFile, "pipCfgFile", "./pip.toml", "config file (default is $HOME/pip.toml)")

	// Use config file from the flag.
	viper.SetConfigFile(pipCfgFile)
	return pipCmd
}

func runPip(cmd *cobra.Command, args []string) {

	inputFilters := []string{"HTTP_WZXY_FETCH_PASSERS"}
	outputFilters := []string{"HTTP_WZXY_SAVE_PASSERS"}
	processorFilters := []string{"HTTP_WZXY_FACE_COMPARE"}

	run(
		inputFilters,
		outputFilters,
		processorFilters,
	)
}

func run(inputFilters, outputFilters, processorFilters []string) {
	ctx, _ := context.WithCancel(context.Background())
	runAgent(
		ctx,
		inputFilters,
		outputFilters,
		processorFilters,
	)
}

func runAgent(ctx context.Context,
	inputFilters []string,
	outputFilters []string,
	processorFilters []string,
) error {
	c := config.NewConfig()
	c.InputFilters = inputFilters
	c.OutputFilters = outputFilters
	//c.ProcessorsFilters = processorFilters

	c.LoadConfig("./pip_config.toml")

	ag, err := agent.NewAgent(c)
	if err != nil {
		return err
	}

	return ag.Run(ctx)
}
