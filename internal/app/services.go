package app

// relays contains all available webhook destinations keyed by service name.
var relays = map[string]RelayFunc{
	"alertzy":  sendToAlertzy,
	"pushover": sendToPushover,
}
