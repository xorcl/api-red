package metronetwork

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Response struct {
	APIStatus string          `json:"api_status"`
	Time      string          `json:"time"`
	Issues    bool            `json:"issues"`
	Lines     []*LineResponse `json:"lines"`
}

type LineResponse struct {
	Name                     string             `json:"name"`
	ID                       string             `json:"id"`
	Issues                   bool               `json:"issues"`
	StationsClosedBySchedule int                `json:"stations_closed_by_schedule"`
	Stations                 []*StationResponse `json:"stations"`
}

type StationResponse struct {
	Name               string         `json:"name"`
	ID                 string         `json:"id"`
	Status             StatusCode     `json:"status"`
	Lines              []string       `json:"lines,omitempty"`
	Description        string         `json:"description,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	IsClosedBySchedule bool           `json:"is_closed_by_schedule"`
	Schedule           *CompositeTime `json:"schedule"`
}

type KeyValResponse map[string]struct {
	Estaciones []struct {
		Nombre string `json:"nombre"`
		Codigo string `json:"codigo"`
	} `json:"estaciones"`
}

type DayResponse struct {
	LunesViernes string `json:"lunes_viernes"`
	Sabado       string `json:"sabado"`
	Domingo      string `json:"domingo"`
}
type OpenCloseResponse struct {
	Abrir  DayResponse `json:"abrir"`
	Cerrar DayResponse `json:"cerrar"`
}

type ScheduleResponse struct {
	Estacion OpenCloseResponse `json:"estacion"`
	//	Boleteria OpenCloseResponse `json:""`
}

type WeekTime struct {
	Weekdays string `json:"weekdays"`
	Saturday string `json:"saturday"`
	Holidays string `json:"holidays"`
}
type CompositeTime struct {
	Open  WeekTime `json:"open"`
	Close WeekTime `json:"close"`
}

func (ct *CompositeTime) IsClosed(isHoliday bool) (bool, error) {
	now := time.Now()
	var openStr, closeStr string
	var open, close time.Time
	var err error
	switch {
	case isHoliday || now.Weekday() == time.Sunday:
		openStr = ct.Open.Holidays
		closeStr = ct.Close.Holidays
	case now.Weekday() == time.Saturday:
		openStr = ct.Open.Holidays
		closeStr = ct.Close.Holidays
	default:
		openStr = ct.Open.Holidays
		closeStr = ct.Close.Holidays
	}
	open, err = time.ParseInLocation("15:04", openStr, now.Location())
	if err != nil {
		logrus.Errorf("error checking if holiday: %s", err)
		return false, err
	}
	open = open.AddDate(now.Year(), 0, now.YearDay()-1)
	close, err = time.Parse("15:04", closeStr)
	if err != nil {
		logrus.Errorf("error checking if holiday: %s", err)
		return false, err
	}
	close = close.AddDate(now.Year(), 0, now.YearDay()-1)
	if close.Before(open) {
		close = close.AddDate(0, 0, 1)
	}
	logrus.Infof("Current Date: %s", now.String())
	logrus.Infof("Open Date: %s", open.String())
	logrus.Infof("Close Date: %s", close.String())
	return now.Before(open) || now.After(close), nil
}

type StatusCode int

var ToStatusCode = map[string]StatusCode{
	"estado1": 0,
	"estado2": 1,
	"estado3": 2,
	"estado4": 3,
}
