package polling

import (
	"github.com/saikitanabe/go-engine.io/transport"
)

var Creater = transport.Creater{
	Name:      "polling",
	Upgrading: false,
	Server:    NewServer,
	Client:    NewClient,
}
