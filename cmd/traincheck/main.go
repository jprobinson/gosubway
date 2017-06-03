package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jprobinson/gosubway"
	"golang.org/x/net/context"
)

var (
	key    = flag.String("k", "", "mta.info API key")
	stop   = flag.String("stop_id", "L11", "mta.info subway stop id. (http://web.mta.info/developers/data/nyct/subway/google_transit.zip)")
	lTrain = flag.Bool("ltrain", true, "pull from L train feed. If false, pulls 1,2,3,4,5,6,S feed")
)

func main() {
	flag.Parse()

	feed, err := gosubway.GetFeed(context.Background(), *key, *lTrain)
	if err != nil {
		log.Fatal(err)
	}

	alerts, mhtn, bkln := feed.NextTrainTimes(*stop)
	nextB := bkln[0].Sub(time.Now())
	nextM := mhtn[0].Sub(time.Now())
	fmt.Println("Next Brooklyn Bound Train Departs From", *stop, "in", nextB)
	fmt.Println("Next Manhattan Bound Train Departs From", *stop, "in", nextM)
	if len(alerts) > 0 {
		for _, a := range alerts {
			fmt.Printf("There is an alert due to %s causing %s: %s\n",
				a.Cause.String(), a.Effect.String(), a.HeaderText.String())
		}
	}
}
