package busstop

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xorcl/api-red/common"
)

const BASE_URL = "http://web.smsbus.cl/web/buscarAction.do"
const SESSION_URL = BASE_URL + "?d=cargarServicios"

type Parser struct {
	Request       *http.Request
	Session       string
	BusStopRegexp *regexp.Regexp
}

func (bp *Parser) GetRoute() string {
	return "bus-stop/:stopid"
}

func (bp *Parser) StartParser() {
	req, err := http.NewRequest("GET", SESSION_URL, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"parser": "busstop-parser",
		}).Error("error starting parser: %s", err)
		return
	}
	bp.Request = req
	bp.getSession()
	bp.BusStopRegexp = regexp.MustCompile("^[Pp][A-Ja-j][0-9]{1,5}$")
}

func (bp *Parser) Parse(c *gin.Context) {
	bp.getSession() // TODO: Get the session only once
	stopID := c.Param("stopid")
	response := Response{
		Services: make([]*ServiceResponse, 0),
	}
	if stopID == "" {
		response.SetStatus(11)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("Missing Bus Stop ID")
		c.JSON(400, &response)
		return
	}
	isValid := bp.BusStopRegexp.MatchString(stopID)
	if !isValid {
		response.SetStatus(12)
		logrus.WithFields(logrus.Fields{
			"error":  response.StatusDescription,
			"stopID": stopID,
		}).Error("error parsing Bus Stop Schedule: Invalid Bus Stop Code Format")
		c.JSON(400, &response)
		return
	}
	response.ID = stopID
	form := url.Values{}
	form.Add("d", "busquedaParadero")
	form.Add("ingresar_paradero", stopID)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", BASE_URL, form.Encode()), nil)
	if err != nil {
		response.SetStatus(20)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Errorf("error creating Bus Stop Request: %s", err)
		c.JSON(400, &response)
		// bp.getSession()
		return
	}
	req.Header.Add("Cookie", fmt.Sprintf("JSESSIONID=%s", bp.Session))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		response.SetStatus(21)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Errorf("error parsing Bus Stop Schedule: %s", err)
		// bp.getSession()
		c.JSON(400, &response)
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		response.SetStatus(20)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bus Stop Schedule: %s", err)
		// bp.getSession()
		c.JSON(400, &response)
		return
	}
	var serviceNumber string
	response.ID, response.Name, response.StatusDescription, serviceNumber = getStopData(doc)
	if len(response.Name) == 0 {
		response.SetStatus(20)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bus Stop Schedule: Empty response")
		// bp.getSession()
		c.JSON(400, &response)
		return
	}
	if response.StatusDescription != "" && serviceNumber == "" {
		response.StatusCode = 30
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bus Stop Schedule: %s", err)
		// bp.getSession()
		c.JSON(400, &response)
		return
	}
	if serviceNumber == "" {
		response.Services = append(response.Services, getInvalidServices(doc)...)
		response.Services = append(response.Services, getValidServices(doc)...)
	} else {
		response.Services = append(response.Services, getSingleService(doc, serviceNumber))
	}
	response.SetStatus(0)
	c.JSON(200, &response)
}

func (bp *Parser) StopParser() {

}

func (bp *Parser) getSession() {
	client := http.Client{}
	resp, err := client.Do(bp.Request)
	if err != nil {
		logrus.Error("Cannot get Session: %s", err)
		return
	}
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			bp.Session = cookie.Value
			return
		}
	}
}

func (p *Parser) GetCronTasks() []*common.CronTask {
	return make([]*common.CronTask, 0)
}
