package util

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	Out, Err         io.Writer
	iLogger, eLogger *log.Logger
}

func (l *Logger) LogI(format string, args ...interface{}) {
	l.iLogger.Printf(format, args...)
}

func (l *Logger) LogE(format string, args ...interface{}) {
	l.eLogger.Printf(format, args...)
}

func (l *Logger) Construct() *Logger {
	if l.Out == nil {
		l.Out = os.Stdout
	}
	if l.Err == nil {
		l.Err = os.Stderr
	}
	l.iLogger = log.New(l.Out, "--> I ", log.Lmicroseconds)
	l.eLogger = log.New(l.Err, "--> E ", log.Lmicroseconds)
	return l
}
