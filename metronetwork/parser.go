package metronetwork

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

const STATUS_SELECTOR = "#estadoRed .row.pTop30 > .col-md-12 > .row"
const ESTACION_SELECTOR = ".estadoEstaciones > li"
const URL = "https://metro.cl/tu-viaje/estado-red"

type Parser struct {
	HTTPRequest http.Request
}

func (bp *Parser) GetRoute() string {
	return "metro-network"
}

func (bp *Parser) StartParser() {
}

func (bp *Parser) Parse(c *gin.Context) {
	response := &Response{
		Lines: make([]*LineResponse, 0),
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
			line.Stations = append(line.Stations, &StationResponse{
				Name:        name,
				ID:          slug.Make(strings.TrimSpace(name)),
				Status:      status,
				Description: strings.TrimSpace(description),
				Reason:      reason,
			})
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
