package tango

import (
	"io"
	"time"

	"github.com/lunny/log"
)

type Logger interface {
	Debugf(format string, v ...interface{})
	Debug(v ...interface{})
	Infof(format string, v ...interface{})
	Info(v ...interface{})
	Warnf(format string, v ...interface{})
	Warn(v ...interface{})
	Errorf(format string, v ...interface{})
	Error(v ...interface{})
}

func NewLogger(out io.Writer) Logger {
	l := log.New(out, "[tango] ", log.Ldefault())
	l.SetOutputLevel(log.Ldebug)
	return l
}

type LogInterface interface {
	SetLogger(Logger)
}

type Log struct {
	Logger
}
func (l *Log) SetLogger(log Logger) {
	l.Logger = log
}

type Logging struct {
	logger Logger
}

func NewLogging(logger Logger) *Logging {
	return &Logging{
		logger: logger,
	}
}

func (l *Logging) Inject(o interface{}) {
	if g, ok := o.(LogInterface); ok {
		g.SetLogger(l.logger)
	}
}

func (itor *Logging) Handle(ctx *Context) {
	start := time.Now()
	itor.logger.Debug("Started", ctx.Req().Method,
		ctx.Req().URL.Path, "for", ctx.Req().RemoteAddr)

	if action := ctx.Action(); action != nil {
		if l, ok := action.(LogInterface); ok {
			l.SetLogger(itor.logger)
		}
	}

	ctx.Next()

	if ctx.Written() {
		statusCode := ctx.Status()
		requestPath := ctx.Req().URL.Path
		escape := time.Now().Sub(start)

		if statusCode >= 200 && statusCode < 400 {
			itor.logger.Info(ctx.Req().Method, statusCode, escape, requestPath)
		} else {
			itor.logger.Error(ctx.Req().Method, statusCode, escape, requestPath)
		}
	}
}
