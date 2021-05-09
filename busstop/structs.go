package busstop

type Response struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	StatusCode        ErrorCode          `json:"status_code"`
	StatusDescription string             `json:"status_description"`
	Services          []*ServiceResponse `json:"services"`
}

type ServiceResponse struct {
	ID                string         `json:"id"`
	Valid             bool           `json:"valid"`
	StatusDescription string         `json:"status_description"`
	Buses             []*BusResponse `json:"buses"`
}

type BusResponse struct {
	ID             string `json:"id"`
	MetersDistance int    `json:"meters_distance"`
	MinArrivalTime int    `json:"min_arrival_time"`
	MaxArrivalTime int    `json:"max_arrival_time"`
}

type ErrorCode int

var Errors = map[ErrorCode]string{
	0:  "Itinerario obtenido satisfactoriamente",
	10: "Error indeterminado al interpretar parámetro ID de Parada",
	11: "Parámetro ID de Parada faltante",
	12: "Parámetro ID de Parada mal formado",
	20: "Error indeterminado al parsear información desde SMSBus",
	21: "SMSBus no contesta",
	22: "SMSBus contesta, pero no entrega información interpretable",
	23: "Imposible interpretar valores de itinerario",
	30: "Error indeterminado con paradero",
}

func (br *Response) SetStatus(code ErrorCode) {
	br.StatusCode = code
	if description, ok := Errors[code]; ok {
		br.StatusDescription = description
	} else {
		br.StatusDescription = "Error indeterminado"
	}
}
