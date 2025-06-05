package parsetime

import (
	"github.com/mitchellh/mapstructure"
	"regexp"
)

type Time struct {
	Hours   int `mapstructure:"Hours"`
	Minutes int `mapstructure:"Minutes"`
	Seconds int `mapstructure:"Seconds"`
}

func ParseTime(s string) (*Time, error) {
	r, err := regexp.Compile(`(?:(?P<Hours>\d{2}):)?(?P<Minutes>\d{2}):(?P<Seconds>\d{2})`)
	if err != nil {
		return nil, err
	}

	res := r.FindStringSubmatch(s)
	resMap := make(map[string]string)
	for i, key := range r.SubexpNames() {
		resMap[key] = res[i]
	}

	time := &Time{}
	err = mapstructure.WeakDecode(resMap, &time)
	if err != nil {
		return nil, err
	}

	return time, nil
}

func ParseTimeToSeconds(s string) (int, error) {
	time, err := ParseTime(s)
	if err != nil {
		return 0, err
	}

	return time.Hours*3600 + time.Minutes*60 + time.Seconds, nil
}
