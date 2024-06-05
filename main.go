package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	client := NewClient(cfg.PGK, cfg.FID)
	counters, err := client.Counters()
	if err != nil {
		log.Fatal(err)
	}

	counter := counters.Counter(cfg.Gym)

	if cfg.Storage != "" {
		storage := NewStorage(cfg.Storage)
		err := storage.Store(counter)
		if err != nil {
			log.Fatal(err)
		}
		if c, ok := storage.Last(); ok {
			counter = c
		}
	}

	switch count := counter.Count; count {
	case 1:
		fmt.Println("One person on the wall")
	default:
		fmt.Printf("%d people on the wall\n", count)
	}

	os.Exit(0)
}
