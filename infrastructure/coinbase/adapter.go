package coinbase

import (
	"github.com/4rt3mio/cryptoCore/domain/repository"
)

type CryptoRepositoryAdapter struct {
	client *Client
}

func NewCryptoRepositoryAdapter(client *Client) repository.CryptoRepository {
	return &CryptoRepositoryAdapter{client: client}
}

func (a *CryptoRepositoryAdapter) GetPrice(symbol string) (float64, error) {
	return a.client.GetPrice(symbol)
}

type CurrencyRepositoryAdapter struct {
	client *Client
}

func NewCurrencyRepositoryAdapter(client *Client) repository.CurrencyRepository {
	return &CurrencyRepositoryAdapter{client: client}
}

func (a *CurrencyRepositoryAdapter) ListCurrencies() ([]string, error) {
	return a.client.List()
}

func (a *CryptoRepositoryAdapter) GetDailyPrices(symbol string) ([]float64, error) {
	return a.client.GetDailyPrices(symbol)
}
