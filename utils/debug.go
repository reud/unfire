package utils

import (
	"io"
	"io/ioutil"

	"github.com/labstack/gommon/log"
)

// ref: https://mattn.kaoriya.net/software/lang/go/20171026101727.htm
func DebugResponse(r io.Reader) {

	bodyBytes, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Info(bodyString)

}
