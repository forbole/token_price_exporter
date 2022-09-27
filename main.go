package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	token "github.com/forbole/token-price-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	coingecko "github.com/superoo7/go-gecko/v3"
)

var (
	// Used for flags
	cfgFile            string
	port               string
	Config             token.Config
	tokenPriceGaugeVec = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "token_price",
			Help: "Price of the token",
		},
		[]string{"token_id", "denom"},
	)
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
	var tokenPrice *map[string]map[string]float32
	getPrice(tokens, tokenPrice)

	http.Handle("/metrics", promhttp.Handler())
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

func getPrice(tokens []token.Token, tokenPrice *map[string]map[string]float32) {
	go func() {
		for {
			httpClient := &http.Client{
				Timeout: time.Second * 10,
			}

			cg := coingecko.NewClient(httpClient)
			ids := []string{}
			for _, token := range tokens {
				ids = append(ids, token.ID)
			}

			vc := []string{"usd"}

			sp, err := cg.SimplePrice(ids, vc)
			if err == nil {
				for _, token := range tokens {
					if tokenPrice, ok := (*sp)[token.ID]; ok {
						tokenPriceGaugeVec.WithLabelValues(token.ID, token.Denom).Set(float64(tokenPrice["usd"]))
					}
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()
}
