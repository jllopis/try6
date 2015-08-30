/*
Package log is a wrapper around logrus that format an ordered slice of key/value
pairs and logs it through logrus.

The params must be of type interface{} and will be interpreted as follows:

- if a message must be printed, it will be the first item of the slice

- the rest of the elements will be treated as key/value pairs being the first the
key and the second the value. They will be added to a logrus.Entry Fields

- the keys mus be represented as a []byte or string and the values can be of any type

Ex:
  logI("i am a message", "somekey", "avalue", "otherkey", 23)

The resulting logrus.Entry will be logged using either logrus.Entry.Info, logrus.Entry.Error,
logrus.Entry.Fatal, logrus.Entry.Warn or logrus.Entry.Debug as for the method called.
*/
package log

import (
	"os"

	"github.com/Sirupsen/logrus"
)

// DefaultLogger is the logger used if no other is specified. It is a logrus.Logger
var DefaultLogger = logrus.New()

// init initialize the DefaultLogger with some good defaults
func init() {
	DefaultLogger.Level = logrus.InfoLevel
	DefaultLogger.Out = os.Stderr
	//DefaultLogger.Formatter = &logrus.JSONFormatter{}
	DefaultLogger.Formatter = &logrus.TextFormatter{DisableColors: true}
}

// SetLevel set the level of the DefaultLogger. Only messages equal or above the level
// will be showed.
func SetLevel(level logrus.Level) {
	DefaultLogger.Level = level
}

// LogI will log using logrus.Entry.Info
func LogI(m ...interface{}) {
	if len(m) == 1 {
		DefaultLogger.Info(m...)
		return
	}
	entry, msg := prepareEntry(m)
	entry.Info(msg)
}

// LogE will log using logrus.Entry.Error
func LogE(m ...interface{}) {
	if len(m) == 1 {
		DefaultLogger.Error(m...)
		return
	}
	entry, msg := prepareEntry(m)
	entry.Error(msg)
}

// LogF will log using logrus.Entry.Fatal
func LogF(m ...interface{}) {
	if len(m) == 1 {
		DefaultLogger.Fatal(m...)
		return
	}
	entry, msg := prepareEntry(m)
	entry.Fatal(msg)
}

// LogD will log using logrus.Entry.Debug
func LogD(m ...interface{}) {
	if len(m) == 1 {
		DefaultLogger.Debug(m...)
		return
	}
	entry, msg := prepareEntry(m)
	entry.Debug(msg)
}

// LogW will log using logrus.Entry.Warn
func LogW(m ...interface{}) {
	if len(m) == 1 {
		DefaultLogger.Warn(m...)
		return
	}
	entry, msg := prepareEntry(m)
	entry.Warn(msg)
}

// LogW will log using logrus.Entry.Warn
func LogP(m ...interface{}) {
	if len(m) == 1 {
		DefaultLogger.Panic(m...)
		return
	}
	entry, msg := prepareEntry(m)
	entry.Panic(msg)
}

// prepareEntry will parse the input params and build a logrus.Entry from them.
// It will return the *logrus.Entry and the message to be printed
func prepareEntry(m []interface{}) (*logrus.Entry, interface{}) {
	var msg interface{}
	data := logrus.Fields{}
	if len(m)%2 != 0 {
		msg = m[0]
		m = m[1:]
	}
	for i := 0; i < len(m); i = i + 2 {
		data[m[i].(string)] = m[i+1]
	}
	return logrus.NewEntry(DefaultLogger).WithFields(data), msg
}
