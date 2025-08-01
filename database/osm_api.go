package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-router/models"
)

// Returns raw response from OSM and the address parts used for upserting locations
func FetchBarByNode(osmNode string) (models.NodeDetails, models.AddressParts, error) {
	var node models.NodeDetails
	var addr models.AddressParts

	url := fmt.Sprintf("https://nominatim.openstreetmap.org/lookup?osm_ids=%s&format=json&extratags=1", osmNode)
	fmt.Println(url)
	resp, err := apiResponse(url)
	fmt.Println(resp)
	if err != nil {
		return node, addr, err
	}
	defer resp.Body.Close()

	var data []models.NodeDetails
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return node, addr, fmt.Errorf("failed to decode response: %w", err)
	}
	if len(data) == 0 {
		return node, addr, errors.New("no node data found")
	}

	node = data[0]
	addr, _ = getAddressParts(&node)
	lat, _ := strconv.ParseFloat(node.Lat, 64)
	lon, err := strconv.ParseFloat(node.Lon, 64)
	addr.Lat = lat
	addr.Lon = lon
	if err != nil {
		return node, addr, err
	}
	return node, addr, nil
}

func apiResponse(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 7 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	return resp, nil
}

func splitStringByComma(data string) []string {
	parts := strings.Split(data, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func getAddressParts(osmNodeDetails *models.NodeDetails) (models.AddressParts, error) {
	parts := splitStringByComma(osmNodeDetails.DisplayName)

	// Accounts for places without postcode
	indexModifier := 0
	if osmNodeDetails.Address.Postcode == "" {
		indexModifier = 1
	}

	// Exits early if there are two few parts in the display name
	length := len(parts)
	if length < (5 - indexModifier) {
		return models.AddressParts{}, errors.New("not enough address parts in display_name")
	}

	// Algoritm to assign strings to keys. Typical format of diplay_name:
	// "Mad Goat Tap House, Teaterplassen, Grønland, Gamle Oslo, Oslo, 0188, Norge"
	modAddress := models.AddressParts{
		Sted:     parts[length-(5-indexModifier)],
		Kommune:  parts[length-(4-indexModifier)],
		Fylke:    parts[length-(3-indexModifier)],
		Postcode: parts[length-(2-indexModifier)],
	}

	err := addressQualityControl(&modAddress, osmNodeDetails.Address)
	if err != nil {
		return models.AddressParts{}, err
	}

	return modAddress, nil
}

func addressQualityControl(modAddress *models.AddressParts, controlAddress models.Address) error {
	if modAddress.Postcode != controlAddress.Postcode {
		return fmt.Errorf("postkode matcher ikke, registert postkode er %s og postkoden vi fikk er %s", controlAddress.Postcode, modAddress.Postcode)
	}
	if controlAddress.CountryCode != "no" {
		return fmt.Errorf("landskode er ikke NO, men %s", controlAddress.CountryCode)
	}
	if modAddress.Sted == controlAddress.Road {
		modAddress.Sted = ""
	}
	return nil
}
