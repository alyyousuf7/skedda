package skedda

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Skedda struct
type Skedda struct {
	cookiejar       *cookiejar.Jar
	username        string
	password        string
	isAuthenticated bool
}

var (
	// ErrAuthFailed is returned when authentication is failed
	ErrAuthFailed = fmt.Errorf("authentication failed")

	// ErrCredsMissing is returned when credentials are missing
	ErrCredsMissing = fmt.Errorf("missing credentials")
)

// New initializes Skedda instance
func New() (*Skedda, error) {
	c, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	return &Skedda{
		c,
		"",
		"",
		false,
	}, nil
}

// NewWithCreds initializes Skedda instance with credentials
func NewWithCreds(username, password string) (*Skedda, error) {
	c, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	return &Skedda{
		c,
		username,
		password,
		false,
	}, nil
}

func (s *Skedda) hasCredentials() bool {
	return s.username != "" && s.password != ""
}

// Auth authenticates into Skedda and stores session into a cookiejar
func (s *Skedda) Auth() error {
	if s.isAuthenticated {
		return nil
	}

	if !s.hasCredentials() {
		return ErrCredsMissing
	}

	c := http.Client{
		Jar: s.cookiejar,
	}

	bodyMap := map[string]map[string]interface{}{
		"login": {
			"username":        s.username,
			"password":        s.password,
			"rememberMe":      false,
			"arbitraryerrors": nil,
		},
	}

	body, err := json.Marshal(bodyMap)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://www.skedda.com/logins", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var e error
		detail, err := s.errorDetail(res.Body)
		if err != nil {
			e = fmt.Errorf("unknown status: %d", res.StatusCode)
		} else {
			e = errors.New(detail)
		}

		return fmt.Errorf("%w: %s", ErrAuthFailed, e)
	}

	s.isAuthenticated = true
	return nil
}

// PrimaryDomain gets the main Skedda subdomain against the credentials
func (s *Skedda) PrimaryDomain() (string, error) {
	if !s.isAuthenticated {
		if err := s.Auth(); err != nil {
			return "", err
		}
	}

	c := http.Client{
		Jar: s.cookiejar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	body := strings.NewReader(fmt.Sprintf("username=%s", s.username))
	req, err := http.NewRequest("POST", "https://www.skedda.com/account/login", body)
	if err != nil {
		return "", nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 302 {
		return "", fmt.Errorf("unknown status: %d", res.StatusCode)
	}

	redirectURL, err := url.Parse(res.Header.Get("Location"))
	if err != nil {
		return "", err
	}

	suffix := ".skedda.com"
	host := redirectURL.Hostname()
	if strings.HasPrefix(host, "www.") || !strings.HasSuffix(host, suffix) {
		if strings.HasSuffix(host, suffix) {
			q := redirectURL.Query()
			if err, ok := q["err"]; ok {
				return "", fmt.Errorf("request failed: %s", err[0])
			}
		}

		return "", fmt.Errorf("unknown URL: %s", redirectURL)
	}

	return strings.TrimSuffix(host, suffix), nil
}

// Domains gets all the Skedda subdomains against the credentials
func (s *Skedda) Domains(primaryDomain string) ([]string, error) {
	token, err := s.verificationToken(primaryDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification token: %w", err)
	}

	c := http.Client{
		Jar: s.cookiejar,
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s.skedda.com/webs", primaryDomain), nil)
	if err != nil {
		return nil, nil
	}
	req.Header.Add("X-Skedda-RequestVerificationToken", token)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unknown status: %d", res.StatusCode)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	bodyMap := struct {
		Web struct {
			OtherSubdomains map[string]string
		}
	}{}
	if err := json.Unmarshal(buf, &bodyMap); err != nil {
		return nil, err
	}

	domains := []string{}
	for k := range bodyMap.Web.OtherSubdomains {
		domains = append(domains, k)
	}

	return domains, nil
}

// Venue fetches venue details for a given domain
func (s *Skedda) Venue(domain string) (*Venue, []*Space, error) {
	token, err := s.verificationToken(domain)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get verification token: %w", err)
	}

	c := http.Client{
		Jar: s.cookiejar,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.skedda.com/webs", domain), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("X-Skedda-RequestVerificationToken", token)

	res, err := c.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		detail, err := s.errorDetail(res.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("unknown status: %d", res.StatusCode)
		}

		return nil, nil, errors.New(detail)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	bodyMap := struct {
		Venue  []Venue
		Spaces []Space
	}{}
	if err := json.Unmarshal(buf, &bodyMap); err != nil {
		return nil, nil, err
	}

	if len(bodyMap.Venue) != 1 {
		return nil, nil, fmt.Errorf("no venue found")
	}

	venue := &bodyMap.Venue[0]
	spaces := []*Space{}
	for i := range bodyMap.Spaces {
		spaces = append(spaces, &bodyMap.Spaces[i])
	}

	return venue, spaces, nil
}

// Bookings fetches all bookings between a domain during a time period
func (s *Skedda) Bookings(domain string, from, to time.Time) ([]*Booking, error) {
	token, err := s.verificationToken(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification token: %w", err)
	}

	c := http.Client{
		Jar: s.cookiejar,
	}

	dateFormat := "2006-01-02T15:04:05"
	url := fmt.Sprintf("https://%s.skedda.com/bookingslists?start=%s&end=%s", domain, url.QueryEscape(from.Format(dateFormat)), url.QueryEscape(to.Format(dateFormat)))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Skedda-RequestVerificationToken", token)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		detail, err := s.errorDetail(res.Body)
		if err != nil {
			return nil, fmt.Errorf("unknown status: %d", res.StatusCode)
		}

		return nil, errors.New(detail)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	bodyMap := struct {
		Bookings []Booking
	}{}
	if err := json.Unmarshal(buf, &bodyMap); err != nil {
		return nil, err
	}

	bookings := []*Booking{}
	for i, b := range bodyMap.Bookings {
		add := true
		// We have to manually filter out the recurring bookings as they are not
		// filtered by Skedda
		if len(b.RecurrenceRule.All()) > 0 {
			t1Start := time.Date(0, 0, 0, from.Hour(), from.Minute(), from.Second(), 0, time.UTC)
			t1End := time.Date(0, 0, 0, to.Hour(), to.Minute(), to.Second(), 0, time.UTC)

			t2Start := time.Date(0, 0, 0, b.StartTime.Hour(), b.StartTime.Minute(), b.StartTime.Second(), 0, time.UTC)
			t2End := time.Date(0, 0, 0, b.EndTime.Hour(), b.EndTime.Minute(), b.EndTime.Second(), 0, time.UTC)

			add = TimeOverlaps(t1Start, t1End, t2Start, t2End)
		}

		if add {
			bookings = append(bookings, &bodyMap.Bookings[i])
		}
	}

	return bookings, nil
}

// Book books a space in a domain
func (s *Skedda) Book(domain string, venueID int, spaceIDs []int, title string, from, to time.Time) error {
	token, err := s.verificationToken(domain)
	if err != nil {
		return fmt.Errorf("failed to get verification token: %w", err)
	}

	c := http.Client{
		Jar: s.cookiejar,
	}

	dateFormat := "2006-01-02T15:04:05"
	bodyMap := map[string]map[string]interface{}{
		"booking": {
			"start":  from.Truncate(1 * time.Minute).Format(dateFormat),
			"end":    to.Truncate(1 * time.Minute).Format(dateFormat),
			"title":  title,
			"venue":  venueID,
			"spaces": spaceIDs,
			"type":   1,
			"price":  0,
		},
	}

	body, err := json.Marshal(bodyMap)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s.skedda.com/bookings", domain), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("X-Skedda-RequestVerificationToken", token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		detail, err := s.errorDetail(res.Body)
		if err != nil {
			return fmt.Errorf("unknown status: %d", res.StatusCode)
		}

		return errors.New(detail)
	}

	return nil
}

func (s *Skedda) errorDetail(r io.Reader) (string, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	resBody := map[string][]map[string]interface{}{}
	if err := json.Unmarshal(buf, &resBody); err != nil {
		return "", err
	}

	detail, ok := resBody["errors"][0]["detail"]
	if !ok {
		return "", fmt.Errorf("unknown error")
	}

	return detail.(string), nil
}

func (s *Skedda) verificationToken(domain string) (string, error) {
	c := http.Client{
		Jar: s.cookiejar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			parts := strings.Split(req.URL.Hostname(), ".")
			if parts[0] == "www" {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.skedda.com/booking", domain), nil)
	if err != nil {
		return "", err
	}

	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == 302 {
		return "", fmt.Errorf("invalid domain")
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("unknown status: %d", res.StatusCode)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`<input name="__RequestVerificationToken" (?:.*) value="(.*)" \/>`)
	matches := re.FindStringSubmatch(string(buf))

	if len(matches) != 2 {
		return "", fmt.Errorf("verification token not found")
	}

	return matches[1], nil
}
