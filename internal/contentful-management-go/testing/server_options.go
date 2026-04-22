package cmtesting

import (
	"errors"
	"time"
)

const defaultRateLimitPerSecond = 5

var (
	errRateLimitPerSecondMustBePositive = errors.New("rate limit per second must be greater than zero")
	errRateLimitNowMustNotBeNil         = errors.New("rate limit clock must not be nil")
)

type ServerOption func(*serverConfig) error

type serverConfig struct {
	rateLimitNow       func() time.Time
	rateLimitPerSecond int
}

func buildServerConfig(opts ...ServerOption) (serverConfig, error) {
	cfg := serverConfig{
		rateLimitNow:       time.Now,
		rateLimitPerSecond: defaultRateLimitPerSecond,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		err := opt(&cfg)
		if err != nil {
			return serverConfig{}, err
		}
	}

	return cfg, nil
}

func WithRateLimitNow(now func() time.Time) ServerOption {
	return func(cfg *serverConfig) error {
		if now == nil {
			return errRateLimitNowMustNotBeNil
		}

		cfg.rateLimitNow = now

		return nil
	}
}

func WithRateLimitPerSecond(limit int) ServerOption {
	return func(cfg *serverConfig) error {
		if limit <= 0 {
			return errRateLimitPerSecondMustBePositive
		}

		cfg.rateLimitPerSecond = limit

		return nil
	}
}
