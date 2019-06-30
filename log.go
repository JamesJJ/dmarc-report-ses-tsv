package main

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Debug *log.Logger
	Info  *log.Logger
	Error *log.Logger
)

func logInit(conf config) {

	errorHandle := os.Stderr
	infoHandle := os.Stdout

	debugHandle := ioutil.Discard
	if *conf.logVerbose {
		debugHandle = os.Stderr
	}

	Debug = log.New(debugHandle,
		"DEB: ",
		log.Ldate|log.Lmicroseconds|log.LUTC)

	Info = log.New(infoHandle,
		"INF: ",
		log.Ldate|log.Lmicroseconds|log.LUTC)

	Error = log.New(errorHandle,
		"ERR: ",
		log.Ldate|log.Lmicroseconds|log.LUTC)

	// no condition here, as you'll only see the message if
	// Verbose logging really is enabled!
	Debug.Printf("Verbose logging enabled")

}
