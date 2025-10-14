package brreg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateOrgIfNotExist(ctx context.Context, conn *pgxpool.Pool, orgnr string) error {
	_, err := getOrgID(ctx, conn, "underenheter", orgnr)
	if err != nil {
		underenhet, err := fetchUnderenhet(orgnr)
		if err != nil {
			return fmt.Errorf("failed fetching underenhet from brreg: %w", err)
		}
		_, err = getOrgID(ctx, conn, "organisasjoner", underenhet.Orgnummer)
		if err != nil {
			hovedenhet, err := fetchHovedenhet(underenhet.Parent)
			if err != nil {
				return fmt.Errorf("failed fetching hovedenhet from brreg: %w", err)
			}
			if err := createHovedenhet(ctx, conn, hovedenhet); err != nil {
				return fmt.Errorf("failed creating hovedenhet: %w", err)
			}
		}
		if err := createUnderenhet(ctx, conn, underenhet); err != nil {
			return fmt.Errorf("failed creating hovedenhet: %w", err)
		}
	}
	return nil
}

func fetchUnderenhet(orgnum string) (Underenhet, error) {
	api := fmt.Sprintf("https://data.brreg.no/enhetsregisteret/api/underenheter/%s", orgnum)
	resp, err := http.Get(api)
	var underenhet Underenhet
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

func fetchHovedenhet(orgnum string) (Hovedenhet, error) {
	api := fmt.Sprintf("https://data.brreg.no/enhetsregisteret/api/enheter/%s", orgnum)
	resp, err := http.Get(api)
	var enhet Hovedenhet
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
