package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jprobinson/gosubway"
)

var (
	key    = flag.String("k", "", "mta.info API key")
	stop   = flag.String("stop_id", "L11", "mta.info subway stop id. (http://web.mta.info/developers/data/nyct/subway/google_transit.zip)")
	lTrain = flag.Bool("ltrain", true, "pull from L train feed. If false, pulls 1,2,3,4,5,6,S feed")
)

func main() {
	flag.Parse()

	feed, err := gosubway.GetFeed(*key, *lTrain)
	if err != nil {
		log.Fatal(err)
	}

	mhtn, bkln := feed.NextTrains(*stop)

	fmt.Println("Next Brooklyn Bound Train Departs From", *stop, "in", bkln)
	fmt.Println("Next Manhattan Bound Train Departs From", *stop, "in", mhtn)
}
