package dtf

type Configuration struct {
	DefaultIp string
	Ip        string `json:"ip,omitempty"`
}

type AppInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
