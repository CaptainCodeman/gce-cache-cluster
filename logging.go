package cachecluster

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
)

var logger Logging

type (
	// Logging provides logging abstraction
	Logging interface {
		Debugf(format string, args ...interface{})
	}

	stackdriverLogging struct {
		client *logging.Client
		logger *logging.Logger
	}
)

func init() {
	logger, _ = NewStackdriverLogging()
}

// NewStackdriverLogging creates a new stackdriver logger
func NewStackdriverLogging() (Logging, error) {
	ctx := context.Background()

	project, _ := metadata.ProjectID()
	client, err := logging.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}

	logger := client.Logger("cache-cluster")

	return &stackdriverLogging{
		client: client,
		logger: logger,
	}, nil
}

func (s *stackdriverLogging) Debugf(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	s.logger.Log(logging.Entry{Severity: logging.Debug, Payload: text})
}
