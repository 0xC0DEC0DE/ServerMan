package api

type Server struct {
	ID           int    `json:"id"`
	Domain       string `json:"domain"`
	RegDate      string `json:"regdate"`
	BillingCycle string `json:"billingcycle"`
	NextDueDate  string `json:"nextduedate"`
	DomainStatus string `json:"domainstatus"`
}

type OsType struct {
	ID   int    `json:"id"`
	Name string `json:"operatingsystem"`
}

type AppType struct {
	ID   int    `json:"id"`
	App  string `json:"app"`
	Name string `json:"operatingsystem"`
}

type ServerDetail struct {
	Name            string `json:"name"`
	State           string `json:"state"`
	IpAddress       string `json:"ip"`
	OperatingSystem string `json:"operatingsystem"`
	Memory          string `json:"mem"`
	Disk            string `json:"disk"`
	Cpu             string `json:"cpu"`
	VncStatus       string `json:"vncstatus"`
	DailySnapshots  string `json:"dailysnapshots"`
}

type StatusMessage struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

type VncConsoleCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type ServerSnapshot struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at"`
	SizeGB    float64 `json:"size_gb"`
	Status    string  `json:"status"`
}

type RootPasswordResponse struct {
	Password string `json:"password"`
}

type ServerConfig struct {
	Domain   string
	Port     int
	Password string
}
