package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/jprobinson/gtfs/nyct_subway_proto"
	"github.com/jprobinson/gtfs/transit_realtime"

	"code.google.com/p/goprotobuf/proto"
)

var (
	key    = flag.String("k", "", "mta.info API key")
	stop   = flag.String("stop_id", "L11", "mta.info subway stop id. (http://web.mta.info/developers/data/nyct/subway/google_transit.zip)")
	lTrain = flag.Bool("ltrain", true, "pull from L train feed. If false, pulls 1,2,3,4,5,6,S feed")
)

func main() {
	flag.Parse()

	url := fmt.Sprint("http://datamine.mta.info/mta_esi.php?key=", *key)
	if *lTrain {
		url = fmt.Sprint(url, "&feed_id=2")
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("unable to get mta url:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("unable to read body: ", err)
	}
	transit := &transit_realtime.FeedMessage{}
	err = proto.Unmarshal(body, transit)
	if err != nil {
		log.Fatal("unable to marshal proto: ", err)
	}

	mhtn := []*transit_realtime.TripUpdate_StopTimeUpdate{}
	bkln := []*transit_realtime.TripUpdate_StopTimeUpdate{}

	for _, ent := range transit.Entity {
		if ent.TripUpdate != nil {
			fmt.Println()
			for _, upd := range ent.TripUpdate.StopTimeUpdate {
				if strings.HasPrefix(*upd.StopId, *stop) {
					if strings.HasSuffix(*upd.StopId, "N") {
						mhtn = append(mhtn, upd)
					} else {
						bkln = append(bkln, upd)
					}
				}
			}
		}

	}

	fmt.Println("Manhattan bound trains:")
	minM := time.Now().AddDate(0, 1, 0)
	for _, upd := range mhtn {
		dept := time.Unix(*upd.Departure.Time, 0)
		fmt.Print(dept)
		if dept.Before(minM) {
			minM = dept
		}
		printUpdate(upd)
	}
	fmt.Println("Brooklyn bound trains:")

	minB := time.Now().AddDate(0, 1, 0)
	for _, upd := range bkln {
		dept := time.Unix(*upd.Departure.Time, 0)
		if dept.Before(minB) {
			minB = dept
		}
		printUpdate(upd)
	}

	fmt.Println("Next Brooklyn Bound Train Departes At ", minB)
	fmt.Println("Next Manhattan Bound Train Departes At ", minM)
}

func printUpdate(upd *transit_realtime.TripUpdate_StopTimeUpdate) {
	dept := time.Unix(*upd.Departure.Time, 0)
	delay := upd.Departure.Delay
	unc := upd.Departure.Uncertainty
	fmt.Print(*upd.StopId, ":\nExpected: ", dept,
		"\nDelay:", delay,
		"\nUncertainty:", unc,
		"\n\n")

}
