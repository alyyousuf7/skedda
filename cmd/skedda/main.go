package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alyyousuf7/skedda"
	"github.com/urfave/cli/v2"
)

func main() {
	defaultConfigPath := ""
	homedir, err := os.UserHomeDir()
	if err == nil {
		defaultConfigPath = path.Join(homedir, ".skedda")
	}

	var (
		noCache    = false
		configPath = defaultConfigPath
	)

	noCacheFlag := cli.BoolFlag{
		Name:        "no-cache",
		Aliases:     []string{"x"},
		Usage:       "Do not load venues and spaces from cache",
		Value:       noCache,
		DefaultText: "false",
		Destination: &noCache,
	}

	app := &cli.App{
		Name:  "skedda",
		Usage: "Book a space with Skedda",
		Commands: []*cli.Command{
			{
				Name:    "configure",
				Aliases: []string{"config"},
				Usage:   "Configure Skedda credentials",
				Action: func(c *cli.Context) error {
					config, _ := loadConfig(configPath)
					u, p := readCredentials(config.Username, config.Password)

					config.Username = u
					config.Password = p

					if err := saveConfig(configPath, config); err != nil {
						return err
					}

					fmt.Println("\n\nConfigured!")
					return nil
				},
			}, {
				Name:    "cache",
				Aliases: []string{"c"},
				Usage:   "Cache the venues and spaces",
				Action: func(c *cli.Context) error {
					config, err := loadConfig(configPath)
					if err != nil {
						return skedda.ErrCredsMissing
					}

					s, err := skedda.NewWithCreds(config.Username, config.Password)
					if err != nil {
						return err
					}

					venues, spaces, err := loadFromSkedda(s)
					if err != nil {
						return err
					}

					if err := saveToCache(venues, spaces, configPath); err != nil {
						fmt.Println("Failed to save cache")
					}

					return nil
				},
			}, {
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List venues and spaces",
				Flags: []cli.Flag{
					&noCacheFlag,
				},
				Action: func(c *cli.Context) error {
					config, err := loadConfig(configPath)
					if err != nil {
						return skedda.ErrCredsMissing
					}

					s, err := skedda.NewWithCreds(config.Username, config.Password)
					if err != nil {
						return err
					}

					venues, spaces, err := load(s, noCache, configPath)
					if err != nil {
						return err
					}

					for _, v := range venues {
						fmt.Println(v.Name)
						for _, s := range spaces {
							if s.VenueID == v.ID {
								fmt.Println("\t", s.Name)
							}
						}
					}
					return nil
				},
			}, {
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Find bookings",
				Flags: []cli.Flag{
					&noCacheFlag,
					&cli.StringFlag{
						Name:    "venue",
						Aliases: []string{"v"},
						Usage:   "Venue to check (selects all spaces in the venue)",
					},
					&cli.StringSliceFlag{
						Name:    "spaces",
						Aliases: []string{"space", "s"},
						Usage:   "Spaces to check",
					},
					&cli.StringFlag{
						Name:        "on",
						Aliases:     []string{"d"},
						Usage:       "`DATE` to check (possible values: today, tomorrow, YYYY-MM-DD)",
						DefaultText: "today",
					},
					&cli.GenericFlag{
						Name:        "from",
						Aliases:     []string{"a"},
						Usage:       "`TIME` from (H:mma)",
						DefaultText: time.Now().Truncate(15 * time.Minute).Format("3:04pm"),
						Value: &FlexibleTimestamp{
							Layouts: []string{"3:04pm", "3pm"},
						},
					},
					&cli.GenericFlag{
						Name:        "till",
						Aliases:     []string{"b"},
						Usage:       "`TIME` till (H:mma)",
						DefaultText: time.Now().Add(30 * time.Minute).Truncate(15 * time.Minute).Format("3:04pm"),
						Value: &FlexibleTimestamp{
							Layouts: []string{"3:04pm", "3pm"},
						},
					},
				},
				Action: func(c *cli.Context) error {
					var from, till time.Time
					onStr := strings.ToLower(c.String("on"))
					fromTmp := c.Value("from").(*time.Time)
					tillTmp := c.Value("till").(*time.Time)

					now := time.Now().UTC().Truncate(24 * time.Hour)
					var onDate time.Time
					switch onStr {
					case "":
						fallthrough
					case "today":
						onDate = now
					case "tomorrow":
						onDate = now.Add(24 * time.Hour)
					default:
						onDate, err = time.Parse("2006-01-02", onStr)
						if err != nil {
							return err
						}
					}
					if fromTmp == nil && tillTmp == nil { // Consider full day
						from = onDate
						till = from.Add(24 * time.Hour)
					} else if fromTmp != nil && tillTmp != nil {
						from = *fromTmp
						from = time.Date(onDate.Year(), onDate.Month(), onDate.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), time.UTC)
						till = *tillTmp
						till = time.Date(onDate.Year(), onDate.Month(), onDate.Day(), till.Hour(), till.Minute(), till.Second(), till.Nanosecond(), time.UTC)
					} else if fromTmp == nil && tillTmp != nil {
						return fmt.Errorf("--from is required when --till is provided")
					} else if fromTmp != nil && tillTmp == nil {
						from = *fromTmp
						from = time.Date(onDate.Year(), onDate.Month(), onDate.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), time.UTC)
						till = from.Add(30 * time.Minute)
					} else {
						return fmt.Errorf("report the inputs to the developer")
					}

					// Pull 'till' to the same date in case of full day
					if !from.Truncate(24 * time.Hour).Equal(till.Truncate(24 * time.Hour)) {
						till = till.Truncate(24 * time.Hour).Add(-15 * time.Minute)
					}

					if !from.Before(till) {
						return fmt.Errorf("--from cannot be ahead of --till")
					}

					config, _ := loadConfig(configPath)
					s, err := skedda.NewWithCreds(config.Username, config.Password)
					if err != nil {
						return err
					}

					venues, spaces, err := load(s, noCache, configPath)
					if err != nil {
						return err
					}

					var filteredSpaces skedda.SpaceList
					if c.String("venue") != "" {
						l := make([]fmt.Stringer, len(venues))
						for k, v := range venues {
							l[k] = v
						}
						matcher := NewMatcher(l)
						r := matcher.Match(c.String("venue"))

						list := make(skedda.VenueList, len(r))
						for k, v := range r {
							list[k] = v.(*skedda.Venue)
						}

						if len(list) == 0 {
							return fmt.Errorf("no venue found")
						}

						if len(list) > 1 {
							venueNames := list.Map(func(i int, v skedda.Venue) string {
								return v.Name
							})

							return fmt.Errorf("found multiple matching venues, be more specific: %s", strings.Join(venueNames, ", "))
						}

						venue := list[0]
						for _, s := range spaces {
							if s.VenueID == venue.ID {
								filteredSpaces = append(filteredSpaces, s)
							}
						}
					} else {
						l := make([]fmt.Stringer, len(spaces))
						for k, v := range spaces {
							l[k] = v
						}
						matcher := NewMatcher(l)
						r := matcher.MatchMultiple(c.StringSlice("spaces"))

						list := make(skedda.SpaceList, len(r))
						for k, v := range r {
							list[k] = v.(*skedda.Space)
						}

						filteredSpaces = list
					}

					if len(filteredSpaces) == 0 {
						return fmt.Errorf("no spaces found")
					}

					dateFormat := "Mon 02 Jan"
					timeFormat := "3:04pm"
					fmt.Printf("Finding bookings in %s on %s, between %s and %s...\n", strings.Join(filteredSpaces.Map(func(i int, s skedda.Space) string {
						return s.Name
					}), ", "), onDate.Format(dateFormat), from.Format(timeFormat), till.Format(timeFormat))

					if err := s.Auth(); err != nil {
						fmt.Printf("Failed to authenticate. You will not see the title of the bookings.\n\n")
					}

					type Result struct {
						Venue    *skedda.Venue
						Bookings []*skedda.Booking
						Error    error
					}

					worker := func(venue *skedda.Venue, resultCh chan<- Result, wg *sync.WaitGroup) {
						bookings, err := s.Bookings(venue.Domain, from, till)
						resultCh <- Result{venue, bookings, err}
						wg.Done()
					}

					filteredVenues := map[*skedda.Venue]bool{}
					for _, s := range filteredSpaces {
						venue := venues.FindByID(s.VenueID)
						filteredVenues[venue] = true
					}

					resultCh := make(chan Result, len(filteredVenues))
					var wg sync.WaitGroup
					for venue := range filteredVenues {
						wg.Add(1)
						go worker(venue, resultCh, &wg)
					}
					wg.Wait()
					close(resultCh)

					spaceBookings := map[*skedda.Space][]*skedda.Booking{}

					// create empty keys
					for _, space := range filteredSpaces {
						spaceBookings[space] = []*skedda.Booking{}
					}

					// fill up with result
					for result := range resultCh {
						if result.Error != nil {
							return result.Error
						}

						for _, booking := range result.Bookings {
							for _, spaceID := range booking.SpaceIDs {
								space := filteredSpaces.FindByID(spaceID)
								if space != nil {
									spaceBookings[space] = append(spaceBookings[space], booking)
								}
							}
						}
					}

					// sort
					keys := skedda.SpaceList{}
					for space := range spaceBookings {
						keys = append(keys, space)
					}
					sort.Sort(keys)

					for _, space := range keys {
						bookings := spaceBookings[space]
						venue := venues.FindByID(space.VenueID)
						fmt.Printf("\n%s -- %s\n", venue.Name, space.Name)
						if len(bookings) == 0 {
							fmt.Printf("\t* Slot is available *\n")
							continue
						}

						for i, booking := range bookings {
							fmt.Printf("\t%d. %s\n", i+1, booking)
						}
					}
					return nil
				},
			}, {
				Name:    "book",
				Aliases: []string{"f"},
				Usage:   "Book spaces for meeting",
				Flags: []cli.Flag{
					&noCacheFlag,
					&cli.StringSliceFlag{
						Name:     "spaces",
						Aliases:  []string{"space", "s"},
						Usage:    "Spaces to book",
						Required: true,
					},
					&cli.StringFlag{
						Name:        "on",
						Aliases:     []string{"d"},
						Usage:       "`DATE` to book (possible values: today, tomorrow, YYYY-MM-DD)",
						DefaultText: "today",
					},
					&cli.GenericFlag{
						Name:        "from",
						Aliases:     []string{"a"},
						Usage:       "`TIME` from (H:mma)",
						DefaultText: time.Now().Truncate(15 * time.Minute).Format("3:04pm"),
						Value: &FlexibleTimestamp{
							Layouts: []string{"3:04pm", "3pm"},
						},
					},
					&cli.GenericFlag{
						Name:        "till",
						Aliases:     []string{"b"},
						Usage:       "`TIME` till (H:mma)",
						DefaultText: time.Now().Add(30 * time.Minute).Truncate(15 * time.Minute).Format("3:04pm"),
						Value: &FlexibleTimestamp{
							Layouts: []string{"3:04pm", "3pm"},
						},
					},
					&cli.StringFlag{
						Name:     "title",
						Aliases:  []string{"t"},
						Usage:    "Title for the booking",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "assume-yes",
						Aliases: []string{"yes", "y"},
						Usage:   "Assume yes to al prompts and run non-interactively",
					},
				},
				Action: func(c *cli.Context) error {
					var from, till time.Time
					onStr := strings.ToLower(c.String("on"))
					fromTmp := c.Value("from").(*time.Time)
					tillTmp := c.Value("till").(*time.Time)

					now := time.Now().UTC().Truncate(15 * time.Minute).Truncate(24 * time.Hour)
					var onDate time.Time
					switch onStr {
					case "":
						fallthrough
					case "today":
						onDate = now
					case "tomorrow":
						onDate = now.Add(24 * time.Hour)
					default:
						onDate, err = time.Parse("2006-01-02", onStr)
						if err != nil {
							return err
						}
					}
					if fromTmp == nil && tillTmp == nil { // Consider full day
						from = onDate
						till = from.Add(24 * time.Hour)
					} else if fromTmp != nil && tillTmp != nil {
						from = *fromTmp
						from = time.Date(onDate.Year(), onDate.Month(), onDate.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), time.UTC)
						till = *tillTmp
						till = time.Date(onDate.Year(), onDate.Month(), onDate.Day(), till.Hour(), till.Minute(), till.Second(), till.Nanosecond(), time.UTC)
					} else if fromTmp == nil && tillTmp != nil {
						return fmt.Errorf("--from is required when --till is provided")
					} else if fromTmp != nil && tillTmp == nil {
						from = *fromTmp
						from = time.Date(onDate.Year(), onDate.Month(), onDate.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), time.UTC)
						till = from.Add(30 * time.Minute)
					} else {
						return fmt.Errorf("report the inputs to the developer")
					}

					// Pull 'till' to the same date in case of full day
					if !from.Truncate(24 * time.Hour).Equal(till.Truncate(24 * time.Hour)) {
						till = till.Truncate(24 * time.Hour).Add(-15 * time.Minute)
					}

					if !from.Before(till) {
						return fmt.Errorf("--from cannot be ahead of --till")
					}

					// Booking requires time to be 15min granular
					if !from.Equal(from.Truncate(15*time.Minute)) || !till.Equal(till.Truncate(15*time.Minute)) {
						return fmt.Errorf("--from and --till has to be round to 15 minutes for booking")
					}

					if (c.String("venue") != "") == (len(c.StringSlice("spaces")) > 0) {
						return fmt.Errorf("either provide venue or spaces")
					}

					title := c.String("title")
					title = strings.TrimSpace(title)
					if title == "" {
						return fmt.Errorf("--title is required")
					}

					config, err := loadConfig(configPath)
					if err != nil {
						return skedda.ErrCredsMissing
					}

					s, err := skedda.NewWithCreds(config.Username, config.Password)
					if err != nil {
						return err
					}

					venues, spaces, err := load(s, noCache, configPath)
					if err != nil {
						return err
					}

					var venueID int
					var filteredSpaces skedda.SpaceList
					{
						l := make([]fmt.Stringer, len(spaces))
						for k, v := range spaces {
							l[k] = v
						}
						matcher := NewMatcher(l)
						r := matcher.MatchMultiple(c.StringSlice("spaces"))

						list := make(skedda.SpaceList, len(r))
						for k, v := range r {
							list[k] = v.(*skedda.Space)
						}

						for _, space := range list {
							if venueID == 0 {
								venueID = space.VenueID
							}

							if space.VenueID != venueID {
								return fmt.Errorf("you must choose spaces from a single venue only")
							}
						}
						filteredSpaces = list
					}

					if len(filteredSpaces) == 0 || venueID == 0 {
						return fmt.Errorf("no spaces found")
					}
					venue := venues.FindByID(venueID)
					if venue == nil {
						return fmt.Errorf("could not find details about the venue")
					}

					dateFormat := "Mon 02 Jan"
					timeFormat := "3:04pm"
					fmt.Printf("Booking %s on %s, between %s and %s...\n", strings.Join(filteredSpaces.Map(func(i int, s skedda.Space) string {
						return s.Name
					}), ", "), onDate.Format(dateFormat), from.Format(timeFormat), till.Format(timeFormat))

					if !c.Bool("assume-yes") {
						fmt.Print("\nAre you sure? (y/N): ")
						var answer string
						fmt.Scanln(&answer)

						if !strings.HasPrefix(strings.ToLower(answer), "y") {
							return nil
						}
					}

					if err := s.Auth(); err != nil {
						return err
					}

					spaceIDs := []int{}
					for _, space := range filteredSpaces {
						spaceIDs = append(spaceIDs, space.ID)
					}
					if err := s.Book(venue.Domain, venue.ID, spaceIDs, title, from, till); err != nil {
						return err
					}

					fmt.Println("\nBooked!")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("\nError:", err)

		if errors.Is(err, skedda.ErrCredsMissing) {
			fmt.Printf("\nTry using `%s configure`\n", app.Name)
		} else if errors.Is(err, skedda.ErrAuthFailed) {
			fmt.Printf("\nTry changing credentials using `%s configure`\n", app.Name)
		}
		os.Exit(1)
	}
}
