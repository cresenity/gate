package dtf

type Configuration struct {
	DefaultIp string
}

type AppInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
