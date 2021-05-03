package metronetwork

type Response struct {
	APIStatus string          `json:"api_status"`
	Issues    bool            `json:"issues"`
	Lines     []*LineResponse `json:"lines"`
}

type LineResponse struct {
	Name     string             `json:"name"`
	ID       string             `json:"id"`
	Issues   bool               `json:"issues"`
	Stations []*StationResponse `json:"stations"`
}

type StationResponse struct {
	Name        string     `json:"name"`
	ID          string     `json:"id"`
	Status      StatusCode `json:"status"`
	Description string     `json:"description,omitempty"`
}

type StatusCode int

var ToStatusCode = map[string]StatusCode{
	"estado1": 0,
	"estado2": 1,
	"estado3": 2,
	"estado4": 3,
}
