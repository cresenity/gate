package dtf

type Response struct {
	Status  bool        `json:"status" default:"true"`
	Code    int         `json:"code,omitempty" default:"0"`
	Message string      `json:"message,omitempty" default:""`
	Data    interface{} `json:"data,omitempty"`
}
