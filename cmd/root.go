package cmd

import (
	pip "ezreal.com.cn/pip/cmd/pip"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "root",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
				This application is a tool to generate the needed files
				to quickly create a Cobra application.
				`,
	}
)

// Execute executes the root command.
func Execute() error {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "cfgFile", "./maya.toml", "config file (default is $HOME/maya.toml)")

	// Use config file from the flag.
	viper.SetConfigFile(cfgFile)

	//NewPipCmd
	rootCmd.AddCommand(pip.NewPipCmd())

	return rootCmd.Execute()
}
