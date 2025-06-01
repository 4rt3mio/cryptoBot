package coinbase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"CourseWork/infrastructure/logger"
)

const (
	spotURL             = "https://api.coinbase.com/v2/prices/%s-USD/spot"
	currenciesURL       = "https://api.coinbase.com/v2/currencies"
	cryptoCurrenciesURL = "https://api.coinbase.com/v2/assets?filter=listed" // Фильтр для листинговых криптовалют
	exchangeRatesURL    = "https://api.coinbase.com/v2/exchange-rates"
	historicURLFmt      = "https://api.coinbase.com/v2/prices/%s-USD/historic?period=day"
)

type coinbasePriceResponse struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

type Asset struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type exchangeRatesResponse struct {
	Data struct {
		Currency string            `json:"currency"`
		Rates    map[string]string `json:"rates"`
	} `json:"data"`
}

type historicalPriceResponse struct {
	Data struct {
		Prices []struct {
			Time  string `json:"time"`
			Price string `json:"price"`
		} `json:"prices"`
	} `json:"data"`
}

type Client struct {
	log  *logger.ZapLogger
	http *http.Client
}

func NewClient(log *logger.ZapLogger) *Client {
	return &Client{
		log:  log,
		http: &http.Client{},
	}
}

func (c *Client) GetPrice(symbol string) (float64, error) {
	url := fmt.Sprintf(spotURL, strings.ToUpper(symbol))
	c.log.Debug("Coinbase.GetPrice: запрос", "url", url)

	resp, err := c.http.Get(url)
	if err != nil {
		c.log.Error("Coinbase.GetPrice: http.Get failed", "err", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var pr coinbasePriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		c.log.Error("Coinbase.GetPrice: decode failed", "err", err)
		return 0, err
	}

	price, err := strconv.ParseFloat(pr.Data.Amount, 64)
	if err != nil {
		c.log.Error("Coinbase.GetPrice: parse failed", "amount", pr.Data.Amount, "err", err)
		return 0, err
	}

	c.log.Info("Coinbase.GetPrice: получили цену", "symbol", symbol, "price", price)
	return price, nil
}

func (c *Client) List() ([]string, error) {
    c.log.Debug("Coinbase.GetAllCurrencies: запрос", "url", exchangeRatesURL)

    resp, err := c.http.Get(exchangeRatesURL)
    if err != nil {
        c.log.Error("Coinbase.GetAllCurrencies: http.Get failed", "err", err)
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var er exchangeRatesResponse
    if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
        c.log.Error("Coinbase.GetAllCurrencies: decode failed", "err", err)
        return nil, err
    }

    out := make([]string, 0, len(er.Data.Rates))
    for symbol := range er.Data.Rates {
        if len(symbol) <= 5 && symbol != "USD" {
            out = append(out, symbol)
        }
    }

    c.log.Info("Coinbase.GetAllCurrencies: получили список криптовалют", "count", len(out))
    return out, nil
}

func (c *Client) GetDailyPrices(symbol string) ([]float64, error) {
	url := fmt.Sprintf(historicURLFmt, strings.ToUpper(symbol))
	c.log.Debug("Coinbase.GetDailyPrices: запрос", "url", url)

	resp, err := c.http.Get(url)
	if err != nil {
		c.log.Error("Coinbase.GetDailyPrices: http.Get failed", "err", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var hr historicalPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&hr); err != nil {
		c.log.Error("Coinbase.GetDailyPrices: decode failed", "err", err)
		return nil, err
	}

	if len(hr.Data.Prices) == 0 {
		msg := fmt.Sprintf("no historic data for %s", symbol)
		c.log.Warn("Coinbase.GetDailyPrices:", "warning", msg)
		return nil, fmt.Errorf("%s", msg)
	}

	prices := make([]float64, 0, len(hr.Data.Prices))
	for _, p := range hr.Data.Prices {
		if v, err := strconv.ParseFloat(p.Price, 64); err == nil {
			prices = append(prices, v)
		} else {
			c.log.Warn("Coinbase.GetDailyPrices: parse failed", "price", p.Price, "err", err)
		}
	}

	c.log.Info("Coinbase.GetDailyPrices: получили исторические цены", "symbol", symbol, "count", len(prices))
	return prices, nil
}
