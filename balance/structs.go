package balance

type Response struct {
	ID                string    `json:"id"`
	StatusCode        ErrorCode `json:"status_code"`
	StatusDescription string    `json:"status_description"`
	Balance           int       `json:"balance"`
	Error             string    `json:"error,omitempty"`
}

type ErrorCode int

var Errors = map[ErrorCode]string{
	0:  "Saldo obtenido satisfactoriamente",
	10: "Error indeterminado al interpretar parámetro Bip ID",
	11: "Parámetro Bip ID faltante",
	12: "Parámetro Bip ID mal formado",
	20: "Error indeterminado al parsear información desde RedBip",
	21: "RedBip no contesta",
	22: "RedBip contesta, pero no entrega información interpretable",
	23: "Imposible interpretar valor de saldo",
}

func (br *Response) SetStatus(code ErrorCode) {
	br.StatusCode = code
	if description, ok := Errors[code]; ok {
		br.StatusDescription = description
	} else {
		br.StatusDescription = "Error indeterminado"
	}
}
