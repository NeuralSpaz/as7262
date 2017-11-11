package as7262

import (
	"log"

	"github.com/NeuralSpaz/i2cmux"
)

type AS7276 struct {
	dev *i2cmux.Device
}

func NewSensor(mux i2cmux.Multiplexer, port uint8, opts ...func(*AS7276) error) (*AS7276, error) {
	a := new(AS7276)

	for _, option := range opts {
		option(a)
	}
	var err error
	a.dev, err = i2cmux.Open(0x49, mux, port)
	if err != nil {
		log.Panic(err)
	}
	// a.setConfig()
	return a, nil
}
func (a *AS7276) Close() error {
	return a.dev.Close()
}
