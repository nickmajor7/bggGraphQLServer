// Package logger is the helper to write standardized log.
//
// It is designed to implement the new rules of logging,
// which tried it best to maintaining compatibility with v0 API and the rules.
//
// The package separates log's priority by four levels:
//
//	Error(0), Warn(1), Info(2) and Trace(3).
//
// User can use the vairables declared in the package to print log for each level.
//
//	logger.Info.Println("Hello world")
//	logger.Error.Println("Hello world")
//
// Users can filter logs by SetLevel function.
// If the level is set to 'n', only level smaller or equal to 'n' will
// print to standard output.
//
// The header of logger follows fixed formation in the referenced link.
// (http://wiki.emotibot.com:8090/pages/viewpage.action?pageId=10467739)
package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type transformer struct {
	w     io.Writer
	mu    *sync.Mutex
	Level string

	hasFF   bool
	showNum bool
}

var now = time.Now

// Pipeline
func (l *transformer) Write(p []byte) (n int, err error) {
	var (
		b   bytes.Buffer
		sep int
	)
	tm := now()
	// first part of header is almost fixed length
	// YYYY-MM-DD HH:MM:SS.SSS LEVEL 0 --- [0]
	b.Grow(40)
	// b.Write(formatTime(tm))
	b.Write([]byte(tm.Format(time.RFC3339Nano)))
	b.WriteString(" ")
	b.WriteString(l.Level)
	b.WriteString(" ")
	// b.WriteString("0 --- [0] ")
	// extract fileline from log header
	if l.hasFF {
		fi := bytes.IndexByte(p, ':')
		if fi > 0 {
			if si := bytes.IndexByte(p[fi+1:], ':'); si > 0 && l.showNum {
				sep = fi + 1 + si
				b.Write(p[:sep])
				// offset 1 by fixed empty space from log
				sep += 2
			} else {
				sep = fi
				b.Write(p[:sep])
				sep += si + 3
			}
		} else { // failed to extract fileline
			b.WriteString("???")
		}
	}
	b.WriteString(" : ")
	b.Write(p[sep:])
	// Before 1.14, defer has heavy overhead.
	l.mu.Lock()
	n, err = l.w.Write(b.Bytes())
	l.mu.Unlock()
	return n, err
}

func formatTime(time time.Time) []byte {
	b := make([]byte, 0, 23)
	local := time.Local()
	y, m, d := local.Date()
	itoa(&b, y, 4)
	b = append(b, '-')
	itoa(&b, int(m), 2)
	b = append(b, '-')
	itoa(&b, d, 2)
	b = append(b, ' ')
	h, min, s := local.Clock()
	itoa(&b, h, 2)
	b = append(b, ':')
	itoa(&b, min, 2)
	b = append(b, ':')
	itoa(&b, s, 2)
	b = append(b, '.')
	itoa(&b, local.Nanosecond()/1e6, 3)
	return b
}

// Trace is for debug only, any unsure behavior logging should go in there.
var Trace StdLogger

// Info is for informal log, logging for important bussiness logic and event.
var Info StdLogger

// Warn is for potential or recovered error.
var Warn StdLogger

// Error represents critical error, which can not be recovered
// and it should be notified to operator and developer.
var Error StdLogger

// The
const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelTrace
)

var (
	levelCount = 4
	threshold  = LevelInfo
)

func init() {
	SetLevel(threshold)
}

func setup(err, warn, info, trace io.Writer) {
	Error = log.New(err, "", log.Lshortfile)
	Warn = log.New(warn, "", log.Lshortfile)
	Info = log.New(info, "", log.Lshortfile)
	Trace = log.New(trace, "", log.Lshortfile)
}

// SetLevel will set the minimum standard log level print to os.StdOut and
// return the previous setting.
// Level should be one of predefined values:
//
//	LevelERROR, LevelWARN, LevelINFO, LevelTRACE.
//
// If level < 0, it does not change the setting.
// If level > LevelTrace, it will be changed to LevelTrace.
func SetLevel(level int) int {
	if level < 0 {
		return threshold
	} else if level > LevelTrace {
		level = LevelTrace
	}
	old := threshold
	showNum := level > LevelWarn
	threshold = level
	output := createOutputIO(threshold, showNum)
	setup(output[0], output[1], output[2], output[3])
	return old
}

func SetLevelString(level string) int {
	switch level {
	case "ERROR":
		return SetLevel(LevelError)
	case "WARN":
		return SetLevel(LevelWarn)
	case "INFO":
		return SetLevel(LevelInfo)
	case "TRACE":
		return SetLevel(LevelTrace)
	default:
		return SetLevel(LevelInfo)
	}
}

func createOutputIO(threshold int, showNum bool) []io.Writer {
	prefix := []string{"ERROR", "WARN", "INFO", "TRACE"}
	output := []io.Writer{io.Discard, io.Discard, io.Discard, io.Discard}
	var mu sync.Mutex
	for i := range output {
		if i <= threshold {
			output[i] = &transformer{
				w:       os.Stdout,
				mu:      &mu,
				Level:   prefix[i],
				hasFF:   true,
				showNum: showNum,
			}
		} else {
			break
		}
	}
	return output
}

// itoa copied from log package
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// StdLogger is the standard interface to define the required logging behavior.
// It is a direct referenced from logrus.StdLogger.
// This log package using it to encapsulate the details to avoid user accidentally change the setting.
type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})

	Output(int, string) error
}
