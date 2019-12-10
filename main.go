package main

import (
	"log"

	"github.com/orensimple/otus_events_reminder/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
