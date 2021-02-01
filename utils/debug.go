package utils

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/labstack/gommon/log"
)

// ref: https://mattn.kaoriya.net/software/lang/go/20171026101727.htm
func DebugResponse(r io.Reader) {
	tr := io.TeeReader(r, os.Stderr)
	bodyBytes, err := ioutil.ReadAll(tr)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Info(bodyString)
}
