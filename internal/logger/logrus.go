package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	PanicLevel = logrus.PanicLevel
	FatalLevel = logrus.FatalLevel
	ErrorLevel = logrus.ErrorLevel
	WarnLevel  = logrus.WarnLevel
	InfoLevel  = logrus.InfoLevel
	DebugLevel = logrus.DebugLevel
	TraceLevel = logrus.TraceLevel
)

type MyFormatter struct{}

func (mf *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	level := entry.Level
	strList := strings.Split(entry.Caller.File, "/")
	fileName := strList[len(strList)-1]
	b.WriteString(fmt.Sprintf("[%s] - %s - %s - [line:%d] - %s\n",
		strings.ToUpper(level.String()),
		entry.Time.Format("2006-01-02 15:04:05,678"),
		fileName,
		entry.Caller.Line,
		entry.Message))
	return b.Bytes(), nil
}

type Logger struct {
	Log     *logrus.Logger
	logFile *os.File
}

func NewLogger(level logrus.Level, writeToFile bool) (*Logger, error) {
	logger := logrus.New()
	logger.Level = level
	logger.SetReportCaller(true)
	logger.SetFormatter(&MyFormatter{})
	var logFile *os.File
	var w io.Writer

	if writeToFile {
		logFilePath := os.Getenv("LOG_FILE_PATH")
		if logFilePath == "" {
			logFilePath = "logs.txt" // Fallback на значение по умолчанию
		}
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}

		w = io.MultiWriter(os.Stdout, logFile)
	} else {
		w = os.Stdout
	}

	logger.SetOutput(w)

	return &Logger{
		Log:     logger,
		logFile: logFile,
	}, nil
}

func (l *Logger) Close() {
	l.logFile.Close()
}
