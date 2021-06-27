package main

import (
	"github.com/kishaningithub/kafka-perf/cmd"
	"log"
)

func main() {
	log.SetFlags(0)
	cmd.Execute()
}
