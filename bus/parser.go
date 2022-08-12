package bus

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
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
	response, err := http.Get(url)
	if err != nil {
		log.Printf("Error decoding info from external api for bus parser: %s", err)
		c.JSON(400, gin.H{"error": "No puedo obtener la información"})
		return
	}
	reader := response.Body
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")
	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, nil)
	response.Body.Close()
}

func (bp *Parser) StopParser() {

}

func (bp *Parser) getSession() {
	resp, err := http.Get(SESSION_URL)
	if err != nil {
		log.Printf("Error getting session for bus parser: %s", err)
		return
	}
	// read all body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading session page for bus parser: %s", err)
		return
	}
	resp.Body.Close()
	// find jwt
	jwtB64 := bp.BusStopRegexp.FindSubmatch(body)
	jwt, err := base64.StdEncoding.DecodeString(string(jwtB64[1]))
	if err != nil {
		log.Printf("Error decoding jwt for bus parser: %s", err)
	}
	bp.Session = string(jwt)
}

func (p *Parser) GetCronTasks() []*common.CronTask {
	return make([]*common.CronTask, 0)
}
