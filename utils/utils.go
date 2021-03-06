package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Validator interface {
	Validate() error
}

type Month struct {
	Name string
	Days int
}

const (
	SimpleTimeFormat = "15:04:05"
	FullDateFormat   = "02/01/06 03:04:05 PM"
)

var monthDays = [12]Month{
	{"January", 31},
	{"February", 28},
	{"March", 31},
	{"April", 30},
	{"May", 31},
	{"June", 30},
	{"July", 31},
	{"August", 31},
	{"September", 30},
	{"October", 31},
	{"November", 30},
	{"December", 30},
}

func LoadFromViper(viperSession *viper.Viper, envVarPrefix string, configurationToSet Validator, defaultConfiguration Validator) (err error) {
	// Load Defaults
	var defaults map[string]interface{}
	err = mapstructure.Decode(defaultConfiguration, &defaults)
	if err != nil {
		return
	}
	err = viperSession.MergeConfigMap(defaults)
	if err != nil {
		return
	}

	// Load Environment variables
	viperSession.SetEnvPrefix(envVarPrefix)
	viperSession.AllowEmptyEnv(false)
	viperSession.AutomaticEnv()
	viperSession.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Merge together all the sources and unmarshal into struct
	if err := viperSession.Unmarshal(configurationToSet); err != nil {
		return fmt.Errorf("unable to decode config into struct, %w", err)
	}

	// Run validation
	err = configurationToSet.Validate()
	return
}

func BindFlagToEnvironmentVariable(viperSession *viper.Viper, envVarPrefix string, envVar string, flag *pflag.Flag) (err error) {
	err = viperSession.BindPFlag(envVar, flag)
	if err != nil {
		return
	}
	trimmed := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(envVar), strings.ToLower(envVarPrefix)), "_")
	err = viperSession.BindPFlag(trimmed, flag)
	if err != nil {
		return
	}
	err = viperSession.BindPFlag(strings.ReplaceAll(trimmed, "_", "."), flag)
	if err != nil {
		return
	}
	err = viperSession.BindPFlag(strings.ReplaceAll(envVar, "_", "."), flag)
	return
}

func IsValidDate(s string) bool {
	re := regexp.MustCompile(`(^((0[1-9]|[12]\d|3[01])\/(0[13578]|1[02]))|((0[1-9]|[12]\d|30)\/(0[13456789]|1[012]))|((0[1-9]|1\d|2[0-8])\/02)|(29\/02))$`)
	return re.MatchString(s)
}

func RemoveChars(s string, chars []string) (newS string) {
	newS = s
	for _, char := range chars {
		newS = strings.ReplaceAll(newS, char, "")
	}
	return
}

func InHourInterval(n int, timeToCheck time.Time) bool {
	startTime := fmt.Sprintf("%s:00:00", AppendZero(n))
	endTime := fmt.Sprintf("%s:00:00", AppendZero(n+1))

	start, err := time.Parse(SimpleTimeFormat, startTime)
	if err != nil {
		return false
	}
	end, err := time.Parse(SimpleTimeFormat, endTime)
	if err != nil {
		return false
	}
	check, err := time.Parse(SimpleTimeFormat, timeToCheck.Format(SimpleTimeFormat))
	if err != nil {
		return false
	}

	return !check.Before(start) && !check.After(end)
}

func SplitCommand(input string) []string {
	s := strings.Split(input, " ")
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func AppendZero(i int) string {
	if i < 10 {
		return fmt.Sprintf("0%d", i)
	}
	return fmt.Sprintf("%d", i)
}

func ConvertYearDayToDate(day string) (date string, err error) {
	count, err := strconv.Atoi(day)
	if err != nil {
		return "", fmt.Errorf("error parsing day as date %w", err)
	}
	for _, month := range monthDays {
		count -= month.Days
		if count < 0 {
			day := count + month.Days
			date := fmt.Sprintf("%s %d", month.Name, day)
			return date, nil
		}
	}
	return "", fmt.Errorf("error parsing day as date %w", err)
}

func DaysInThisYear() int {
	y := time.Now().Year()
	if (y%4 == 0 && y%100 != 0) || y%400 == 0 {
		return 366
	}
	return 365
}

func Contains(arr interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(arr)
	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

func AddNumSuffix(i int) string {
	switch i % 10 {
	case 1:
		if i == 11 {
			return fmt.Sprintf("%dth", i)
		}
		return fmt.Sprintf("%dst", i)
	case 2:
		if i == 12 {
			return fmt.Sprintf("%dth", i)
		}
		return fmt.Sprintf("%dnd", i)
	case 3:
		if i == 13 {
			return fmt.Sprintf("%dth", i)
		}
		return fmt.Sprintf("%drd", i)
	default:
		return fmt.Sprintf("%dth", i)
	}
}
