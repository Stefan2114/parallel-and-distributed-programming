package config

type Process struct {
	ID   int
	Port string
	Host string
}

type VariableConfig struct {
	ID          int
	OwnerID     int
	Subscribers []int
}

var AllProcesses = []Process{{ID: 0, Port: "8080", Host: "localhost"},
	{ID: 1, Port: "8081", Host: "localhost"},
	{ID: 2, Port: "8082", Host: "localhost"}}

var Variables = []VariableConfig{
	{ID: 0, OwnerID: 0, Subscribers: []int{0, 1, 2}},
	{ID: 1, OwnerID: 1, Subscribers: []int{0, 1}},
	{ID: 2, OwnerID: 2, Subscribers: []int{1, 2}},
}

func GetProcess(id int) Process {
	for _, p := range AllProcesses {
		if p.ID == id {
			return p
		}
	}
	return Process{}
}
