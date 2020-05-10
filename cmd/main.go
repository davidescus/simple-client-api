package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davidescus/simple-client-api/pkg/nearearth"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	URL        = "https://api.nasa.gov/neo/rest/v1/feed"
	startTime  = "2020-01-01T00:00:00.00Z"
	timeFormat = time.RFC3339
)

type Config struct {
	// your key to access nasa api
	ApiKey string `env:"NASA_API_KEY"`
	// aggregate data by number of days
	GroupDaysNumber int `env:"GROUP_DAYS_NUMBER" env-default:"7"`
}

// IsValid will validate config before start
func (c *Config) IsValid() error {
	if c.ApiKey == "" {
		return errors.New("NASA_API_KEY should be registered as env var")
	}

	if c.GroupDaysNumber < 1 || c.GroupDaysNumber > 7 {
		return errors.New("GROUP_DAYS_NUMBER should be in range 1 to 7")
	}
	return nil
}

func main() {
	ctx := context.Background()
	logger := log.New(os.Stdout, "", 0)

	c := &Config{}
	err := cleanenv.ReadEnv(c)
	if err != nil {
		logger.Fatal(err)
	}
	if err = c.IsValid(); err != nil {
		logger.Fatal(err)
	}

	startDate, err := time.Parse(timeFormat, startTime)
	if err != nil {
		logger.Fatal(err)
	}

	nearEarthConf := &nearearth.Config{
		StartDate:       startDate,
		URL:             URL,
		ApiKey:          c.ApiKey,
		GroupDaysNumber: c.GroupDaysNumber,
	}
	ne := nearearth.New(ctx, nearEarthConf, logger)

	logger.Printf("\nPotentially hazardous asteroids since: %s", startDate.Format("2006-01-02"))

	// wait here till finish fetching all data
	ne.Run()

	for _, v := range ne.Out() {
		fmt.Println(v)
	}
}
