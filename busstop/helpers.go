package busstop

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Data Selector Constants
const STOP_ERROR = "#respuesta_error"

const STOP_ID_SELECTOR = "#numero_parada_respuesta .texto_h2"
const STOP_NAME_SELECTOR = "#nombre_paradero_respuesta"

// Multiple service stops

const SERVICE_OK_SELECTOR = "#proximo_solo_paradero, #siguiente_respuesta"
const SERVICE_ERROR_SELECTOR = "#servicio_error_solo_paradero"
const SERVICE_ERROR_DESCRIPTION_SELECTOR = "#respuesta_error_solo_paradero"

// Single Service Stops

const SINGLE_SERVICE_OK_SELECTOR = "#proximo_respuesta"
const SINGLE_SERVICE_ERROR_SELECTOR = "#respuesta_error"

// Service Values

const BUS_SERVICE_SELECTOR = "#servicio_respuesta_solo_paradero"
const BUS_ID_SELECTOR = "#bus_respuesta_solo_paradero, #proximo_bus_respuesta"
const BUS_TIME_SELECTOR = "#tiempo_respuesta_solo_paradero, #proximo_tiempo_respuesta"
const BUS_DISTANCE_SELECTOR = "#distancia_respuesta_solo_paradero, #proximo_distancia_respuesta"

var TimeRegex = regexp.MustCompile("[0-9]{1,2}")
var MenosRegex = regexp.MustCompile("[Mm]enos")

func getStopData(doc *goquery.Document) (stopID, stopName string, stopStatus, serviceNumber string) {
	stopID = strings.TrimSpace(doc.Find(STOP_ID_SELECTOR).First().Text())
	if len(doc.Find(STOP_ID_SELECTOR).Nodes) > 1 {
		serviceNumber = strings.TrimSpace(strings.Trim(doc.Find(STOP_ID_SELECTOR).Last().Text(), " .\t\n"))
	}
	stopName = strings.TrimSpace(strings.TrimPrefix(doc.Find(STOP_NAME_SELECTOR).First().Text(), "Paradero: "))
	stopStatus = strings.TrimSpace(strings.Trim(doc.Find(STOP_ERROR).First().Text(), " .\t\n"))
	return
}

func getInvalidServices(doc *goquery.Document) []*ServiceResponse {
	services := make([]*ServiceResponse, 0)
	doc.Find(SERVICE_ERROR_SELECTOR).Each(func(i int, s *goquery.Selection) {
		serviceID := strings.TrimSpace(s.Text())
		if len(serviceID) > 0 {
			services = append(services, &ServiceResponse{
				ID:    serviceID,
				Valid: false,
				Buses: make([]*BusResponse, 0),
			})
		}
	})
	doc.Find(SERVICE_ERROR_DESCRIPTION_SELECTOR).Each(func(i int, s *goquery.Selection) {
		if i < len(services) {
			services[i].StatusDescription = s.Text()
		}
	})
	return services
}

func getValidServices(doc *goquery.Document) []*ServiceResponse {
	services := make(map[string]*ServiceResponse)
	doc.Find(SERVICE_OK_SELECTOR).Each(func(i int, s *goquery.Selection) {
		serviceID := s.Find(BUS_SERVICE_SELECTOR).First().Text()
		if len(serviceID) > 0 {
			service, ok := services[serviceID]
			if !ok {
				service = &ServiceResponse{
					ID:                serviceID,
					Valid:             true,
					StatusDescription: "Servicio en Horario Hábil",
					Buses:             make([]*BusResponse, 0),
				}
				services[serviceID] = service
			}
			newBus, err := parseBus(s)
			if err == nil {
				service.Buses = append(service.Buses, newBus)
			}
		}
	})
	servicesArr := make([]*ServiceResponse, 0)
	for _, service := range services {
		servicesArr = append(servicesArr, service)
	}
	return servicesArr
}

func getSingleService(doc *goquery.Document, serviceID string) (service *ServiceResponse) {
	service = &ServiceResponse{
		ID:    serviceID,
		Valid: false,
		Buses: make([]*BusResponse, 0),
	}
	doc.Find(SINGLE_SERVICE_ERROR_SELECTOR).Each(func(i int, s *goquery.Selection) {
		error := strings.TrimSpace(s.Text())
		if len(error) > 0 {
			service.StatusDescription = error
		}
	})
	doc.Find(SINGLE_SERVICE_OK_SELECTOR).Each(func(i int, s *goquery.Selection) {
		if !service.Valid {
			service.Valid = true
			service.StatusDescription = "Servicio en Horario Hábil"
		}
		newBus, err := parseBus(s)
		if err == nil {
			service.Buses = append(service.Buses, newBus)
		}
	})
	return
}

func parseBus(s *goquery.Selection) (*BusResponse, error) {
	busID := strings.TrimSpace(s.Find(BUS_ID_SELECTOR).First().Text())
	distance, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(s.Find(BUS_DISTANCE_SELECTOR).First().Text()), " mts."))
	if err != nil {
		return nil, err
	}
	time := strings.TrimSpace(s.Find(BUS_TIME_SELECTOR).First().Text())
	timeCaptured := TimeRegex.FindAllString(time, 2)
	minTime := 0
	maxTime := 60 // Max minutes
	if len(timeCaptured) >= 2 {
		minTime, _ = strconv.Atoi(timeCaptured[0])
		maxTime, _ = strconv.Atoi(timeCaptured[1])
	} else if len(timeCaptured) == 1 {
		if MenosRegex.MatchString(time) {
			maxTime, _ = strconv.Atoi(timeCaptured[0])
		} else {
			minTime, _ = strconv.Atoi(timeCaptured[0])
		}
	} else {
		// Llegando = (0-3 min)
		maxTime = 3
	}
	return &BusResponse{
		ID:             busID,
		MetersDistance: distance,
		MinArrivalTime: minTime,
		MaxArrivalTime: maxTime,
	}, nil
}
