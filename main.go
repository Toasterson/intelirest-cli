package main

import (
	"encoding/json"
	"intelirest-cli/parser"
	"intelirest-cli/runtime"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "rest-cli",
	Short: "RFC-2616 compliant request file runner for CLI's",
	Args:  cobra.MinimumNArgs(1),
	RunE:  execute,
}

func main() {
	f := rootCmd.Flags()
	f.StringP("environment", "e", "", "specify environment to run")
	f.IntP("maxconns", "M", 4, "maximum number of connections for the client")
	f.BoolP("verbose", "v", false, "enable verbose output")

	if err := viper.BindPFlags(f); err != nil {
		panic(err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func execute(_ *cobra.Command, args []string) error {
	envName := viper.GetString("environment")
	env, err := runtime.ReadEnvironment(envName)
	if err != nil {
		return err
	}

	p, err := parser.New(args[0], env)
	if err != nil {
		return err
	}

	requests, err := p.Parse()
	if err != nil {
		return err
	}

	client := runtime.New(viper.GetInt("maxconns"))
	if viper.GetBool("verbose") {
		client.SetVerbose()
	}

	responses, err := client.Do(requests)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(responses); err != nil {
		return err
	}

	return nil
}
