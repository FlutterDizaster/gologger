package gologger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// TODO: Сделать возможность ведения нескольких журналов с указанием имени журнала
// TODO: Реализовать вывод оставшихся сообщений из очереди перед выходом из цикла

// LogLevel инициализируют уровни логирования.
type LogLevel int

const (
	LogLevelFatal   LogLevel = 1
	LogLevelError   LogLevel = 2
	LogLevelWarning LogLevel = 3
	LogLevelInfo    LogLevel = 4
	LogLevelDebug   LogLevel = 5
)

type message struct {
	datetime string
	level    string
	pkg      string
	fnc      string
	data     string
}

// Logger struct provides acces to logging functions.
// Use NewLogger() function to create Logger instance.
type Logger struct {
	logfile  string
	loglevel LogLevel
	messages chan message
	stopchan chan struct{}
}

// NewLogger() creates Logger instance.
// The path parameter used to define directory where log file will be created.
// The level parameter specify maximum logging level.
// level 1 = FATAL,
// level 2 = ERROR,
// level 3 = WARNING,
// level 4 = INFO,
// level 5 = DEBUG.
// If the level is less than 1, the Logger will not write any messages.
func NewLogger(path string, level LogLevel) *Logger {
	datetime := time.Now()
	filename := fmt.Sprintf("%d.%d.%d_%d:%d.log", datetime.Day(), datetime.Month(), datetime.Year(), datetime.Hour(), datetime.Minute())
	filename = filepath.Join(path, filename)

	logger := &Logger{ // Return a pointer to the Logger
		logfile:  filename,
		loglevel: level,
		messages: make(chan message, 256),
		stopchan: make(chan struct{}),
	}

	go logger.logWriter()

	return logger
}

func (l *Logger) logWriter() {
	logfile, err := os.Create(l.logfile)
	if err != nil {
		log.Println(err) // TODO: Edit this error handling
		return
	}

	defer func() {
		logfile.Close()
		close(l.stopchan)
	}()

	for {
		select {
		case msg := <-l.messages:
			msgstring := fmt.Sprintf("%s | %s | %s", msg.datetime, msg.level, msg.data)
			err := l.writeFile(*logfile, msgstring)
			if err != nil {
				log.Println(err) // TODO: Edit this error handling
				return
			}
			l.writeConsole(msgstring)
		case <-l.stopchan:
			return // Exit the loop when a stop signal is received
		}
	}
}

func (l *Logger) writeFile(logfile os.File, msgstring string) error {
	_, err := logfile.WriteString(msgstring + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (l *Logger) writeConsole(msgstring string) {
	log.Println(msgstring)
}

// Stop() stops Logger loop.
func (l *Logger) Stop() {
	close(l.messages)
	l.stopchan <- struct{}{}
}

// newMessage() creates a message struct and adds it to messages queue.
func (l *Logger) newMessage(level LogLevel, pkg string, fnc string, data string) {
	if level > l.loglevel {
		return
	}

	var levelstr string

	switch level {
	case LogLevelFatal:
		levelstr = "FATAL"
	case LogLevelError:
		levelstr = "ERROR"
	case LogLevelWarning:
		levelstr = "WARNING"
	case LogLevelInfo:
		levelstr = "INFO"
	case LogLevelDebug:
		levelstr = "DEBUG"
	default:
		levelstr = "CUSTOM"
	}

	datetime := time.Now()
	msgtime := fmt.Sprintf("%d.%d.%d | %d:%d:%d", datetime.Day(), datetime.Month(), datetime.Year(), datetime.Hour(), datetime.Minute(), datetime.Second())
	msg := message{
		datetime: msgtime,
		level:    levelstr,
		pkg:      pkg,
		fnc:      fnc,
		data:     data,
	}

	l.messages <- msg

	return
}

// Fatal() creates "FATAL" level log message with given parameters.
// The pkg parameter must contains package name, and
// the fnc parameter must contains function name where Fatal() called.
func (l *Logger) Fatal(pkg string, fnc string, err error) {
	l.newMessage(LogLevelFatal, pkg, fnc, err.Error())
}

// Error() creates "ERROR" level log message with given parameters.
// The pkg parameter must contains package name, and
// the fnc parameter must contains function name where Error() called.
func (l *Logger) Error(pkg string, fnc string, err error) {
	l.newMessage(LogLevelError, pkg, fnc, err.Error())
}

// Warning() creates "WARNING" level log message with given parameters.
// The pkg parameter must contains package name, and
// the fnc parameter must contains function name where Warning() called.
func (l *Logger) Warning(pkg string, fnc string, data string) {
	l.newMessage(LogLevelWarning, pkg, fnc, data)
}

// Info() creates "INFO" level log message with given parameters.
// The pkg parameter must contains package name, and
// the fnc parameter must contains function name where Info() called.
func (l *Logger) Info(pkg string, fnc string, data string) {
	l.newMessage(LogLevelInfo, pkg, fnc, data)
}

// Debug() creates "DEBUG" level log message with given parameters.
// The pkg parameter must contains package name, and
// the fnc parameter must contains function name where Debug() called.
func (l *Logger) Debug(pkg string, fnc string, data string) {
	l.newMessage(LogLevelDebug, pkg, fnc, data)
}
