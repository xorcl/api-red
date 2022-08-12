package bus

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/xorcl/api-red/common"
)

const BASE_URL = "https://www.red.cl/predictor/prediccion?t=%s&codsimt=%s&codser="
const SESSION_URL = "https://www.red.cl/planifica-tu-viaje/cuando-llega/"

type Parser struct {
	Request       *http.Request
	Session       string
	BusStopRegexp *regexp.Regexp
}

func (bp *Parser) GetRoute() string {
	return "bus/:stopid"
}

func (bp *Parser) StartParser() {
	bp.BusStopRegexp = regexp.MustCompile("\\$jwt = '([A-Za-z0-9=-_]+)'")
}

func (bp *Parser) Parse(c *gin.Context) {
	bp.getSession() // TODO: Get the session only once
	stopID := c.Param("stopid")
	url := fmt.Sprintf(BASE_URL, bp.Session, stopID)
	fmt.Printf(url)
	response, _ := http.Get(url)
	reader := response.Body
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")
	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, nil)
}

func (bp *Parser) StopParser() {

}

func (bp *Parser) getSession() {
	resp, _ := http.Get(SESSION_URL)
	defer resp.Body.Close()
	// read all body
	body, _ := ioutil.ReadAll(resp.Body)
	// find jwt
	jwtB64 := bp.BusStopRegexp.FindSubmatch(body)
	jwt, _ := base64.StdEncoding.DecodeString(string(jwtB64[1]))
	bp.Session = string(jwt)
}

func (p *Parser) GetCronTasks() []*common.CronTask {
	return make([]*common.CronTask, 0)
}
