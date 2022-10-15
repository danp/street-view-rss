package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/tools/blog/atom"
)

//go:embed index.html
var index []byte

func main() {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")

	feedUpdated := time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})

	mux.HandleFunc("/atom.xml", func(w http.ResponseWriter, r *http.Request) {
		ls := r.URL.Query()["l"]
		if len(ls) == 0 {
			http.Error(w, "need ls", http.StatusBadRequest)
			return
		}
		sort.Strings(ls)

		fenc := base64.RawURLEncoding.EncodeToString([]byte(strings.Join(ls, " ")))

		feed := atom.Feed{
			Title:   "Street view updates",
			ID:      "urn:street-view-updates:feed:" + fenc,
			Updated: atom.Time(feedUpdated),
		}

		for _, l := range ls {
			enc := base64.RawURLEncoding.EncodeToString([]byte(l))

			updated, err := check(r.Context(), apiKey, l)
			if err != nil {
				log.Println(l, err)
				http.Error(w, "error", http.StatusInternalServerError)
				return
			}

			if !updated.IsZero() {
				feed.Entry = append(feed.Entry, &atom.Entry{
					Title:   "Update for " + l,
					ID:      "urn:street-view-updates:item:" + enc + ":" + updated.Format("20060102"),
					Link:    []atom.Link{{Href: "https://www.google.com/maps/place/" + url.QueryEscape(l)}},
					Content: &atom.Text{Type: "text", Body: l + " was updated " + updated.Format("2006-01-02")},
				})
			}
		}

		w.Header().Set("Content-Type", "application/atom+xml")

		enc := xml.NewEncoder(w)
		enc.Indent("", "  ")
		if err := enc.Encode(feed); err != nil {
			log.Println(ls, err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
	})

	addr := os.Getenv("ADDR")
	port := os.Getenv("PORT")
	if addr == "" {
		if port != "" {
			addr = ":" + port
		} else {
			addr = "127.0.0.1:8000"
		}
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Fatal(srv.ListenAndServe())
}

func check(ctx context.Context, apiKey, location string) (time.Time, error) {
	v := make(url.Values)
	v.Set("key", apiKey)
	v.Set("location", location)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://maps.googleapis.com/maps/api/streetview/metadata?"+v.Encode(), nil)
	if err != nil {
		return time.Time{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("bad HTTP status %d", resp.StatusCode)
	}

	var body struct {
		Status string
		Date   string
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return time.Time{}, err
	}

	switch body.Status {
	case "OK":
	case "ZERO_RESULTS":
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("bad body status %s", body.Status)
	}

	for _, l := range []string{"2006-01-02", "2006-01", "2006"} {
		t, err := time.Parse(l, body.Date)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("bad body date %s", body.Date)
}
