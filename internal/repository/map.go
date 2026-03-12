package repository

import (
	"context"

	"id-100/internal/database"
)

// CityContrib holds a city name and its contribution count for the map endpoint.
// Geocoding (lat/lon) is handled client-side via Nominatim.
type CityContrib struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// GetCityContribCounts returns all distinct cities ordered by contribution count.
func GetCityContribCounts(ctx context.Context) ([]CityContrib, error) {
	rows, err := database.DB.Query(ctx, `
		SELECT user_city, COUNT(*) AS cnt
		FROM contributions
		WHERE user_city IS NOT NULL AND user_city != ''
		GROUP BY user_city
		ORDER BY cnt DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []CityContrib
	for rows.Next() {
		var c CityContrib
		if err := rows.Scan(&c.Name, &c.Count); err != nil {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}
