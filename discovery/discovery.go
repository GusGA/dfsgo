package discovery

import "time"

type Node struct {
	ServerID          string    `json:"server_id,omitempty" reder:"server_id"`
	IP                string    `json:"ip,omitempty" redis:"ip"`
	Hostmane          string    `json:"hostmane,omitempty" redis:"hostmane"`
	Port              int       `json:"port,omitempty" redis:"port"`
	CreatedAt         time.Time `json:"created_at,omitempty" redis:"created_at"`
	ConnectedToClient bool      `json:"connected_to_client,omitempty" redis:"connected_to_client"`
	Address           string    `json:"address,omitempty" reder:"address"`
}

type DiscoveryService interface {
	GetNodes() ([]Node, error)
	AddNode(string, Node) error
	// ReportDeadNode(Node) error
	Close() error
}
