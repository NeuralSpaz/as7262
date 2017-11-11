package as7262

import (
	"log"

	"github.com/NeuralSpaz/i2cmux"
	"golang.org/x/exp/io/i2c"
)

type AS7276 struct {
	dev *i2cmux.Device
}

func NewSensor(i2cbus string, mux i2cmux.Multiplexer, port uint8, opts ...func(*AS7276) error) (*AS7276, error) {
	a := new(AS7276)

	for _, option := range opts {
		option(a)
	}
	var err error
	a.dev, err = i2cmux.Open(&i2c.Devfs{Dev: i2cbus}, 0x49, mux, port)
	if err != nil {
		log.Panic(err)
	}
	// a.setConfig()
	return a, nil
}
func (a *AS7276) Close() error {
	return a.dev.Close()
}
