package as7262

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/NeuralSpaz/i2cmux"
)

type AS7276 struct {
	dev *i2cmux.Device
}

// Spectrum (450nm, 500nm, 550nm, 570nm, 600nm, 650nm).
type Spectrum struct {
	V, B, G, Y, O, R uint16
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
	a.setConfig()
	return a, nil
}
func (a *AS7276) Close() error {
	return a.dev.Close()
}

const (
	I2C_AS72XX_SLAVE_STATUS_REG = 0x00
	I2C_AS72XX_SLAVE_WRITE_REG  = 0x01
	I2C_AS72XX_SLAVE_READ_REG   = 0x02
	I2C_AS72XX_SLAVE_TX_VALID   = 0x02
	I2C_AS72XX_SLAVE_RX_VALID   = 0x01
)

func (a *AS7276) writeReg(reg byte, buf []byte) error {
	if err := a.checkTxPending(); err != nil {
		return err
	}

	// rx := make([]byte, 1)
	if err := a.dev.WriteReg(I2C_AS72XX_SLAVE_WRITE_REG, []byte{reg | 0x80}); err != nil {
		return err
	}

	if err := a.checkTxPending(); err != nil {
		return err
	}

	err := a.dev.WriteReg(I2C_AS72XX_SLAVE_WRITE_REG, buf)
	return err

}

func (a *AS7276) readReg(reg byte) (byte, error) {
	if err := a.checkTxPending(); err != nil {
		return 0, err
	}

	// rx := make([]byte, 1)
	if err := a.dev.WriteReg(I2C_AS72XX_SLAVE_WRITE_REG, []byte{reg | 0x80}); err != nil {
		return 0, err
	}

	if err := a.checkRxPending(); err != nil {
		return 0, err
	}
	buf := make([]byte, 1)
	err := a.dev.ReadReg(I2C_AS72XX_SLAVE_READ_REG, buf)
	return buf[0], err

}

func (a *AS7276) checkRxPending() error {
	for {
		log.Println("checking rx pending")
		rx := make([]byte, 1)
		if err := a.dev.ReadReg(I2C_AS72XX_SLAVE_STATUS_REG, rx); err != nil {
			return err
		}
		if rx[0]&I2C_AS72XX_SLAVE_RX_VALID == 0 {
			return nil
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func (a *AS7276) checkTxPending() error {
	for {
		log.Println("checking tx pending")
		rx := make([]byte, 1)
		if err := a.dev.ReadReg(I2C_AS72XX_SLAVE_STATUS_REG, rx); err != nil {
			return err
		}
		if rx[0]&I2C_AS72XX_SLAVE_TX_VALID == 0 {
			return nil
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func (a *AS7276) setConfig() error {
	fmt.Println("setConfig")
	if err := a.writeReg(0x04, []byte{0x60}); err != nil {
		return err
	}
	if err := a.writeReg(0x06, []byte{0xFF}); err != nil {
		return err
	}
	// LED OFF
	if err := a.writeReg(0x07, []byte{0x00}); err != nil {
		return err
	}
	return nil

}

func (a *AS7276) clearData() error {
	fmt.Println("clearData")
	value, err := a.readReg(0x04)
	if err != nil {
		return err
	}
	value = setBit(value, 1)
	if err := a.writeReg(0x04, []byte{value}); err != nil {
		return err
	}

	return nil
}

func (a *AS7276) setMode(mode uint8) error {
	fmt.Println("setmode")

	if mode > 3 {
		mode = 3
	}

	control, err := a.readReg(0x04)
	if err != nil {
		return err
	}
	control &= 0xf3
	control |= (mode << 2)
	if err := a.writeReg(0x04, []byte{control}); err != nil {
		return err
	}
	return nil
}

func (a *AS7276) dataReady() (bool, error) {
	fmt.Println("dataReady?")

	control, err := a.readReg(0x04)
	if err != nil {
		return false, err
	}
	return hasBit(control, 1), err

}

func (a *AS7276) ReadAll() (Spectrum, error) {
	fmt.Println("readall")
	err := a.clearData()
	if err != nil {
		log.Println(err)
	}
	err = a.setMode(3)
	if err != nil {
		log.Println(err)
	}
	ready, err := a.dataReady()
	if err != nil {
		log.Println(err)
	}
	for !ready {
		time.Sleep(time.Millisecond * 50)
		ready, err = a.dataReady()
		if err != nil {
			log.Println(err)
		}
	}

	// LED ON
	// if err := a.writeReg(0x07, []byte{0x09}); err != nil {
	// 	return Spectrum{}, err
	// }

	vh, err := a.readReg(0x08)
	if err != nil {
		return Spectrum{}, err
	}
	vl, err := a.readReg(0x09)
	if err != nil {
		return Spectrum{}, err
	}

	bh, err := a.readReg(0x0a)
	if err != nil {
		return Spectrum{}, err
	}
	bl, err := a.readReg(0x0b)
	if err != nil {
		return Spectrum{}, err
	}
	gh, err := a.readReg(0x0c)
	if err != nil {
		return Spectrum{}, err
	}
	gl, err := a.readReg(0x0d)
	if err != nil {
		return Spectrum{}, err
	}
	yh, err := a.readReg(0x0e)
	if err != nil {
		return Spectrum{}, err
	}
	yl, err := a.readReg(0x0f)
	if err != nil {
		return Spectrum{}, err
	}
	oh, err := a.readReg(0x10)
	if err != nil {
		return Spectrum{}, err
	}
	ol, err := a.readReg(0x11)
	if err != nil {
		return Spectrum{}, err
	}
	rh, err := a.readReg(0x12)
	if err != nil {
		return Spectrum{}, err
	}
	rl, err := a.readReg(0x13)
	if err != nil {
		return Spectrum{}, err
	}

	// LED OFF
	// if err := a.writeReg(0x07, []byte{0x00}); err != nil {
	// 	return Spectrum{}, err
	// }

	v := binary.LittleEndian.Uint16([]byte{vh, vl})
	b := binary.LittleEndian.Uint16([]byte{bh, bl})
	g := binary.LittleEndian.Uint16([]byte{gh, gl})
	y := binary.LittleEndian.Uint16([]byte{yh, yl})
	o := binary.LittleEndian.Uint16([]byte{oh, ol})
	r := binary.LittleEndian.Uint16([]byte{rh, rl})

	return Spectrum{v, b, g, y, o, r}, nil
}

func clearBit(n byte, pos uint8) byte {
	mask := ^(1 << pos)
	n &= byte(mask)
	return n
}
func setBit(n byte, pos uint8) byte {
	n |= (1 << pos)
	return n
}
func hasBit(n byte, pos uint8) bool {
	val := n & (1 << pos)
	return (val > 0)
}
