package logger

import (
	"context"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	DefaultLogger    *logrus.Logger
	DefaultBaseEntry *logrus.Entry
	initOnce         sync.Once
)

// GetStringer describe an object that has capacity to return a string via GetString
// It uses as Gin Context in case we want to cut off gin dependency here
type GetStringer interface {
	GetString(key string) string
}

func Init(name string) {
	initOnce.Do(func() {
		DefaultLogger = logrus.New()
		if l, e := logrus.ParseLevel(os.Getenv("LOG_LEVEL")); e == nil {
			DefaultLogger.SetLevel(l)
		}
		if os.Getenv("LOG_FORMAT") == "json" {
			DefaultLogger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat:  "",
				DisableTimestamp: false,
				DataKey:          "",
				FieldMap:         nil,
				CallerPrettyfier: nil,
				PrettyPrint:      false,
			})
		}
		DefaultLogger.SetOutput(os.Stdout)
		DefaultBaseEntry = DefaultLogger.WithField("service", name)
	})
}

// Tag sets a tag name then returns a log entry ready to write
func Tag(tag string) *logrus.Entry {
	if DefaultBaseEntry == nil {
		Init("common")
	}
	return DefaultBaseEntry.WithField("tag", tag)
}

func WithTag(tag string) *logrus.Entry {
	l := Tag(tag)
	return l
}

func WithCtx(ctx context.Context, tag string) *logrus.Entry {
	l := Tag(tag)
	if requestID, ok := ctx.Value("x-request-id").(string); ok && requestID != "" {
		l = l.WithField("x-request-id", requestID)
	}
	return l
}

func WithField(key string, value interface{}) *logrus.Entry {
	return DefaultBaseEntry.WithField(key, value)
}

// LogError ghi log lỗi với thông điệp và chi tiết lỗi
func LogError(log *logrus.Entry, err error, message string) {
	log.WithError(err).Error("*** " + message + " ***")
}
