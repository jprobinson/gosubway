package gosubway

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/jprobinson/gtfs/transit_realtime"
)

type FeedMessage struct {
	transit_realtime.FeedMessage
}

func GetFeed(key string, lTrain bool) (*FeedMessage, error) {
	url := fmt.Sprint("http://datamine.mta.info/mta_esi.php?key=", key)
	if lTrain {
		url = fmt.Sprint(url, "&feed_id=2")
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	transit := &transit_realtime.FeedMessage{}
	err = proto.Unmarshal(body, transit)
	if err != nil {
		log.Fatal("unable to marshal proto: ", err)
	}
	return &FeedMessage{*transit}, nil
}

type StopTimeUpdate struct {
	transit_realtime.TripUpdate_StopTimeUpdate
}

func (f *FeedMessage) Trains(stopId string) (northbound, southbound []*StopTimeUpdate) {

	for _, ent := range f.Entity {
		if ent.TripUpdate != nil {
			for _, upd := range ent.TripUpdate.StopTimeUpdate {
				if strings.HasPrefix(*upd.StopId, stopId) {
					if strings.HasSuffix(*upd.StopId, "N") {
						northbound = append(northbound, &StopTimeUpdate{*upd})
					} else {
						southbound = append(southbound, &StopTimeUpdate{*upd})
					}
				}
			}
		}

	}

	return
}

func NextTrain(updates []*StopTimeUpdate) time.Duration {
	next := NextTrainTime(updates)
	return next.Sub(time.Now())
}

func NextTrainTime(updates []*StopTimeUpdate) time.Time {
	// default to a month in the future
	next := time.Now().AddDate(0, 1, 0)

	for _, upd := range updates {
		unix := *upd.Departure.Time
		if upd.Departure.Delay != nil {
			unix += int64(*upd.Departure.Delay)
		}
		dept := time.Unix(unix, 0)
		if dept.Before(next) {
			next = dept
		}
	}
	return next
}
