package geocoder

import "github.com/teddyandria/sentinel/internal/domain"

// DefaultCities renvoie le dictionnaire de villes de départ du StaticGeocoder.
func DefaultCities() map[string]domain.Location {
	return map[string]domain.Location{
		"Paris":         {Name: "Paris", Lat: 48.8566, Lon: 2.3522},
		"London":        {Name: "London", Lat: 51.5074, Lon: -0.1278},
		"Berlin":        {Name: "Berlin", Lat: 52.5200, Lon: 13.4050},
		"Madrid":        {Name: "Madrid", Lat: 40.4168, Lon: -3.7038},
		"Rome":          {Name: "Rome", Lat: 41.9028, Lon: 12.4964},
		"Brussels":      {Name: "Brussels", Lat: 50.8503, Lon: 4.3517},
		"Amsterdam":     {Name: "Amsterdam", Lat: 52.3676, Lon: 4.9041},
		"New York":      {Name: "New York", Lat: 40.7128, Lon: -74.0060},
		"San Francisco": {Name: "San Francisco", Lat: 37.7749, Lon: -122.4194},
		"Washington":    {Name: "Washington", Lat: 38.9072, Lon: -77.0369},
		"Tokyo":         {Name: "Tokyo", Lat: 35.6762, Lon: 139.6503},
		"Beijing":       {Name: "Beijing", Lat: 39.9042, Lon: 116.4074},
		"Moscow":        {Name: "Moscow", Lat: 55.7558, Lon: 37.6173},
		"Bangalore":     {Name: "Bangalore", Lat: 12.9716, Lon: 77.5946},
		"Sydney":        {Name: "Sydney", Lat: -33.8688, Lon: 151.2093},
	}
}
