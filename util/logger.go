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

func NewLogger(out io.Writer, err io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}

	if err == nil {
		err = os.Stderr
	}

	return &Logger{
		Out:     out,
		Err:     err,
		iLogger: log.New(out, "--> I ", log.Lmicroseconds),
		eLogger: log.New(err, "--> E ", log.Lmicroseconds),
	}
}
