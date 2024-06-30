package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type Event struct {
	EventID        string
	Wallet         string
	TonIn          *big.Int
	TonOut         *big.Int
	TokenIn        *big.Int
	TokenOut       *big.Int
	JettonAddress  string
	JettonName     string
	JettonSymbol   string
	JettonDecimals *big.Int
	BuyOrder       bool
	SellOrder      bool
	Timestamp      int64
}

type Events struct {
	Events []*Event
	mutex  sync.RWMutex
}

func Init() *Events {
	return &Events{
		Events: make([]*Event, 0),
	}
}

func (e *Events) TotalEvents() []*Event {
	return e.Events
}

func (e *Events) StoreEvent(event map[string]interface{}) error {
	eventID, ok := event["EventID"].(string)
	if !ok {
		return errors.New("error: event id no encontrado")
	}

	exists := false
	for _, ev := range e.Events {
		if ev.EventID == eventID {
			exists = true
			break
		}
	}
	if exists {
		return errors.New("error: el evento ya existe")
	}

	wallet, ok := event["Wallet"].(string)
	if !ok {
		return errors.New("error: wallet no encontrado")
	}

	tonIn, _ := getBigInt(event["TonIn"])
	tonOut, _ := getBigInt(event["TonOut"])
	tokenOut, _ := getBigInt(event["TokenOut"])
	tokenIn, _ := getBigInt(event["TokenIn"])

	jettonAddress, ok := event["JettonAddress"].(string)
	if !ok {
		return errors.New("error: no se pudo obtener la direccion del jetton")
	}

	jettonName, ok := event["JettonName"].(string)
	if !ok {
		return errors.New("error: no se pudo obtener la direccion del jetton")
	}

	jettonSymbol, ok := event["JettonSymbol"].(string)
	if !ok {
		return errors.New("error: no se pudo obtener la direccion del jetton")
	}

	jettonDecimals, _ := getBigInt(event["JettonDecimals"])

	buyOrder, ok := event["BuyOrder"].(bool)
	if !ok {
		return errors.New("error: no se pudo obtener el tipo de orden")
	}

	sellOrder, ok := event["SellOrder"].(bool)
	if !ok {
		return errors.New("error: no se pudo obtener el tipo de orden")
	}

	timestamp, _ := getInt64(event["Timestamp"])

	newEvent := &Event{
		EventID:        eventID,
		Wallet:         wallet,
		TonIn:          tonIn,
		TonOut:         tonOut,
		TokenIn:        tokenOut,
		TokenOut:       tokenIn,
		JettonAddress:  jettonAddress,
		JettonName:     jettonName,
		JettonSymbol:   jettonSymbol,
		JettonDecimals: jettonDecimals,
		BuyOrder:       buyOrder,
		SellOrder:      sellOrder,
		Timestamp:      timestamp,
	}

	e.Events = append(e.Events, newEvent)

	return nil
}

func (e *Events) GetLastEvent(routerAddr string, api string) (*Event, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	events, err := getEvents(routerAddr, api)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("error: arreglo vacío")
	}

	/* 	log.Print(data["events"].([]interface{})) */

	eventsData, ok := events["events"].([]interface{})
	if !ok {
		return nil, errors.New("error: no se pudo obtener los datos del evento")
	}

	for _, eventData := range eventsData {
		event, ok := eventData.(map[string]interface{})
		if !ok {
			return nil, errors.New("error: formato de evento inválido")
		}

		actionsData, ok := event["actions"].([]interface{})
		if !ok {
			return nil, errors.New("error: no se pudieron obtener las acciones del evento")
		}

		log.Println("agregando evento con hash:", event["event_id"].(string))

		for _, actionData := range actionsData {
			action, ok := actionData.(map[string]interface{})
			if !ok {
				return nil, errors.New("error: formato de acción inválido")
			}

			actionType, ok := action["type"].(string)
			if !ok {
				return nil, errors.New("error: no se pudo obtener el tipo de acción")
			}

			txStatus, ok := action["status"].(string)
			if !ok {
				return nil, errors.New("error: estado de transacción no encontrado")
			}

			if actionType == "JettonSwap" && txStatus == "ok" {
				jettonAction, ok := action["JettonSwap"].(map[string]interface{})
				if !ok {
					return nil, errors.New("error: formato de cartera de usuario inválido")
				}

				log.Print(jettonAction)

				userWallet, ok := jettonAction["user_wallet"].(map[string]interface{})
				if !ok {
					return nil, errors.New("error: formato de cartera de usuario inválido")
				}

				var jettonAddress map[string]interface{}
				if jettonAction["jetton_master_out"] != nil {
					jettonAddress, ok = jettonAction["jetton_master_out"].(map[string]interface{})
				} else {
					jettonAddress, ok = jettonAction["jetton_master_in"].(map[string]interface{})
				}
				if !ok {
					return nil, errors.New("error: formato de dirección jetton inválido")
				}

				eventID, ok := event["event_id"].(string)
				if !ok {
					return nil, errors.New("error: no se pudo obtener el ID del evento")
				}

				exists := false
				for _, ev := range e.Events {
					if ev.EventID == eventID {
						exists = true
						break
					}
				}
				if exists {
					return nil, errors.New("error: el evento ya existe")
				}

				tonIn, _ := getBigInt(jettonAction["ton_in"])
				tonOut, _ := getBigInt(jettonAction["ton_out"])
				tokenIn, _ := getBigInt(jettonAction["amount_in"])
				tokenOut, _ := getBigInt(jettonAction["amount_out"])

				orderBuy, _ := getInt64(jettonAction["ton_in"])
				orderSell, _ := getInt64(jettonAction["ton_out"])

				timestamp, _ := getInt64(event["timestamp"])

				decimals, _ := getBigInt(jettonAddress["decimals"])

				newEvent := &Event{
					EventID:        eventID,
					BuyOrder:       orderBuy != 0,
					SellOrder:      orderSell != 0,
					Wallet:         userWallet["address"].(string),
					TonIn:          tonIn,
					TonOut:         tonOut,
					TokenIn:        tokenIn,
					TokenOut:       tokenOut,
					JettonAddress:  jettonAddress["address"].(string),
					JettonName:     jettonAddress["name"].(string),
					JettonSymbol:   jettonAddress["symbol"].(string),
					JettonDecimals: decimals,
					Timestamp:      timestamp,
				}

				log.Print("Nuevo evento desde STON.FI")

				log.Println("Nuevo evento detectado:", newEvent)

				e.Events = append(e.Events, newEvent)

				log.Println("evento guardado:", eventID)

				return newEvent, nil

			} else if actionType == "SmartContractExec" && txStatus == "ok" {
				eventID, ok := event["event_id"].(string)
				if !ok {
					return nil, errors.New("error: no se pudo obtener el ID del evento")
				}

				operationData, _ := traceEvents(eventID, api)

				actions, ok := operationData["actions"].([]interface{})
				if !ok {
					return nil, errors.New("error: no se pudo obtener las acciones del evento")
				}

				if len(actions) > 0 {
					contractExec := actions[0].(map[string]interface{})

					if buyOperation, ok := contractExec["SmartContractExec"].(map[string]interface{}); ok {
						executor := buyOperation["executor"].(map[string]interface{})
						operator := executor["address"].(string)
						tonIn, err := getBigInt(buyOperation["ton_attached"])
						if err != nil {
							log.Printf("se obtuvo un error al convertir el monto de tonIn, con el error: %v", err)
						}

						if len(actions) > 3 {
							jettonExec := actions[3].(map[string]interface{})

							jettonTransfer, ok := jettonExec["JettonTransfer"].(map[string]interface{})
							if !ok {
								return nil, errors.New("error obteniendo la transferencia de tokens")
							}

							tokenOut, err := getBigInt(jettonTransfer["amount"])
							if err != nil {
								log.Printf("se obtuvo un error al convertir el monto de tokenOut, con el error: %v", err)
							}

							jettonAddress, ok := jettonTransfer["jetton"].(map[string]interface{})
							if !ok {
								return nil, errors.New("error obteniendo los datos del tokens")
							}

							exists := false
							for _, ev := range e.Events {
								if ev.EventID == eventID {
									exists = true
									break
								}
							}
							if exists {
								return nil, errors.New("error: el evento ya existe")
							}

							decimals, _ := getBigInt(jettonAddress["decimals"])

							newEvent := &Event{
								EventID:        eventID,
								BuyOrder:       true,
								SellOrder:      false,
								Wallet:         operator,
								TonIn:          tonIn,
								TonOut:         big.NewInt(0),
								TokenIn:        big.NewInt(0),
								TokenOut:       tokenOut,
								JettonAddress:  jettonAddress["address"].(string),
								JettonName:     jettonAddress["name"].(string),
								JettonSymbol:   jettonAddress["symbol"].(string),
								Timestamp:      int64(event["timestamp"].(float64)),
								JettonDecimals: decimals,
							}

							log.Print("Nuevo evento desde DEDUST")

							log.Println("Nuevo evento detectado:", newEvent)

							e.Events = append(e.Events, newEvent)

							log.Println("evento guardado:", eventID)
							return newEvent, nil

						}
					}

					if sellOperation, ok := contractExec["JettonTransfer"].(map[string]interface{}); ok {
						executor := sellOperation["sender"].(map[string]interface{})
						operator := executor["address"].(string)

						jettonAddress, ok := sellOperation["jetton"].(map[string]interface{})
						if !ok {
							return nil, errors.New("error obteniendo los datos del tokens")
						}

						decimals, _ := getBigInt(jettonAddress["decimals"])

						tokenIn, err := getBigInt(sellOperation["amount"])
						if err != nil {
							log.Printf("se obtuvo un error al convertir el monto de tokenIn, con el error: %v", err)
						}

						if len(actions) > 3 {
							tonExec := actions[3].(map[string]interface{})

							tonTransfer, ok := tonExec["TonTransfer"].(map[string]interface{})
							if !ok {
								return nil, errors.New("error obteniendo la transferencia de tokens")
							}

							exists := false
							for _, ev := range e.Events {
								if ev.EventID == eventID {
									exists = true
									break
								}
							}
							if exists {
								return nil, errors.New("error: el evento ya existe")
							}

							tonOut, err := getBigInt(tonTransfer["amount"])
							if err != nil {
								log.Printf("se obtuvo un error al convertir el monto de tonOut, con el error: %v", err)
							}

							newEvent := &Event{
								EventID:        eventID,
								BuyOrder:       false,
								SellOrder:      true,
								Wallet:         operator,
								TonIn:          big.NewInt(0),
								TonOut:         tonOut,
								TokenIn:        tokenIn,
								TokenOut:       big.NewInt(0),
								JettonAddress:  jettonAddress["address"].(string),
								JettonName:     jettonAddress["name"].(string),
								JettonSymbol:   jettonAddress["symbol"].(string),
								Timestamp:      int64(event["timestamp"].(float64)),
								JettonDecimals: decimals,
							}

							log.Print("Nuevo evento desde DEDUST")

							log.Println("Nuevo evento detectado:", newEvent)

							e.Events = append(e.Events, newEvent)

							log.Println("evento guardado:", eventID)
							return newEvent, nil
						}
					}
				}

			}
		}
	}
	return nil, errors.New("error: no se pudo crear el evento")
}

/* Internal Functions */

func getEvents(lpAddress string, api string) (map[string]interface{}, error) {
	headers := map[string]string{
		"X-API-KEY": api,
	}
	url := fmt.Sprintf("https://tonapi.io/v2/accounts/%s/events?initiator=false&subject_only=false&limit=1", lpAddress)

	// Crear un contexto con un tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar la solicitud HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}

	client := &http.Client{}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

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

func getBigInt(value interface{}) (*big.Int, error) {
	switch v := value.(type) {
	case float64:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int:
		return big.NewInt(int64(v)), nil
	case int64:
		return big.NewInt(v), nil
	case string:
		bigInt := new(big.Int)
		_, ok := bigInt.SetString(v, 10)
		if !ok {
			return big.NewInt(0), nil
		}
		return bigInt, nil
	default:
		return big.NewInt(0), nil
	}
}

func getInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, errors.New("no se pudo convertir a int64")
	}
}

func traceEvents(event_id string, api string) (map[string]interface{}, error) {
	headers := map[string]string{
		"X-API-KEY": api,
	}
	url := fmt.Sprintf("https://tonapi.io/v2/events/%s", event_id)

	// Crear un contexto con un tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar la solicitud HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}

	client := &http.Client{}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

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
