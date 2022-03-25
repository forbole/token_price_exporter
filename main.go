package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	collector "github.com/forbole/token-price-exporter/collectors"
	token "github.com/forbole/token-price-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags
	cfgFile string
	port    string
	Config  token.Config
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-price-exporter",
		Short: "Scrape the token price from Coingecko and expose as Prometheus metrics.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cfgFile != "" {
				// Use config file from the flag.
				viper.SetConfigFile(cfgFile)
			} else {
				// Search config in home directory with name ".cobra" (without extension).
				viper.AddConfigPath("$HOME/.token_price_exporter/")
				viper.SetConfigType("yaml")
				viper.SetConfigName("config")
			}

			if err := viper.ReadInConfig(); err != nil {
				return err
			}

			err := viper.Unmarshal(&Config)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: Executor,
	}
	return cmd
}

func Executor(cmd *cobra.Command, args []string) error {
	tokens := Config.Tokens
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(
		collector.NewTokensPriceGauge(tokens),
	)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.New(os.Stderr, log.Prefix(), log.Flags()),
		ErrorHandling: promhttp.ContinueOnError,
	})

	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(port, nil))
	fmt.Printf("Start listening on port %s", port)
	return nil
}

func main() {
	cmd := NewRootCommand()
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.token_price_exporter/config.yaml)")
	cmd.PersistentFlags().StringVar(&port, "port", ":26666", "Exporter port")
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
