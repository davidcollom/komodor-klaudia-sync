package klaudia

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Infof(string, ...any)
	Debugf(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
}

type LogrusLogger struct {
	entry *logrus.Entry
}

func NewLogger(out io.Writer, logLevel string) *LogrusLogger {
	if out == nil {
		out = os.Stdout
	}
	logger := logrus.New()
	logger.SetOutput(out)
	switch logLevel {
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000000Z07:00",
	})
	return &LogrusLogger{entry: logrus.NewEntry(logger)}
}

func NewStdLogger(out io.Writer) *LogrusLogger { return NewLogger(out, "info") }

func (l *LogrusLogger) Infof(format string, args ...any)  { l.entry.Infof(format, args...) }
func (l *LogrusLogger) Debugf(format string, args ...any) { l.entry.Debugf(format, args...) }
func (l *LogrusLogger) Warnf(format string, args ...any)  { l.entry.Warnf(format, args...) }
func (l *LogrusLogger) Errorf(format string, args ...any) { l.entry.Errorf(format, args...) }
