package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	ston_fi     = "stonfi"
	dedust      = "dedust"
	quote_token = "ton_EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c"
)

type Pools struct {
	StonFi string
	Dedust string
}

func InitPools() *Pools {
	return &Pools{
		StonFi: "",
		Dedust: "",
	}
}

func (p *Pools) GetPools(jetton string) error {
	response, err := getPools(jetton)
	if err != nil {
		return err
	}

	if data, ok := response["data"].([]interface{}); ok {
		for _, poolData := range data {
			if pool, ok := poolData.(map[string]interface{}); ok {
				relationship, ok := pool["relationships"].(map[string]interface{})
				if !ok {
					return errors.New("error: no se pudo obtener la relacion del token")
				}

				quote, ok := relationship["quote_token"].(map[string]interface{})
				if !ok {
					return errors.New("error: no se pudo obtener el quote token")
				}

				quote_data, ok := quote["data"].(map[string]interface{})
				if !ok {
					return errors.New("error: no se pudo obtener el quoote data")
				}

				quote_id, ok := quote_data["id"].(string)
				if !ok {
					return errors.New("error: no se pudo obtener el id del quote")
				}

				dex_data, ok := relationship["dex"].(map[string]interface{})
				if !ok {
					return errors.New("error: no se pudo obtener el dex")
				}

				dex, ok := dex_data["data"].(map[string]interface{})
				if !ok {
					return errors.New("error: no se pudo obtener el tipo de dex")
				}

				id, ok := dex["id"].(string)
				if !ok {
					return errors.New("error no se pudo obtener el id")
				}

				if id == ston_fi && quote_id == quote_token {
					attributes, ok := pool["attributes"].(map[string]interface{})
					if !ok {
						return errors.New("error: no se pudo obtener los atributos")
					}

					address, ok := attributes["address"].(string)
					if !ok {
						return errors.New("error: no se pudo obtener la direccion del pool")
					}

					p.StonFi = address
				}

				if id == dedust && quote_id == quote_token {
					attributes, ok := pool["attributes"].(map[string]interface{})
					if !ok {
						return errors.New("error: no se pudo obtener los atributos")
					}

					address, ok := attributes["address"].(string)
					if !ok {
						return errors.New("error: no se pudo obtener la direccion del pool")
					}

					p.Dedust = address
				}
			}
		}
	}

	return nil
}

/* Internal Function */

func getPools(contract string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.geckoterminal.com/api/v2/search/pools?query=%s&network=ton&page=1", contract)

	// Crear un contexto con un tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar la solicitud HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al realizar la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error al leer la respuesta HTTP: %v", err)
	}

	return result, nil
}
