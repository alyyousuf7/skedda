# Skedda
Skedda CLI and library written in Golang.

## CLI Usage
```bash
$ go get -u github.com/alyyousuf7/skedda/cmd/skedda
```

![](https://user-images.githubusercontent.com/14050128/75245881-7a122b00-57f0-11ea-9e49-717a0f63c4a3.png)

## Library Usage
```bash
$ go get -u github.com/alyyousuf7/skedda
```

```golang
package main

import (
	"fmt"
	"time"

	"github.com/alyyousuf7/skedda"
)

func main() {
	s, _ := skedda.NewWithCreds("user@domain.com", "password")

	if err := s.Auth(); err != nil {
		panic(err)
	}

	primaryDomain, _ := s.PrimaryDomain()

	// List all subdomains
	domains, _ := s.Domains(primaryDomain)
	fmt.Println(domains)

	// Detail of each subdomain (venue)
	for _, domain := range domains {
		// Get venue details
		venue, spaces, _ := s.Venue(domain)
		fmt.Println("Venue:", venue)
		fmt.Println("Spaces:", spaces)

		// List bookings in the next hour
		bookings, _ := s.Bookings(domain, time.Now(), time.Now().Add(1*time.Hour))
		fmt.Println(bookings)

		// Book all spaces in the venue
		spaceIDs := []int{}
		for _, space := range spaces {
			spaceIDs = append(spaceIDs, space.ID)
		}
		title := "Demo booking"
		from := time.Now().Truncate(15 * time.Minute)
		till := from.Add(15 * time.Minute)
		s.Book(domain, venue.ID, spaceIDs, title, from, till)
	}
}
```

## TODO
- Make the code more testable
- Write tests
- Add "Remove booking" function
