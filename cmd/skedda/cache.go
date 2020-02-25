package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sync"

	"github.com/alyyousuf7/skedda"
)

func load(s *skedda.Skedda, noCache bool, configPath string) (skedda.VenueList, skedda.SpaceList, error) {
	if noCache {
		fmt.Println("Loading venues and spaces from Skedda...")
		return loadFromSkedda(s)
	}

	if venues, spaces, err := loadFromCache(configPath); err == nil {
		return venues, spaces, nil
	}

	fmt.Println("Caching venues and spaces from Skedda...")

	venues, spaces, err := loadFromSkedda(s)
	if err != nil {
		return nil, nil, err
	}

	if err := saveToCache(venues, spaces, configPath); err != nil {
		fmt.Println("Failed to cache")
	}

	return venues, spaces, nil
}

func loadFromCache(configPath string) (skedda.VenueList, skedda.SpaceList, error) {
	data := struct {
		Venues skedda.VenueList
		Spaces skedda.SpaceList
	}{}

	cacheFilename := path.Join(configPath, "cache")
	buf, err := ioutil.ReadFile(cacheFilename)
	if err != nil {
		return nil, nil, err
	}

	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, nil, err
	}

	return data.Venues, data.Spaces, nil
}

func saveToCache(venues skedda.VenueList, spaces skedda.SpaceList, configPath string) error {
	data := struct {
		Venues skedda.VenueList
		Spaces skedda.SpaceList
	}{venues, spaces}

	buf, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	cacheFilename := path.Join(configPath, "cache")
	return ioutil.WriteFile(cacheFilename, buf, 0600)
}

func loadFromSkedda(s *skedda.Skedda) (skedda.VenueList, skedda.SpaceList, error) {
	venues := skedda.VenueList{}
	spaces := skedda.SpaceList{}

	primaryDomain, err := s.PrimaryDomain()
	if err != nil {
		return nil, nil, fmt.Errorf("fetching primary domain: %w", err)
	}

	domains, err := s.Domains(primaryDomain)
	if err != nil {
		return nil, nil, err
	}

	type Result struct {
		Venue  *skedda.Venue
		Spaces skedda.SpaceList
		Error  error
	}

	worker := func(domain string, ch chan<- Result, wg *sync.WaitGroup) {
		venue, spaces, err := s.Venue(domain)
		ch <- Result{venue, spaces, err}
		wg.Done()
	}

	resultCh := make(chan Result, len(domains))
	var wg sync.WaitGroup
	for _, domain := range domains {
		wg.Add(1)
		go worker(domain, resultCh, &wg)
	}
	wg.Wait()
	close(resultCh)

	for result := range resultCh {
		if result.Error != nil {
			return nil, nil, err
		}

		venues = append(venues, result.Venue)
		spaces = append(spaces, result.Spaces...)
	}

	return venues, spaces, nil
}
