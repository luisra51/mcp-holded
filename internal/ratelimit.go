package internal

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type MultiLimiter struct {
	limiters []*rate.Limiter
}

func NewMultiLimiter(limiters ...*rate.Limiter) *MultiLimiter {
	return &MultiLimiter{limiters: limiters}
}

func (m *MultiLimiter) Wait(ctx context.Context) error {
	for _, limiter := range m.limiters {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// NewDefaultLimiter configures a conservative multi-bucket limiter.
// Tune to match the upstream API's published limits.
func NewDefaultLimiter() *MultiLimiter {
	perSecond := rate.NewLimiter(rate.Every(500*time.Millisecond), 2)
	perMinute := rate.NewLimiter(rate.Every(time.Minute/60), 60)
	perHour := rate.NewLimiter(rate.Every(time.Hour/10000), 100)
	return NewMultiLimiter(perSecond, perMinute, perHour)
}
