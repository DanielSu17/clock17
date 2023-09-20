package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

const (
	configFileName = ".17clock.yaml"
	calendarUrl    = "https://cdn.jsdelivr.net/gh/ruyut/TaiwanCalendar/data/"
)

type timeRange int

const (
	noTime timeRange = iota
	clockInTime
	clockOffTime
)

type calendar struct {
	Date        string `json:"date"`
	Week        string `json:"week"`
	IsHoliday   bool   `json:"isHoliday"`
	Description string `json:"description"`
}

func getCalendarUrl(year int) string {
	return calendarUrl + fmt.Sprintf("%d.json", year)
}

func getCalendar(year int, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	url := getCalendarUrl(year)
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

type answer struct {
	Employee  string
	Identity  string
	Latitude  string
	Longitude string
	Address   string
}

func initConfig() error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, configFileName)
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err == nil {
		return nil
	}

	var qs = []*survey.Question{
		{
			Name:   "employee",
			Prompt: &survey.Password{Message: "Enter Employee ID"},
		},
		{
			Name:   "identity",
			Prompt: &survey.Password{Message: "Enter Identity ID"},
		},
		{
			Name:   "latitude",
			Prompt: &survey.Password{Message: "Enter Latitude"},
		},
		{
			Name:   "longitude",
			Prompt: &survey.Password{Message: "Enter Longitude"},
		},
		{
			Name:   "address",
			Prompt: &survey.Password{Message: "Enter Address"},
		},
	}
	var ans answer
	if err := survey.Ask(qs, &ans); err != nil {
		return err
	}

	viper.Set("employee", ans.Employee)
	viper.Set("identity", ans.Identity)
	viper.Set("latitude", ans.Latitude)
	viper.Set("longitude", ans.Longitude)
	viper.Set("address", ans.Address)
	viper.SetConfigFile(path)
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	color.Yellow("Config saved to %s\n\n", viper.ConfigFileUsed())
	return nil
}

func isTodayHoliday() bool {
	year, month, day := time.Now().Date()
	calendars := []*calendar{}
	if err := getCalendar(year, &calendars); err != nil {
		color.Red(err.Error())
		return false
	}
	monthStr := strconv.Itoa(int(month))
	dayStr := strconv.Itoa(int(day))
	if len(monthStr) == 1 {
		monthStr = "0" + monthStr
	}
	if len(dayStr) == 1 {
		dayStr = "0" + dayStr
	}
	date := fmt.Sprintf("%d%s%s", year, monthStr, dayStr)

	isHoliday := false
	for _, calendar := range calendars {
		if calendar.Date == date {
			isHoliday = calendar.IsHoliday
			break
		}
	}
	return isHoliday
}

func getClockInOrClockOff() timeRange {
	hours, _, _ := time.Now().Clock()
	if hours >= 9 && hours < 11 {
		return clockInTime
	}
	if hours >= 18 && hours < 20 {
		return clockOffTime
	}
	return noTime
}
