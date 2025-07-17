package database

import (
	"encoding/json"
	"fmt"
	"go-router/models"
	"io"
	"net/http"
)

func FetchUnderenhet(orgnum string) (models.Underenhet, error) {
	api := fmt.Sprintf("https://data.brreg.no/enhetsregisteret/api/underenheter/%s", orgnum)
	resp, err := http.Get(api)
	var underenhet models.Underenhet
	if err != nil {
		return underenhet, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return underenhet, err
	}
	if err := json.Unmarshal(body, &underenhet); err != nil {
		return underenhet, err
	}

	return underenhet, nil
}

func FetchHovedenhet(orgnum string) (models.Hovedenhet, error) {
	api := fmt.Sprintf("https://data.brreg.no/enhetsregisteret/api/enheter/%s", orgnum)
	resp, err := http.Get(api)
	var enhet models.Hovedenhet
	if err != nil {
		return enhet, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return enhet, err
	}
	if err := json.Unmarshal(body, &enhet); err != nil {
		return enhet, err
	}

	return enhet, nil
}
