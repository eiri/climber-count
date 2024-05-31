package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var fid, uid, gym string

func init() {
	// pgk
	flag.StringVar(&uid, "uid", os.Getenv("CC_UID"), "URL's uid. (Defaults to env var CC_UID)")
	flag.StringVar(&fid, "fid", os.Getenv("CC_FID"), "URL's fId. (Defaults to env var CC_FID)")
	flag.StringVar(&gym, "gym", os.Getenv("CC_GYM"), "Gym  name. (Defaults to env var CC_GYM)")
}

func main() {
	flag.Parse()

	client := NewClient(uid, fid)
	counters, err := client.Counters()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v\n", counters.Counter(gym))
}
