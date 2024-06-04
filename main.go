package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var pgk, fid, gym, storage string

func init() {
	flag.StringVar(&pgk, "pgk", os.Getenv("CC_PGK"), "URL's pgk. Can be set as env var CC_PGK.")
	flag.StringVar(&fid, "fid", os.Getenv("CC_FID"), "URL's fId. Can be set as env var  CC_FID.")
	flag.StringVar(&gym, "gym", os.Getenv("CC_GYM"), "Gym name. Can be set as env var CC_GYM.")
	flag.StringVar(&storage, "storage", "", "Path to storage file. Skips storing if empty.")
}

func main() {
	flag.Parse()

	requiredFlags := map[string]bool{
		"pgk": false,
		"fid": false,
		"gym": false,
	}

	flag.VisitAll(func(f *flag.Flag) {
		if _, ok := requiredFlags[f.Name]; ok && f.Value.String() != "" {
			requiredFlags[f.Name] = true
		}
	})

	missingFlags := []string{}
	for flagName, isSet := range requiredFlags {
		if !isSet {
			missingFlags = append(missingFlags, flagName)
		}
	}

	if len(missingFlags) > 0 {
		fmt.Fprintf(os.Stderr, "error: missing required flags: %v\n", missingFlags)
		flag.Usage()
		os.Exit(1)
	}

	client := NewClient(pgk, fid)
	counters, err := client.Counters()
	if err != nil {
		log.Fatal(err)
	}

	counter := counters.Counter(gym)

	if storage != "" {
		storage := NewStorage(storage)
		err := storage.Store(counter)
		if err != nil {
			log.Fatal(err)
		}
		if c, ok := storage.Last(); ok {
			counter = c
		}
	}

	fmt.Printf("%d people on the wall\n", counter.Count)

	os.Exit(0)
}
