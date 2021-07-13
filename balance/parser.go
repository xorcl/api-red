package balance

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const BASE_URL = "https://cargatubip.metro.cl/CargaTuBipV2/"
const BALANCE_URL = BASE_URL + "faces/main.xhtml?formulario=formulario&formulario%%3Atxt_bip=%d&javax.faces.ViewState=%s&javax.faces.source=formulario%%3Abot_consultar_tarjeta&javax.faces.partial.event=click&javax.faces.partial.execute=formulario%%3Abot_consultar_tarjeta%%20formulario&javax.faces.partial.render=formulario&javax.faces.behavior.event=action&javax.faces.partial.ajax=true"

const VIEWSTATE_SELECTOR = "input[name=\"javax.faces.ViewState\"]"
const BALANCE_SELECTOR = ".card-body .align-middle .h5.text-right"

type Parser struct {
	ViewRequest *http.Request
	ViewState   string
}

func (bp *Parser) GetRoute() string {
	return "balance/:bipid"
}

func (bp *Parser) StartParser() {
	req, err := http.NewRequest("GET", BASE_URL, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"parser": "balance-parser",
		}).Error("error starting parser: %s", err)
		return
	}
	bp.ViewRequest = req
	bp.getViewState()
}

func (bp *Parser) Parse(c *gin.Context) {
        bp.getViewState() // TODO: Do this once
	response := Response{}
	if c.Param("bipid") == "" {
		response.SetStatus(11)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bip balance")
		c.JSON(400, &response)
		return
	}
	intBipID, err := strconv.Atoi(c.Param("bipid"))
	if err != nil {
		response.SetStatus(12)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bip balance: %s", err)
		c.JSON(400, &response)
		return
	}
	response.ID = c.Param("bipid")
	balanceURL := fmt.Sprintf(BALANCE_URL, intBipID, url.QueryEscape(bp.ViewState))
	resp, err := http.Get(balanceURL)
	if err != nil {
		response.SetStatus(21)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bip balance: %s", err)
		c.JSON(500, &response)
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		response.SetStatus(20)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bip balance: %s", err)
		// bp.getViewState()
		c.JSON(500, &response)
		return
	}
	node := doc.Find(BALANCE_SELECTOR).First()
	if node.Length() != 1 {
		response.SetStatus(22)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bip balance: Balance node not found")
		// bp.getViewState()
		c.JSON(500, &response)
		return
	}
	balanceStr := strings.Trim(strings.ReplaceAll(node.Text(), ",", ""), " $\t\n")
	balance, err := strconv.Atoi(balanceStr)
	if err != nil {
		response.SetStatus(23)
		logrus.WithFields(logrus.Fields{
			"error": response.StatusDescription,
		}).Error("error parsing Bip balance: %s", err)
		// bp.getViewState()
		c.JSON(500, &response)
		return
	}
	response.Balance = balance
	response.SetStatus(0)
	c.JSON(200, &response)
}

func (bp *Parser) StopParser() {
}

func (bp *Parser) getViewState() {
	client := http.Client{}
	resp, err := client.Do(bp.ViewRequest)
	if err != nil {
		logrus.Error("Cannot get ViewState: %s", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logrus.Error("Cannot get ViewState: status code %d", resp.StatusCode)
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logrus.Error("Cannot get ViewState: %s", err)
		return
	}
	node := doc.Find(VIEWSTATE_SELECTOR).First()
	if node.Length() != 1 {
		logrus.Error("Cannot get ViewState: Unable to find hidden input node")
		return
	}
	viewstate, ok := node.Attr("value")
	if !ok {
		logrus.Error("Cannot get ViewState: Unable to find value from hidden input node")
		return
	}
	bp.ViewState = viewstate
}
