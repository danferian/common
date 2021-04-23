package logs

import (
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
)

type (
	Option struct {
		FilePath     string
		FileName     string
		Formatter    Formatter
		Stdout       bool
		ReportCaller bool
	}

	Formatter string
)

const (
	TextFormatter Formatter = "TEXT"
	JSONFormatter Formatter = "JSON"
)

func New(opt *Option) (*logrus.Logger, error) {
	if _, err := os.Stat(opt.FilePath); err != nil {
		if err := os.MkdirAll(opt.FilePath, os.ModePerm); err != nil {
			log.Fatalln(err)
		}
	}

	file, err := os.OpenFile(opt.FilePath+opt.FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	client := logrus.New()

	if opt.Formatter == JSONFormatter {
		client.SetFormatter(&logrus.JSONFormatter{})
	} else {
		client.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	client.SetOutput(file)
	if opt.Stdout {
		client.SetOutput(io.MultiWriter(file, os.Stdout))
	}

	client.SetReportCaller(opt.ReportCaller)

	return client, nil
}