package log

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
)

var (
	rootLogr = klogr.NewWithOptions()
	loggers  = map[context.Context]*logr.Logger{}
)

const HORIZON = "--------------------------------------------------------------------------------" // '-' * 80

func FromContext(ctx context.Context, names ...string) *logr.Logger {
	logger := loggers[ctx]
	if logger == nil {
		logger = &rootLogr
		for _, n := range names {
			newLogger := logger.WithName(n)
			logger = &newLogger
		}
		loggers[ctx] = logger
	}
	return logger
}
