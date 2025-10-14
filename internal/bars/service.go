package bars

import (
	"context"
	"fmt"
	"go-router/internal/brreg"
	"go-router/internal/osm"
	"go-router/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateBar(ctx context.Context, conn *pgxpool.Pool, input BarManual) error {
	bar := Bar{BarManual: input}
	slug := utils.ToURL(bar.Name)
	bar.Slug = slug

	node, locData, err := osm.GetBarLocationData(input.OsmID)
	if err != nil {
		return err
	}

	addrID, err := upsertLocations(ctx, conn, locData)
	if err != nil {
		return err
	}

	bar.BarLocation = addrID

	err = brreg.CreateOrgIfNotExist(ctx, conn, bar.OrgNummer)
	if err != nil {
		return err
	}

	id, err := createBarRow(ctx, conn, bar)
	if err != nil {
		return fmt.Errorf("error adding bar to Supabase: %w", err)
	}

	if bar.LinkedBar {
		extraDetails := ExtractBarMetadata(id, &node)
		createBarMetadata(ctx, conn, extraDetails)
	}

	return nil
}

func upsertLocations(ctx context.Context, conn *pgxpool.Pool, adr osm.AddressParts) (BarLocation, error) {
	ids := BarLocation{Latitude: adr.Lat, Longitude: adr.Lon}

	fylke, err := GetLocationIdByName(ctx, conn, adr.Fylke, "fylke")
	if err != nil {
		return ids, fmt.Errorf("fylke not found: %w", err)
	}

	ids.Fylke = fylke

	kommune, err := GetLocationIdByName(ctx, conn, adr.Kommune, "kommune")
	if err != nil {
		newKommune := Location{
			BaseLocation: BaseLocation{
				Name: adr.Kommune,
				Slug: utils.ToURL(adr.Kommune),
			},
			Hierarchy: "kommune",
			Parent:    &fylke,
		}
		kommune, err = createNewLocation(ctx, conn, newKommune)
		if err != nil {
			return ids, fmt.Errorf("could not create kommune: %w", err)
		}
	}

	ids.Kommune = kommune

	if adr.Sted != "" {
		sted, err := GetLocationIdByName(ctx, conn, adr.Sted, "sted")
		if err != nil {
			newSted := Location{
				BaseLocation: BaseLocation{
					Name: adr.Sted,
					Slug: utils.ToURL(adr.Sted),
				},
				Hierarchy: "sted",
				Parent:    &kommune,
			}
			sted, err = createNewLocation(ctx, conn, newSted)
			if err != nil {
				return ids, fmt.Errorf("could not create sted: %w", err)
			}
		}

		ids.Sted = &sted
	}

	return ids, nil
}
