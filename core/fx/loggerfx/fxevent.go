package loggerfx

import (
	"context"
	"log/slog"
	"strings"

	"go.uber.org/fx/fxevent"
)

type FxEvent string

const (
	FxEventOnStart    FxEvent = "onStart"
	FxEventOnStop     FxEvent = "onStop"
	FxEventSupplied   FxEvent = "supplied"
	FxEventProvided   FxEvent = "provided"
	FxEventReplaced   FxEvent = "replaced"
	FxEventDecorated  FxEvent = "decorated"
	FxEventInvoking   FxEvent = "invoking"
	FxEventInvoked    FxEvent = "invoked"
	FxEventRun        FxEvent = "run"
	FxEventStopping   FxEvent = "stopping"
	FxEventStopped    FxEvent = "stopped"
	FxEventRollback   FxEvent = "rollback"
	FxEventStarted    FxEvent = "started"
	FxEventLoggerInit FxEvent = "loggerInit"
)

type FxEventState string

const (
	FxStateDone  FxEventState = "done"
	FxStateError FxEventState = "error"
)

func (e FxEvent) String() string {
	return string(e)
}

func (e FxEvent) WithState(state FxEventState) string {
	return string(e) + "." + string(state)
}

func (e FxEvent) WithAttrs(args ...any) []any {
	for i, arg := range args {
		if key, ok := arg.(FxAttrKey); ok {
			args[i] = key.String()
		}
	}
	return args
}

type FxAttrKey string

const (
	FxAttrCallee      FxAttrKey = "callee"
	FxAttrCaller      FxAttrKey = "caller"
	FxAttrType        FxAttrKey = "type"
	FxAttrModule      FxAttrKey = "module"
	FxAttrConstructor FxAttrKey = "constructor"
	FxAttrDecorator   FxAttrKey = "decorator"
	FxAttrFunction    FxAttrKey = "function"
	FxAttrName        FxAttrKey = "name"
	FxAttrSignal      FxAttrKey = "signal"
	FxAttrError       FxAttrKey = "error"
)

func (k FxAttrKey) String() string {
	return string(k)
}

type FxSlogLogger struct {
	Logger *slog.Logger
}

func NewFxSlogLogger(logger *slog.Logger) fxevent.Logger {
	return &FxSlogLogger{Logger: logger}
}

func (l *FxSlogLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logInfo(FxEventOnStart, FxAttrCallee, e.FunctionName, FxAttrCaller, e.CallerName)
	case *fxevent.OnStartExecuted:
		l.logResult(FxEventOnStart, e.Err, FxAttrCallee, e.FunctionName, FxAttrCaller, e.CallerName)
	case *fxevent.OnStopExecuting:
		l.logInfo(FxEventOnStop, FxAttrCallee, e.FunctionName, FxAttrCaller, e.CallerName)
	case *fxevent.OnStopExecuted:
		l.logResult(FxEventOnStop, e.Err, FxAttrCallee, e.FunctionName, FxAttrCaller, e.CallerName)
	case *fxevent.Supplied:
		l.logDebugResult(FxEventSupplied, e.Err, FxAttrType, e.TypeName, FxAttrModule, e.ModuleName)
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.logDebug(FxEventProvided, FxAttrType, rtype, FxAttrModule, e.ModuleName, FxAttrConstructor, e.ConstructorName)
		}
		if e.Err != nil {
			l.logError(FxEventProvided, e.Err, FxAttrModule, e.ModuleName)
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.logInfo(FxEventReplaced, FxAttrType, rtype, FxAttrModule, e.ModuleName)
		}
		if e.Err != nil {
			l.logError(FxEventReplaced, e.Err, FxAttrModule, e.ModuleName)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.logInfo(FxEventDecorated, FxAttrType, rtype, FxAttrModule, e.ModuleName, FxAttrDecorator, e.DecoratorName)
		}
		if e.Err != nil {
			l.logError(FxEventDecorated, e.Err, FxAttrModule, e.ModuleName)
		}
	case *fxevent.Invoking:
		l.logDebug(FxEventInvoking, FxAttrFunction, e.FunctionName, FxAttrModule, e.ModuleName)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logError(FxEventInvoked, e.Err, FxAttrFunction, e.FunctionName, FxAttrModule, e.ModuleName)
		}
	case *fxevent.Run:
		l.logDebugResult(FxEventRun, e.Err, FxAttrName, e.Name, FxAttrModule, e.ModuleName)
	case *fxevent.Stopping:
		l.logInfo(FxEventStopping, FxAttrSignal, strings.ToUpper(e.Signal.String()))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logError(FxEventStopped, e.Err)
		}
	case *fxevent.RollingBack:
		l.logError(FxEventRollback, e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logError(FxEventRollback, e.Err)
		}
	case *fxevent.Started:
		l.logResult(FxEventStarted, e.Err)
	case *fxevent.LoggerInitialized:
		l.logDebugResult(FxEventLoggerInit, e.Err)
	}
}

func (l *FxSlogLogger) logInfo(event FxEvent, args ...any) {
	l.Logger.Log(context.Background(), slog.LevelInfo, event.String(), event.WithAttrs(args...)...)
}

func (l *FxSlogLogger) logDebug(event FxEvent, args ...any) {
	l.Logger.Log(context.Background(), slog.LevelDebug, event.String(), event.WithAttrs(args...)...)
}

func (l *FxSlogLogger) logDebugResult(event FxEvent, err error, args ...any) {
	if err != nil {
		l.logError(event, err, args...)
	} else {
		l.Logger.Log(context.Background(), slog.LevelDebug, event.WithState(FxStateDone), event.WithAttrs(args...)...)
	}
}

func (l *FxSlogLogger) logError(event FxEvent, err error, args ...any) {
	args = append(args, FxAttrError, err)
	l.Logger.Log(context.Background(), slog.LevelError, event.WithState(FxStateError), event.WithAttrs(args...)...)
}

func (l *FxSlogLogger) logResult(event FxEvent, err error, args ...any) {
	if err != nil {
		l.logError(event, err, args...)
	} else {
		l.Logger.Log(context.Background(), slog.LevelInfo, event.WithState(FxStateDone), event.WithAttrs(args...)...)
	}
}
