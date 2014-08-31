package gosubway

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/goprotobuf/proto"

	_ "github.com/jprobinson/gtfs/nyct_subway_proto"
	"github.com/jprobinson/gtfs/transit_realtime"
)

type FeedMessage struct {
	transit_realtime.FeedMessage
}

// GetFeed takes an API key generated from http://datamine.mta.info/user/register
// and a boolean specifying which feed (1,2,3,4,5,6,S trains OR L train) and
// it will return a transit_realtime.FeedMessage with NYCT extensions.
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
		return nil, err
	}
	return &FeedMessage{*transit}, nil
}

type StopTimeUpdate struct {
	transit_realtime.TripUpdate_StopTimeUpdate
}

// Trains will accept a stopId (found here: http://web.mta.info/developers/data/nyct/subway/google_transit.zip)
// and returns a list of updates from northbound and southbound trains
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

// NextTrain will return the duration until the next trains arrive
// at the given stop.
func (f *FeedMessage) NextTrains(stopId string) (northbound, southbound time.Duration) {
	north, south := f.Trains(stopId)
	northbound = NextTrain(north)
	southbound = NextTrain(south)
	return
}

// NextTrain will return the duration until the next train arrive
// given the update set.
func NextTrain(updates []*StopTimeUpdate) time.Duration {
	next := NextTrainTime(updates)
	return next.Sub(time.Now())
}

// NextTrainTime will return the time the next will train arrive
// given the update set.
func NextTrainTime(updates []*StopTimeUpdate) time.Time {
	// set to far future date so we can grab min date
	next := time.Now().AddDate(0, 1, 0)

	for _, upd := range updates {
		unix := *upd.Departure.Time
		if upd.Departure.Delay != nil {
			unix += int64(*upd.Departure.Delay)
		}
		dept := time.Unix(unix, 0)
		if dept.Before(next) && time.Now().Before(next) {
			next = dept
		}
	}
	return next
}
