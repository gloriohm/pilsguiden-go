package osm

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-router/internal/utils"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetBarLocationData(osmID string) (NodeDetails, AddressParts, error) {
	// fetch location and bar details from OSM based on OSM Node
	nodeDetails, address, err := fetchBarByNode(osmID)
	if err != nil {
		return nodeDetails, address, err
	}

	return nodeDetails, address, nil
}

// Returns raw response from OSM and the address parts used for upserting locations
func fetchBarByNode(osmNode string) (NodeDetails, AddressParts, error) {
	var node NodeDetails
	var addr AddressParts

	url := fmt.Sprintf("https://nominatim.openstreetmap.org/lookup?osm_ids=%s&format=json&extratags=1", osmNode)
	log.Println("fetching from:", url)
	resp, err := apiResponse(url)
	log.Println(resp)
	if err != nil {
		return node, addr, err
	}
	defer resp.Body.Close()

	var data []NodeDetails
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return node, addr, fmt.Errorf("failed to decode response: %w", err)
	}
	if len(data) == 0 {
		return node, addr, errors.New("no node data found")
	}

	node = data[0]
	addr, err = getAddressParts(&node)
	if err != nil {
		return node, addr, err
	}

	lat, err := strconv.ParseFloat(node.Lat, 64)
	if err != nil {
		return node, addr, err
	}

	lon, err := strconv.ParseFloat(node.Lon, 64)
	if err != nil {
		return node, addr, err
	}

	addr.Lat = lat
	addr.Lon = lon

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

func getAddressParts(osmNodeDetails *NodeDetails) (AddressParts, error) {
	parts := utils.SplitStringByComma(osmNodeDetails.DisplayName)

	// Accounts for places without postcode
	indexModifier := 0
	if osmNodeDetails.Address.Postcode == "" {
		indexModifier = 1
	}

	// Exits early if there are too few parts in the display name
	length := len(parts)
	if length < (5 - indexModifier) {
		return AddressParts{}, errors.New("not enough address parts in display_name")
	}

	// Algoritm to assign strings to keys. Typical format of diplay_name:
	// "Mad Goat Tap House, Teaterplassen, Grønland, Gamle Oslo, Oslo, 0188, Norge"
	modAddress := AddressParts{
		Sted:     parts[length-(5-indexModifier)],
		Kommune:  parts[length-(4-indexModifier)],
		Fylke:    parts[length-(3-indexModifier)],
		Postcode: parts[length-(2-indexModifier)],
	}

	err := addressQualityControl(&modAddress, osmNodeDetails.Address)
	if err != nil {
		return AddressParts{}, err
	}

	return modAddress, nil
}

func addressQualityControl(modAddress *AddressParts, controlAddress address) error {
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
