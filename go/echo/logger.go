package echo

import (
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"io"
)

type LoggerWrapper struct {
	*logrus.Logger
	prefix string
	level  log.Lvl
}

func (l *LoggerWrapper) Output() io.Writer {
	return l.Out
}

func (l *LoggerWrapper) Prefix() string {
	return l.prefix
}

func (l *LoggerWrapper) SetPrefix(p string) {
	l.prefix = p
}

func (l *LoggerWrapper) Level() log.Lvl {
	return l.level
}

func (l *LoggerWrapper) SetLevel(v log.Lvl) {
	l.level = v
}

func (l *LoggerWrapper) SetHeader(h string) {
}

func (l *LoggerWrapper) Printj(j log.JSON) {
	l.Printf("%v\n", j)
}

func (l *LoggerWrapper) Debugj(j log.JSON) {
	l.Debugf("%v\n", j)
}

func (l *LoggerWrapper) Infoj(j log.JSON) {
	l.Infof("%v\n", j)
}

func (l *LoggerWrapper) Warnj(j log.JSON) {
	l.Warnf("%v\n", j)
}

func (l *LoggerWrapper) Errorj(j log.JSON) {
	l.Errorf("%v\n", j)
}

func (l *LoggerWrapper) Fatalj(j log.JSON) {
	l.Fatalf("%v\n", j)
}

func (l *LoggerWrapper) Panicj(j log.JSON) {
	l.Panicf("%v\n", j)
}

func NewLoggerWrapper(logger *logrus.Logger) (*LoggerWrapper, error) {
	return &LoggerWrapper{logger, "echo", log.INFO}, nil
}
