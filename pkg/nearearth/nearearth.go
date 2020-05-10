package nearearth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
)

const (
	apiDateFormat = "2006-01-02"
)

type Config struct {
	StartDate       time.Time
	URL             string
	ApiKey          string
	GroupDaysNumber int
}

type nearEarthCollector struct {
	ctx             context.Context
	logger          *log.Logger
	startDate       time.Time
	url             string
	apiKey          string
	groupDaysNumber int
	// because we want to order outputs
	out    []string
	isDone chan struct{}
}

type dateRange struct {
	start, end time.Time
}

type input struct {
	pos       int
	dateRange dateRange
}

type output struct {
	pos int
	val string
}

func New(ctx context.Context, conf *Config, logger *log.Logger) *nearEarthCollector {
	return &nearEarthCollector{
		ctx:             ctx,
		logger:          logger,
		startDate:       conf.StartDate,
		url:             conf.URL,
		apiKey:          conf.ApiKey,
		groupDaysNumber: conf.GroupDaysNumber,
		out:             []string{},
		isDone:          make(chan struct{}),
	}
}

func (n *nearEarthCollector) Run() {
	// max parallel request to 10, hope api will not bother
	dateRanges := createDateRanges(n.startDate, time.Hour*24*time.Duration(n.groupDaysNumber))
	mpc := len(dateRanges) / 3
	if mpc > 10 {
		mpc = 10
	}
	n.out = make([]string, len(dateRanges))

	// in, out for async requests
	in := make(chan input, mpc)
	out := make(chan output, mpc)

	// launch workers
	wg := sync.WaitGroup{}
	wg.Add(mpc)
	for i := 0; i < mpc; i++ {
		go func() {
			for dr := range in {
				out <- output{
					pos: dr.pos,
					val: n.processRequest(
						dr.dateRange.start.Format(apiDateFormat),
						dr.dateRange.end.Format(apiDateFormat),
					),
				}
			}
			wg.Done()
		}()
	}

	n.logger.Printf("\n --- start %d workers, want to collect results and presents them orderded by date ranges...", mpc)

	// write ranges into input chan
	go func() {
		for i, v := range dateRanges {
			in <- input{
				pos:       i,
				dateRange: v,
			}
		}
		close(in)
	}()

	// collect outputs, and order them
	go func() {
		for v := range out {
			n.out[v.pos] = v.val
		}
		close(out)
	}()

	wg.Wait()
}

func (n *nearEarthCollector) Out() []string {
	return n.out
}

func createDateRanges(s time.Time, r time.Duration) []dateRange {
	var ranges []dateRange

	now := time.Now()
	for s.Before(now) {
		end := s.Add(r)
		if end.After(now) {
			end = now
		}
		ranges = append(ranges, dateRange{
			start: s,
			end:   end,
		})
		s = end
	}

	return ranges
}

type Response struct {
	NearEarthObjects map[string][]NearObject `json:"near_earth_objects"`
}

type NearObject struct {
	IsPotentiallyHazardousAsteroid bool `json:"is_potentially_hazardous_asteroid"`
}

func (n *nearEarthCollector) processRequest(start, end string) string {
	client := client{}
	fullURL := fmt.Sprintf("%s?start_date=%s&end_date=%s&api_key=%s",
		n.url,
		url.PathEscape(start),
		url.PathEscape(end),
		n.apiKey,
	)

	resp, err := client.getData(fullURL)
	if err != nil {
		// TODO retries
		n.logger.Println(err)
	}

	if resp.statusCode != 200 {
		// TODO retries
		if resp.statusCode == 429 {
			// TODO exponential retries
		}
		n.logger.Printf("status code: %d expected 200", resp.statusCode)
	}

	var response Response
	err = json.Unmarshal(resp.body, &response)

	var potentiallyHazardous, notPotentiallyHazardous int
	for _, objects := range response.NearEarthObjects {
		for _, object := range objects {
			if object.IsPotentiallyHazardousAsteroid {
				potentiallyHazardous++
				continue
			}
			notPotentiallyHazardous++
		}
	}

	return fmt.Sprintf("From: %s to: %s\n   potentiallyHazardousObj: %d notPotentiallyHazardousObj: %d",
		start,
		end,
		potentiallyHazardous,
		notPotentiallyHazardous,
	)
}
