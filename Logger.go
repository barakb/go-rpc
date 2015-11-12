package rpc
import (
	"log"
	"io"
	"fmt"
	"io/ioutil"
)

type logger struct {
	debug *log.Logger
	info  *log.Logger
}

type Logger interface {
    Debug(format string, v ...interface{})
    Info(format string, v ...interface{})
}


func (l *logger) Debug(format string, v ...interface{}) {
	l.debug.Output(2, fmt.Sprintf(format, v...))
}

func (l *logger) Info(format string, v ...interface{}) {
	l.info.Output(2, fmt.Sprintf(format, v...))
}

func NewLogger(target io.Writer) *logger {
	return &logger{
		debug : log.New(target,
			"DEBUG: ",
			log.Ldate | log.Ltime | log.Lshortfile),
		info: log.New(target,
			"INFO: ",
			log.Ldate | log.Ltime | log.Lshortfile),
	}
}

func NewEmptyLogger() *logger {
	return &logger{
		debug : log.New(ioutil.Discard,
			"DEBUG: ",
			0),
		info: log.New(ioutil.Discard,
			"INFO: ",
			0),
	}
}
