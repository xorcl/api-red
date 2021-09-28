package metronetwork

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
	"github.com/xorcl/api-red/common"
)

const STATUS_SELECTOR = "#estadoRed .row.pTop30 > .col-md-12 > .row"
const ESTACION_SELECTOR = ".estadoEstaciones > li"
const URL = "https://metro.cl/tu-viaje/estado-red"
const KeyValURL = "https://www.metro.cl/api/estadoRedDetalle.php"
const TimeURL = "https://www.metro.cl/api/horariosEstacion.php?cod=%s"
const HolidayURL = "https://apis.digital.gob.cl/fl/feriados/%d/%d/%d"

type Parser struct {
	HTTPRequest  http.Request
	StationTimes map[string]*CompositeTime
	IsHoliday    bool
}

func (bp *Parser) GetRoute() string {
	return "metro-network"
}

func (bp *Parser) StartParser() {
}

func (bp *Parser) Parse(c *gin.Context) {
	response := &Response{
		Lines: make([]*LineResponse, 0),
		Time:  time.Now().Format("2006-01-02 15:04:05"),
	}
	resp, err := http.Get(URL)
	if err != nil {
		logrus.Errorf("Error retrieving Metro page: %s", err)

		response.APIStatus = "Error al conectarse al sitio de Metro"
		c.JSON(400, &response)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		response.APIStatus = "Error al procesar sitio de Metro"
		logrus.Errorf("Error parsing Metro page: %s", err)
		c.JSON(400, &response)
		return
	}
	transfer_stations := make(map[string][]string)
	doc.Find(STATUS_SELECTOR).First().Children().Each(func(i int, s *goquery.Selection) {
		line := &LineResponse{
			Stations: make([]*StationResponse, 0),
		}
		lineNumber := strings.ToUpper(strings.TrimLeft(s.Find("strong").First().Text(), "Línea "))
		line.Name = "Línea " + lineNumber
		line.ID = "L" + lineNumber
		s.Find(ESTACION_SELECTOR).Each(func(i int, t *goquery.Selection) {
			description, exists := t.Attr("title")
			if !exists {
				return
			}
			class, exists := t.Attr("class")
			if !exists {
				return
			}
			status, exists := ToStatusCode[class]
			if !exists {
				return
			}
			if status != 0 {
				line.Issues = true
				response.Issues = true
			}
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(t.Nodes[0].FirstChild.Data), line.ID))
			var reason string
			if t.Nodes[0].FirstChild.NextSibling != nil {
				reason = strings.TrimSpace(t.Find(".popover-body").Text())
			}
			transfer_stations[name] = append(transfer_stations[name], line.ID)
			station := &StationResponse{
				Name:        name,
				ID:          slug.Make(strings.TrimSpace(name)),
				Status:      status,
				Description: strings.TrimSpace(description),
				Reason:      reason,
			}
			if time, ok := bp.StationTimes[station.ID]; ok {
				station.Schedule = time
				if closed, err := time.IsClosed(bp.IsHoliday); closed && err == nil {
					station.Status = 5
				}
			}
			line.Stations = append(line.Stations, station)
		})
		response.Lines = append(response.Lines, line)
	})
	response.APIStatus = "OK"
	// we just need to set the combinations
	for _, line := range response.Lines {
		for _, station := range line.Stations {
			if lines, ok := transfer_stations[station.Name]; ok {
				station.Lines = lines
			}
		}
	}
	c.JSON(200, &response)
}

func (bp *Parser) StopParser() {

}

func (p *Parser) GetCronTasks() []*common.CronTask {
	regex := regexp.MustCompile(`-l\da?$`)
	return []*common.CronTask{
		{
			Name: "Get all stations",
			Time: "0 1 * * *",
			Execute: func() error {
				logrus.Infof("checking station schedules...")
				kv := make(KeyValResponse)
				completeNames := make(map[string]string)
				resp, err := http.Get(KeyValURL)
				if err != nil {
					logrus.Errorf("Error retrieving Metro Status page: %s", err)
					return err
				}
				defer resp.Body.Close()
				err = json.NewDecoder(resp.Body).Decode(&kv)
				if err != nil {
					logrus.Errorf("Error parsing Metro Status body: %s", err)
					return err
				}
				for _, v := range kv {
					for _, station := range v.Estaciones {
						slug := strings.TrimSpace(slug.Make(station.Nombre))
						completeNames[station.Codigo] = regex.ReplaceAllString(slug, "")
					}
				}
				p.StationTimes = make(map[string]*CompositeTime)
				for code, name := range completeNames {
					logrus.Infof("checking schedule for %s...", name)
					p.StationTimes[name] = &CompositeTime{}
					sr := ScheduleResponse{}
					resp, err := http.Get(fmt.Sprintf(TimeURL, code))
					if err != nil {
						logrus.Errorf("Error retrieving %s station schedule page: %s", name, err)
						continue
					}
					defer resp.Body.Close()
					err = json.NewDecoder(resp.Body).Decode(&sr)
					if err != nil {
						logrus.Errorf("Error parsing Metro Status body: %s", err)
						continue
					}
					p.StationTimes[name].Open.Weekdays = strings.TrimSpace(sr.Estacion.Abrir.LunesViernes)
					p.StationTimes[name].Open.Saturday = strings.TrimSpace(sr.Estacion.Abrir.Sabado)
					p.StationTimes[name].Open.Holidays = strings.TrimSpace(sr.Estacion.Abrir.Domingo)

					p.StationTimes[name].Close.Weekdays = strings.TrimSpace(sr.Estacion.Cerrar.LunesViernes)
					p.StationTimes[name].Close.Saturday = strings.TrimSpace(sr.Estacion.Cerrar.Sabado)
					p.StationTimes[name].Close.Holidays = strings.TrimSpace(sr.Estacion.Cerrar.Domingo)
				}
				return nil
			},
		},
		{
			Name: "Is holiday today?",
			Time: "0 0 * * *",
			Execute: func() error {
				logrus.Infof("checking if today is holiday...")
				p.IsHoliday = false
				holidays := make([]struct{}, 0)
				now := time.Now()
				resp, err := http.Get(fmt.Sprintf(HolidayURL, now.Year(), now.Month(), now.Day()))
				if err != nil {
					logrus.Errorf("Error retrieving Gob Holiday API: %s", err)
					return err
				}
				defer resp.Body.Close()
				err = json.NewDecoder(resp.Body).Decode(&holidays)
				if err == nil && len(holidays) > 0 {
					p.IsHoliday = true
				}
				logrus.Infof("today is holiday: %t", p.IsHoliday)
				return nil
			},
		},
	}
}
