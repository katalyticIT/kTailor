package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var currentLevel int

var levels = map[string]int{
	"DEBUG": 1,
	"INFO":  2,
	"WARN":  3,
	"ERROR": 4,
	"FATAL": 5,
}

// Init sets the desired log level from the configuration
func Init(level string) {
	lvl := strings.ToUpper(level)
	if val, ok := levels[lvl]; ok {
		currentLevel = val
	} else {
		currentLevel = 2 // Default to INFO
	}
}

// Logf formats and prints the message if the level is high enough
func Logf(level, format string, args ...interface{}) {
	lvl := strings.ToUpper(level)
	reqLevel, ok := levels[lvl]
	if !ok {
		reqLevel = 2 // Default
	}

	if reqLevel >= currentLevel {
		msg := fmt.Sprintf(format, args...)
		fmt.Printf("%s %s %s\n", time.Now().Format(time.RFC3339), lvl, msg)
		if lvl == "FATAL" {
			os.Exit(1)
		}
	}
}

// IsDebugEnabled helps us skip expensive operations (like JSON diffs) if DEBUG is off
func IsDebugEnabled() bool {
	return currentLevel <= levels["DEBUG"]
}
