package buildlog

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cragcraig/ccub/protos"
)

const (
	MonthLayout = "2006-Jan"
	DateLayout  = "2006-Jan-02"
)

var kitchenTimePattern = regexp.MustCompile(`^(\d+)(:\d\d)?(AM|PM|am|pm)$`)

type dateForm struct {
	layout string
	adjust func(time.Time) time.Time
}

func timeNoAdjust(t time.Time) time.Time {
	return t
}

func timeAddYear(t time.Time) time.Time {
	return t.AddDate(time.Now().Year(), 0, 0)
}

var validDateLayouts = []dateForm{
	{DateLayout, timeNoAdjust},
	{"2006-_1-_2", timeNoAdjust},
	{"Jan-_2", timeAddYear},
	{"1-_2", timeAddYear},
	{"1/_2", timeAddYear},
}

func ValidDateFormats() []string {
	valid := []string{"today", "yesterday"}
	for _, v := range validDateLayouts {
		valid = append(valid, strings.ReplaceAll(v.layout, "_", "0"))
	}
	return valid
}

func SameDayKitchenTimeDiff(end string, start string) (time.Duration, error) {
	tend, err := ParseKitchenTime(0, 0, 0, end)
	if err != nil {
		return 0, err
	}
	tstart, err := ParseKitchenTime(0, 0, 0, start)
	if err != nil {
		return 0, err
	}
	return tend.Sub(tstart), nil
}

func ParseKitchenTime(year int, month time.Month, day int, kitchen string) (time.Time, error) {
	submatches := kitchenTimePattern.FindStringSubmatch(kitchen)
	if submatches == nil {
		return time.Time{}, fmt.Errorf("Invalid time: %s", kitchen)
	}
	hours, _ := strconv.Atoi(submatches[1])
	minutes := 0
	if len(submatches[2]) > 0 {
		minutes, _ = strconv.Atoi(submatches[2][1:])
	}
	ampm := strings.ToUpper(submatches[3])
	if hours == 12 {
		hours = 0
	}
	if ampm == "PM" {
		hours += 12
	}
	if hours > 23 || minutes > 59 {
		return time.Time{}, fmt.Errorf("Invalid time: %s", kitchen)
	}
	return time.Date(year, month, day, hours, minutes, 0, 0, time.UTC), nil
}

func ParseDateArg(arg string) (time.Time, error) {
	if len(arg) > 0 {
		// Accept any partial spelling of "today" or "yesterday", e.g. "t" or "y"
		if arg == "today"[0:len(arg)] {
			return time.Now(), nil
		} else if arg == "yesterday"[0:len(arg)] {
			return time.Now().AddDate(0, 0, -1), nil
		}
	}
	for _, df := range validDateLayouts {
		if t, err := time.Parse(df.layout, arg); err == nil {
			return df.adjust(t), nil
		}
	}
	return time.Time{}, fmt.Errorf("Bad date %s, valid forms are:\n  %s", arg, strings.Join(ValidDateFormats(), "\n  "))
}

func ParseWorkPeriodsArg(year int, month time.Month, day int, arg string) ([]*protos.TimePeriod, error) {
	var periods []*protos.TimePeriod
	durations := strings.Split(arg, ",")
	for _, v := range durations {
		s := strings.Split(v, "-")
		if len(s) != 2 {
			return nil, fmt.Errorf("Time period must consist of both a start time and an end time")
		}
		start, err := ParseKitchenTime(year, month, day, s[0])
		if err != nil {
			return nil, fmt.Errorf("Bad start time: %s", s[0])
		}
		end, err := ParseKitchenTime(year, month, day, s[1])
		if err != nil {
			return nil, fmt.Errorf("Bad end time: %s", s[1])
		}
		if start.After(end) {
			return nil, fmt.Errorf("Start time %s is after end time %s", s[0], s[1])
		}
		periods = append(periods, &protos.TimePeriod{
			StartTime:   start.Format(time.Kitchen),
			EndTime:     end.Format(time.Kitchen),
			DurationMin: uint32(math.Ceil(end.Sub(start).Minutes())),
		})
	}
	return periods, nil
}

func ParseDateOfLog(log *protos.BuildLogEntry) (time.Time, error) {
	return time.Parse(DateLayout, log.Date)
}

func FormatDateForLog(date time.Time) string {
	return date.Format(DateLayout)
}
