package cachecluster

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/logging"
	"github.com/clockworksoul/smudge"
)

var logger *stackdriverLogging

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
func NewStackdriverLogging() (*stackdriverLogging, error) {
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

func (s *stackdriverLogging) Logger() *log.Logger {
	return s.logger.StandardLogger(logging.Info)
}

const logThreshhold = smudge.LogInfo

func (s *stackdriverLogging) Log(level smudge.LogLevel, a ...interface{}) (int, error) {
	if level >= logThreshhold {
		s.Debugf("%s %v", level.String(), a[0])
	}
	return 0, nil
}

func (s *stackdriverLogging) Logf(level smudge.LogLevel, format string, a ...interface{}) (int, error) {
	if level >= logThreshhold {
		s.Debugf(level.String()+" "+format, a...)
	}

	return 0, nil
}
