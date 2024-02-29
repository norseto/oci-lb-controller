package models

// LoadBalanceTarget represents a target for load balancing.
type LoadBalanceTarget struct {
	Name      string
	IpAddress string
	Port      int
	Weight    int
}
