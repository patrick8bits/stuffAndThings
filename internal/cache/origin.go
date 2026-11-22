package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrOriginNotFound = errors.New("origin object not found")

type Origin struct {
	objects map[string]string
	delay   time.Duration
}

// NewDemoOrigin creates our simulated content origin.
func NewDemoOrigin() *Origin {
	return &Origin{
		objects: map[string]string{
			"/video/a": "Video content A",
			"/video/b": "Video content B",
			"/video/c": "Video content C",
			"/video/d": "Video content D",
			"/video/e": "Video content E",
		},
		delay: 250 * time.Millisecond,
	}
}

// Fetch retrieves an object from the simulated origin.
func (o *Origin) Fetch(
	ctx context.Context,
	key string,
) (string, error) {

	value, found := o.objects[key]

	if !found {
		return "", fmt.Errorf(
			"%w: %s",
			ErrOriginNotFound,
			key,
		)
	}

	timer := time.NewTimer(o.delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return value, nil

	case <-ctx.Done():
		return "", ctx.Err()
	}
}
