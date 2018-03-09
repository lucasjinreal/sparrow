package common

import (
	"log"
	"gitlab.com/jinfagang/colorgo"
)

func LogInit(prefix string) {
	log.SetPrefix(prefix)
	log.SetFlags(log.Ldate | log.Lshortfile)
}

func CheckError(err error) {
	if err != nil {
		cg.PrintlnRed("[Error]: " + err.Error())
	}
}

