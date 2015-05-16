package gosubway

import (
	"io/ioutil"
	"net/http"
	"sort"
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
	url := "http://datamine.mta.info/mta_esi.php?key=" + key
	if lTrain {
		url = url + "&feed_id=2"
	} else {
		url = url + "&feed_id=1"
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
					} else if strings.HasSuffix(*upd.StopId, "S") {
						southbound = append(southbound, &StopTimeUpdate{*upd})
					}
				}
			}
		}

	}

	return
}

// NextTrainTimes will return an ordered slice of upcoming train departure times
// in either direction.
func (f *FeedMessage) NextTrainTimes(stopId string) (northbound, southbound []time.Time) {
	north, south := f.Trains(stopId)
	northbound = NextTrainTimes(north)
	southbound = NextTrainTimes(south)
	return
}

// NextTrainTimes will extract the departure times from the given
// update slice, order and return them.
func NextTrainTimes(updates []*StopTimeUpdate) []time.Time {
	var times []time.Time

	for _, upd := range updates {
		if upd.Departure == nil {
			continue
		}
		unix := *upd.Departure.Time
		if upd.Departure.Delay != nil {
			unix += int64(*upd.Departure.Delay * 1000)
		}
		dept := time.Unix(unix, 0)
		if dept.After(time.Now()) {
			times = append(times, dept)
		}
	}
	sort.Sort(timeSlice(times))
	if len(times) > 5 {
		times = times[:5]
	}
	return times
}

type timeSlice []time.Time

func (t timeSlice) Len() int {
	return len(t)
}

func (t timeSlice) Less(i, j int) bool {
	return t[i].Before(t[j])
}

func (t timeSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
