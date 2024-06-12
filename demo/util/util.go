package util

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	personURLToIDRe   = regexp.MustCompile(`swapi.dev/api/people/(\d+)/`)
	planetURLToIDRe   = regexp.MustCompile(`swapi.dev/api/planets/(\d+)/`)
	speciesURLToIDRe  = regexp.MustCompile(`swapi.dev/api/species/(\d+)/`)
	starshipURLToIDRe = regexp.MustCompile(`swapi.dev/api/starships/(\d+)/`)
	vehicleURLToIDRe  = regexp.MustCompile(`swapi.dev/api/vehicles/(\d+)/`)
	filmURLToIDRe     = regexp.MustCompile(`swapi.dev/api/films/(\d+)/`)
)

func PersonURLToID(url string) (int64, error) {
	return urlToID(url, personURLToIDRe)
}

func PlanetURLToID(url string) (int64, error) {
	return urlToID(url, planetURLToIDRe)
}

func VehicleURLToID(url string) (int64, error) {
	return urlToID(url, vehicleURLToIDRe)
}

func StarshipURLToID(url string) (int64, error) {
	return urlToID(url, starshipURLToIDRe)
}

func SpeciesURLToID(url string) (int64, error) {
	return urlToID(url, speciesURLToIDRe)
}

func FilmURLToID(url string) (int64, error) {
	return urlToID(url, filmURLToIDRe)
}

func urlToID(url string, re *regexp.Regexp) (int64, error) {
	matches := re.FindAllStringSubmatch(url, 1)
	if len(matches) != 1 {
		return 0, fmt.Errorf("unexpected url format: %s", url)
	}
	if len(matches[0]) != 2 {
		return 0, fmt.Errorf("unexpected url format: %s", url)
	}
	id, err := strconv.ParseInt(matches[0][1], 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
