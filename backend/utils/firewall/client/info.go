package client

type FireInfo struct {
	Family   string
	Address  string
	Port     string
	Protocol string
	Strategy string
}

type Forward struct {
	Protocol string
	Address  string
	Port     string
	Target   string
}
