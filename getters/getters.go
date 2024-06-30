package getters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func packAddress(address string) string {
	baseURL := "https://ton-mainnet.s.chainbase.com/2fIrm29WQDAxTLmnD6NVdZWyVGE/v1/packAddress"

	// Construir la URL con el parámetro de la dirección
	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Error al analizar la URL base:", err)
		return ""
	}
	q := u.Query()
	q.Set("address", address)
	u.RawQuery = q.Encode()

	return u.String()
}

func GetAddress(unpakedAddress string) (string, error) {
	url := packAddress(unpakedAddress)

	// Crear un contexto con un tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar la solicitud HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error al realizar la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("error al leer la respuesta HTTP: %v", err)
	}

	if result != nil {
		if r, ok := result["result"].(string); ok {
			return r, nil
		}
	}

	return "", fmt.Errorf("error al desempaquetar la direccion: %v", err)
}

func detectAddress(address string) string {
	baseURL := "https://ton-mainnet.s.chainbase.com/2fIrm29WQDAxTLmnD6NVdZWyVGE/v1/detectAddress"

	// Construir la URL con el parámetro de la dirección
	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Error al analizar la URL base:", err)
		return ""
	}
	q := u.Query()
	q.Set("address", address)
	u.RawQuery = q.Encode()

	return u.String()
}

func IsAddress(address string) (bool, error) {
	// Construir la URL
	url := detectAddress(address)

	// Crear un contexto con un tiempo de espera
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Realizar la solicitud HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("error al crear la solicitud HTTP: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error al realizar la solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()

	// Leer y analizar la respuesta
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, fmt.Errorf("error al leer la respuesta HTTP: %v", err)
	}

	if result != nil {
		if result["result"] != nil {
			if bounceable, ok := result["result"].(map[string]interface{}); ok {
				if b64url, ok := bounceable["bounceable"].(map[string]interface{}); ok {
					log.Printf("debug: %v", b64url["b64url"])
					if addrb64, ok := b64url["b64url"].(string); ok && addrb64 == address {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}
