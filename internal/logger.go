package internal

import (
	"log"
	"os"
)

var (
	Access = log.New(os.Stdout, "", log.LstdFlags)
	Error  = log.New(os.Stderr, "", log.LstdFlags)
)
