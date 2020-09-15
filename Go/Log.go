package main

import (
	"fmt"
	"time"
)

type LogLevel struct {
	level       int
	description string
}

var Debug = LogLevel{level: 0, description: "Debug"}
var Verbose = LogLevel{level: 1, description: "Verbose"}
var Info = LogLevel{level: 2, description: "Info"}
var Action = LogLevel{level: 3, description: "Action"}
var Warning = LogLevel{level: 4, description: "Warning"}
var Error = LogLevel{level: 5, description: "Error"}

func logDebug(header string, msg string) {
	log(Debug, header, msg)
}

func logVerbose(header string, msg string) {
	log(Verbose, header, msg)
}

func logInfo(header string, msg string) {
	log(Info, header, msg)
}

func logAction(header string, msg string) {
	log(Action, header, msg)
}

func logWarning(header string, msg string) {
	log(Warning, header, msg)
}

func logError(header string, msg string) {
	log(Error, header, msg)
}

var outputLevel LogLevel = Action

func log(level LogLevel, header string, msg string) {
	if level.level >= outputLevel.level {
		fmt.Printf("[%s]\t[%s]\t[%s]\t%s\n", time.Now(), level.description, header, msg)
	}
}
