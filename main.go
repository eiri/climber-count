package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var pgk, fid, gym string

func init() {
	flag.StringVar(&pgk, "pgk", os.Getenv("CC_PGK"), "URL's pgk. (Defaults to env var CC_PGK)")
	flag.StringVar(&fid, "fid", os.Getenv("CC_FID"), "URL's fId. (Defaults to env var CC_FID)")
	flag.StringVar(&gym, "gym", os.Getenv("CC_GYM"), "Gym  name. (Defaults to env var CC_GYM)")
}

func main() {
	flag.Parse()

	client := NewClient(pgk, fid)
	counters, err := client.Counters()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v\n", counters.Counter(gym))
}
