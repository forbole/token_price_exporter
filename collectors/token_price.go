package collector

import (
	"net/http"
	"time"

	token "github.com/forbole/token-price-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
	coingecko "github.com/superoo7/go-gecko/v3"
)

type TokensPriceGauge struct {
	Tokens []token.Token
	Desc   *prometheus.Desc
}

func NewTokensPriceGauge(tokens []token.Token) *TokensPriceGauge {
	return &TokensPriceGauge{
		Tokens: tokens,
		Desc: prometheus.NewDesc(
			"token_price",
			"Price of the token",
			[]string{"token_id", "denom"},
			nil,
		),
	}
}

func (collector *TokensPriceGauge) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.Desc
}

func (collector *TokensPriceGauge) Collect(ch chan<- prometheus.Metric) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	cg := coingecko.NewClient(httpClient)
	ids := []string{}
	for _, token := range collector.Tokens {
		ids = append(ids, token.ID)
	}

	vc := []string{"usd"}

	sp, err := cg.SimplePrice(ids, vc)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.Desc, err)
	}

	for _, token := range collector.Tokens {
		tokenPrice := (*sp)[token.ID]
		ch <- prometheus.MustNewConstMetric(collector.Desc, prometheus.GaugeValue, float64(tokenPrice["usd"]), token.ID, token.Denom)
	}

}
