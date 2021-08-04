// Package log provides uniform logging and tracing interfaces for ntt.  Other
// logging backends like Prometheus are hopefully easy to drop in.
//
// Note this package provides a reduced set of log-levels. Especially Warning,
// Error, Fatal, Panic functions are missing:
//
// * Warning: Nobody reads warnings, because by definition nothing went wrong.
// * Error: If you choose to handle the error by logging it, by definition it’s
//   not an error any more — you handled it. The act of logging an error handles
//   the error, hence it is no longer appropriate to log it as an error.
// * Fatal/Panic: It is commonly accepted that libraries should not use panic,
//   but if calling log.Fatal has the same effect, surely this should also be
//   outlawed.
//
// Taken from https://dave.cheney.net/2015/11/05/lets-talk-about-logging
package log

import (
	"context"
	"fmt"
	"os"
)

type Level int

const (
	PrintLevel Level = iota
	VerboseLevel
	DebugLevel
	TraceLevel
)

type Logger interface {
	Output(Level, string) error
}

type Tracer interface {
	// Start a span
	Start(ctx context.Context, spanName string) (context.Context, Span)
}

type Span interface {
	End()
}

var (
	std Logger = &ConsoleLogger{Out: os.Stderr}
	lvl        = PrintLevel
)

func GlobalLevel() Level         { return lvl }
func SetGlobalLevel(level Level) { lvl = level }
func SetGlobalLogger(l Logger)   { std = l }

func Print(v ...interface{})                 { std.Output(PrintLevel, fmt.Sprint(v...)) }
func Printf(format string, v ...interface{}) { std.Output(PrintLevel, fmt.Sprintf(format, v...)) }
func Println(v ...interface{})               { std.Output(PrintLevel, fmt.Sprintln(v...)) }

func Verbose(v ...interface{})                 { std.Output(VerboseLevel, fmt.Sprint(v...)) }
func Verbosef(format string, v ...interface{}) { std.Output(VerboseLevel, fmt.Sprintf(format, v...)) }
func Verboseln(v ...interface{})               { std.Output(VerboseLevel, fmt.Sprintln(v...)) }

func Debug(v ...interface{})                 { std.Output(DebugLevel, fmt.Sprint(v...)) }
func Debugf(format string, v ...interface{}) { std.Output(DebugLevel, fmt.Sprintf(format, v...)) }
func Debugln(v ...interface{})               { std.Output(DebugLevel, fmt.Sprintln(v...)) }

func Trace(v ...interface{})                 { std.Output(TraceLevel, fmt.Sprint(v...)) }
func Tracef(format string, v ...interface{}) { std.Output(TraceLevel, fmt.Sprintf(format, v...)) }
func Traceln(v ...interface{})               { std.Output(TraceLevel, fmt.Sprintln(v...)) }

func init() {
	if s := os.Getenv("NTT_DEBUG"); s != "" {
		lvl = DebugLevel
	}
	if s := os.Getenv("NTT_DEBUG_FILE"); s != "" {
		if file, err := os.Create(s); err == nil {
			SetGlobalLogger(&ConsoleLogger{Out: file})
		}
	}
	if s := os.Getenv("K3_DEBUG"); s != "" {
		lvl = DebugLevel
	}
	if s := os.Getenv("K3_DEBUG_FILE"); s != "" {
		if file, err := os.Create(s); err == nil {
			SetGlobalLogger(&ConsoleLogger{Out: file})
		}
	}
}
