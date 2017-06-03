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
	line   = flag.String("line", "L", "train line; 1,2,3,4,5,6 or L")
)

func main() {
	flag.Parse()

	feed, err := gosubway.GetFeed(context.Background(), *key, *lTrain)
	if err != nil {
		log.Fatal(err)
	}

	alerts, mhtn, bkln := feed.NextTrainTimes(*stop, *line)
	if len(bkln) > 0 {
		nextB := bkln[0].Sub(time.Now())
		fmt.Println("Next Brooklyn Bound Train Departs From", *stop, "in", nextB)
	} else {
		fmt.Println("No Brooklyn Bound Trains For", *stop, "and", *line)
	}
	if len(mhtn) > 0 {
		nextM := mhtn[0].Sub(time.Now())
		fmt.Println("Next Manhattan Bound Train Departs From", *stop, "in", nextM)
	} else {
		fmt.Println("No Manhattan Bound Trains For", *stop, "and", *line)
	}
	if len(alerts) > 0 {
		for _, a := range alerts {
			fmt.Printf("There is an alert due to %s causing %s: %s\n",
				a.Cause.String(), a.Effect.String(), a.HeaderText.String())
		}
	}
}
