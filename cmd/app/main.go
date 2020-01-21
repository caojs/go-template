package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	configFile string
	rootCmd = &cobra.Command{
		Use: "app",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(viper.Get("test"))
		},
	}
)

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file directory")
	rootCmd.PersistentFlags().String("test", "", "test usable")
	if err := viper.BindPFlag("test", rootCmd.PersistentFlags().Lookup("test")); err != nil {
		fmt.Printf("Unable to bind flag test")
		os.Exit(0)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(0)
	}
}

func initConfig() {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Unable to read config: %s", viper.ConfigFileUsed())
	}
}

