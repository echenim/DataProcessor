package logger

import "log"

func Setup() {
}

func Error(args ...interface{}) {
	log.Println(args...)
}
