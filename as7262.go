package as7262

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/NeuralSpaz/i2cmux"
	// "github.com/NeuralSpaz/i2cmux"
)

type AS7262 struct {
	dev *i2cmux.Device
	// dev   *i2c.Device
	debug bool
}

// Spectrum (450nm, 500nm, 550nm, 570nm, 600nm, 650nm).
type Spectrum struct {
	Counts   []Count   `json:"count"`
	Time     time.Time `json:"time"`
	Unixnano int64     `json:"unixnano"`
}

type Count struct {
	Wavelength float64 `json:"wavelength"`
	Value      float64 `json:"value"`
	Raw        uint16  `json:"raw"`
}

func NewSensor(mux i2cmux.Multiplexer, port uint8, opts ...func(*AS7262) error) (*AS7262, error) {
	a := new(AS7262)

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

// func NewSensor(bus string, opts ...func(*AS7262) error) (*AS7262, error) {
// 	a := new(AS7262)

// 	for _, option := range opts {
// 		option(a)
// 	}
// 	var err error
// 	// a.dev, err := i2c.Open(bus, 0x49)
// 	a.dev, err = i2c.Open(&i2c.Devfs{Dev: bus}, 0x49)
// 	// a.dev, err = i2cmux.Open(0x49, mux, port)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	// a.dev = *dev
// 	a.setConfig()
// 	a.debug = true
// 	return a, nil
// }

func (a *AS7262) virtualRegisterWrite(register, data byte) error {
	// if a.debug {
	// log.Printf("virtualRegisterWrite(%02x,%02x)\n", register, data)
	// }
	const (
		SlaveStatusRegister byte = 0x00
		SlaveWriteRegister  byte = 0x01
		SlaveReadRegister   byte = 0x02
	)
	for {
		// if a.debug {
		// log.Printf("Checking SlaveStatusRegister\n")
		// }
		rx := make([]byte, 1)
		if err := a.dev.ReadReg(SlaveStatusRegister, rx); err != nil {
			log.Fatalln(err)
		}
		// if a.debug {
		// log.Printf("Status Register Contents: %02x\n", rx[0])
		// }
		if rx[0]&0x03 == 0x00 {
			break
		}
		if rx[0]&0x03 == 0x01 {
			discard := make([]byte, 1)
			if err := a.dev.ReadReg(SlaveStatusRegister, discard); err != nil {
				log.Fatalln(err)
			}
			// if a.debug {
			// log.Printf("DataLeftInReadBuffer: %02x\n", discard[0])
			// }
		}
		time.Sleep(time.Millisecond)
	}
	// if a.debug {
	// log.Printf("Checking SlaveStatusRegister\n")
	// }
	// if a.debug {
	// log.Printf("WriteReg(%02x,%02x)\n", SlaveWriteRegister, register|0x80)
	// }
	if err := a.dev.WriteReg(SlaveWriteRegister, []byte{register | 0x80}); err != nil {
		log.Fatalln(err)
	}

	for {
		// if a.debug {
		// log.Printf("Checking SlaveStatusRegister\n")
		// }
		rx := make([]byte, 1)
		if err := a.dev.ReadReg(SlaveStatusRegister, rx); err != nil {
			log.Fatalln(err)
		}
		// if a.debug {
		// log.Printf("Status Register Contents: %02x\n", rx[0])
		// }
		if rx[0]&0x03 == 0x00 {
			break
		}

		time.Sleep(time.Millisecond)
	}
	// if a.debug {
	// log.Printf("Writing Data to Slave\n")
	// }
	if err := a.dev.WriteReg(SlaveWriteRegister, []byte{data}); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (a *AS7262) virtualRegisterRead(register byte) (byte, error) {
	// if a.debug {
	// log.Printf("virtualRegisterRead(%02x)\n", register)
	// }
	const (
		SlaveStatusRegister byte = 0x00
		SlaveWriteRegister  byte = 0x01
		SlaveReadRegister   byte = 0x02
	)
	for {
		// if a.debug {
		// log.Printf("Checking Status Register \n")
		// }
		rx := make([]byte, 1)
		if err := a.dev.ReadReg(SlaveStatusRegister, rx); err != nil {
			log.Fatalln(err)
		}
		// if a.debug {
		// log.Printf("Status Register Contents: %02x\n", rx[0])
		// }
		if rx[0]&0x03 == 0x00 {
			break
		}
		// if there is data pending read it but thats all
		if rx[0]&0x03 == 0x01 {
			discard := make([]byte, 1)
			if err := a.dev.ReadReg(SlaveStatusRegister, discard); err != nil {
				log.Fatalln(err)
			}
			// if a.debug {
			// log.Printf("DataLeftInReadBuffer: %02x\n", discard[0])
			// }
		}
		time.Sleep(time.Millisecond)
	}
	// if a.debug {
	// log.Printf("Ready to Write to Status Register\n")
	// }
	// if a.debug {
	// log.Printf("WriteReg(%02x, %02x)\n", SlaveWriteRegister, register)
	// }
	if err := a.dev.WriteReg(SlaveWriteRegister, []byte{register}); err != nil {
		log.Fatalln(err)
	}

	for {
		// if a.debug {
		// log.Printf("Checking Status Register \n")
		// }
		rx := make([]byte, 1)
		if err := a.dev.ReadReg(SlaveStatusRegister, rx); err != nil {
			log.Fatalln(err)
		}
		// if a.debug {
		// log.Printf("Status Register Contents: %02x\n", rx[0])
		// }
		if rx[0]&0x03 == 0x01 {
			break
		}

		time.Sleep(time.Millisecond)
	}
	// if a.debug {
	// log.Printf("Data waiting Read Register\n")
	// }
	data := make([]byte, 1)
	if err := a.dev.ReadReg(SlaveReadRegister, data); err != nil {
		log.Fatalln(err)
	}
	// if a.debug {
	// log.Printf("Data in Read Register: %02x\n", data[0])
	// }
	return data[0], nil
}

// func NewSensor(mux i2cmux.Multiplexer, port uint8, opts ...func(*AS7262) error) (*AS7262, error) {
// 	a := new(AS7262)

// 	for _, option := range opts {
// 		option(a)
// 	}
// 	var err error
// 	a.dev, err = i2cmux.Open(0x49, mux, port)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	a.setConfig()
// 	return a, nil
// }

func (a *AS7262) Close() error {
	a.LEDoff()
	return a.dev.Close()
}

// const (
// 	I2C_AS72XX_SLAVE_STATUS_REG = 0x00
// 	I2C_AS72XX_SLAVE_WRITE_REG  = 0x01
// 	I2C_AS72XX_SLAVE_READ_REG   = 0x02
// 	I2C_AS72XX_SLAVE_TX_VALID   = 0x02
// 	I2C_AS72XX_SLAVE_RX_VALID   = 0x01
// )

// func (a *AS7262) writeReg(reg byte, buf []byte) error {
// 	log.Println("write reg")
// 	if err := a.checkTxPending(); err != nil {
// 		return err
// 	}

// 	// rx := make([]byte, 1)
// 	if err := a.dev.WriteReg(I2C_AS72XX_SLAVE_WRITE_REG, []byte{reg | 0x80}); err != nil {
// 		return err
// 	}

// 	if err := a.checkTxPending(); err != nil {
// 		return err
// 	}

// 	err := a.dev.WriteReg(I2C_AS72XX_SLAVE_WRITE_REG, buf)
// 	return err

// }

// func (a *AS7262) readReg(reg byte) (byte, error) {
// 	log.Println("read reg")
// 	if err := a.checkTxPending(); err != nil {
// 		return 0, err
// 	}

// 	// rx := make([]byte, 1)
// 	if err := a.dev.WriteReg(I2C_AS72XX_SLAVE_WRITE_REG, []byte{reg | 0x80}); err != nil {
// 		return 0, err
// 	}

// 	if err := a.checkRxPending(); err != nil {
// 		return 0, err
// 	}
// 	buf := make([]byte, 1)
// 	err := a.dev.ReadReg(I2C_AS72XX_SLAVE_READ_REG, buf)
// 	return buf[0], err

// }

// func (a *AS7262) checkRxPending() error {
// 	for {
// 		log.Println("checking rx pending")
// 		rx := make([]byte, 1)
// 		if err := a.dev.ReadReg(I2C_AS72XX_SLAVE_STATUS_REG, rx); err != nil {
// 			return err
// 		}
// 		log.Printf("%02x rx\n", rx)
// 		if rx[0]&I2C_AS72XX_SLAVE_RX_VALID == 0 {
// 			log.Println("checking rx ok")
// 			return nil
// 		}
// 		time.Sleep(time.Millisecond *10 * 10)
// 	}
// }

// func (a *AS7262) checkTxPending() error {
// 	for {
// 		log.Println("checking tx pending")
// 		rx := make([]byte, 1)
// 		if err := a.dev.ReadReg(I2C_AS72XX_SLAVE_STATUS_REG, rx); err != nil {
// 			log.Printf("rx error %02x %v\n", rx, err)
// 			// return err
// 		}
// 		if rx[0]&I2C_AS72XX_SLAVE_TX_VALID == 0 {
// 			log.Printf("checking tx ok, %02x\n", rx)
// 			return nil
// 		}
// 		time.Sleep(time.Millisecond *10 * 10)
// 	}
// }

func (a *AS7262) setConfig() error {
	if a.debug {
		fmt.Println("setConfig")
	}
	// if err := a.virtualRegisterWrite(0x04, 0xE0); err != nil {
	// 	return err
	// }
	// time.Sleep(time.Millisecond0)
	if err := a.virtualRegisterWrite(0x04, 0x3C); err != nil {
		return err
	}
	// if err := a.virtualRegisterWrite(0x06, 0xFF); err != nil {
	// 	return err
	// }
	// LED OFF
	err := a.LEDoff()
	return err
}

func (a *AS7262) LEDoff() error {
	// fmt.Println("ledoff")
	err := a.virtualRegisterWrite(0x07, 0x00)
	return err
}

func (a *AS7262) LEDon() error {
	// fmt.Println("ledon")
	err := a.virtualRegisterWrite(0x07, 0x09)
	return err
}

// func (a *AS7262) clearData() error {
// 	fmt.Println("clearData")
// 	value, err := a.virtualRegisterRead(0x04)
// 	if err != nil {
// 		return err
// 	}
// 	value = setBit(value, 1)
// 	if err := a.writeReg(0x04, []byte{value}); err != nil {
// 		return err
// 	}

// 	return nil
// }

func (a *AS7262) setMode(mode uint8) error {
	fmt.Println("setmode")

	if mode > 3 {
		mode = 3
	}

	control, err := a.virtualRegisterRead(0x04)
	if err != nil {
		return err
	}
	control &= 0xf3
	control |= (mode << 2)
	if err := a.virtualRegisterWrite(0x04, control); err != nil {
		return err
	}
	return nil
}

func (a *AS7262) dataReady() (bool, error) {
	// fmt.Println("dataReady?")
	var control byte
	err := retry(10, time.Millisecond*10*50, func() (err error) {
		control, err = a.virtualRegisterRead(0x04)
		return
	})

	if err != nil {
		log.Println(err)
		return false, err
	}

	ready := hasBit(control, 1)

	return ready, err

}

func retry(attempts int, sleep time.Duration, fn func() error) (err error) {
	for i := 0; ; i++ {
		err = fn()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func (a *AS7262) ReadAll() (Spectrum, error) {
	fmt.Println("readall")
	if err := a.setConfig(); err != nil {
		log.Println(err)
	}

	if err := a.LEDon(); err != nil {
		log.Println(err)
	}

	if err := a.setMode(3); err != nil {
		log.Println(err)
	}
	ready, err := a.dataReady()
	if err != nil {
		log.Println(err)
	}
	for !ready {
		// time.Sleep(time.Millisecond *10 * 50)
		ready, err = a.dataReady()
		if err != nil {
			log.Println(err)
		}
	}

	vh, err := a.virtualRegisterRead(0x08)
	if err != nil {
		return Spectrum{}, err
	}
	vl, err := a.virtualRegisterRead(0x09)
	if err != nil {
		return Spectrum{}, err
	}
	bh, err := a.virtualRegisterRead(0x0a)
	if err != nil {
		return Spectrum{}, err
	}
	bl, err := a.virtualRegisterRead(0x0b)
	if err != nil {
		return Spectrum{}, err
	}
	gh, err := a.virtualRegisterRead(0x0c)
	if err != nil {
		return Spectrum{}, err
	}
	gl, err := a.virtualRegisterRead(0x0d)
	if err != nil {
		return Spectrum{}, err
	}
	yh, err := a.virtualRegisterRead(0x0e)
	if err != nil {
		return Spectrum{}, err
	}
	yl, err := a.virtualRegisterRead(0x0f)
	if err != nil {
		return Spectrum{}, err
	}
	oh, err := a.virtualRegisterRead(0x10)
	if err != nil {
		return Spectrum{}, err
	}
	ol, err := a.virtualRegisterRead(0x11)
	if err != nil {
		return Spectrum{}, err
	}
	rh, err := a.virtualRegisterRead(0x12)
	if err != nil {
		return Spectrum{}, err
	}
	rl, err := a.virtualRegisterRead(0x13)
	if err != nil {
		return Spectrum{}, err
	}

	v := binary.BigEndian.Uint16([]byte{vh, vl})
	b := binary.BigEndian.Uint16([]byte{bh, bl})
	g := binary.BigEndian.Uint16([]byte{gh, gl})
	y := binary.BigEndian.Uint16([]byte{yh, yl})
	o := binary.BigEndian.Uint16([]byte{oh, ol})
	r := binary.BigEndian.Uint16([]byte{rh, rl})

	// GET Calibrated Float32

	vcal0, err := a.virtualRegisterRead(0x14)
	if err != nil {
		return Spectrum{}, err
	}
	vcal1, err := a.virtualRegisterRead(0x15)
	if err != nil {
		return Spectrum{}, err
	}
	vcal2, err := a.virtualRegisterRead(0x16)
	if err != nil {
		return Spectrum{}, err
	}
	vcal3, err := a.virtualRegisterRead(0x17)
	if err != nil {
		return Spectrum{}, err
	}
	vcal32 := binary.BigEndian.Uint32([]byte{vcal0, vcal1, vcal2, vcal3})
	vcal := math.Float32frombits(vcal32)

	bcal0, err := a.virtualRegisterRead(0x18)
	if err != nil {
		return Spectrum{}, err
	}
	bcal1, err := a.virtualRegisterRead(0x19)
	if err != nil {
		return Spectrum{}, err
	}
	bcal2, err := a.virtualRegisterRead(0x1A)
	if err != nil {
		return Spectrum{}, err
	}
	bcal3, err := a.virtualRegisterRead(0x1B)
	if err != nil {
		return Spectrum{}, err
	}
	bcal32 := binary.BigEndian.Uint32([]byte{bcal0, bcal1, bcal2, bcal3})
	bcal := math.Float32frombits(bcal32)

	gcal0, err := a.virtualRegisterRead(0x1C)
	if err != nil {
		return Spectrum{}, err
	}
	gcal1, err := a.virtualRegisterRead(0x1D)
	if err != nil {
		return Spectrum{}, err
	}
	gcal2, err := a.virtualRegisterRead(0x1E)
	if err != nil {
		return Spectrum{}, err
	}
	gcal3, err := a.virtualRegisterRead(0x1F)
	if err != nil {
		return Spectrum{}, err
	}
	gcal32 := binary.BigEndian.Uint32([]byte{gcal0, gcal1, gcal2, gcal3})
	gcal := math.Float32frombits(gcal32)

	ycal0, err := a.virtualRegisterRead(0x20)
	if err != nil {
		return Spectrum{}, err
	}
	ycal1, err := a.virtualRegisterRead(0x21)
	if err != nil {
		return Spectrum{}, err
	}
	ycal2, err := a.virtualRegisterRead(0x22)
	if err != nil {
		return Spectrum{}, err
	}
	ycal3, err := a.virtualRegisterRead(0x23)
	if err != nil {
		return Spectrum{}, err
	}
	ycal32 := binary.BigEndian.Uint32([]byte{ycal0, ycal1, ycal2, ycal3})
	ycal := math.Float32frombits(ycal32)

	ocal0, err := a.virtualRegisterRead(0x24)
	if err != nil {
		return Spectrum{}, err
	}
	ocal1, err := a.virtualRegisterRead(0x25)
	if err != nil {
		return Spectrum{}, err
	}
	ocal2, err := a.virtualRegisterRead(0x26)
	if err != nil {
		return Spectrum{}, err
	}
	ocal3, err := a.virtualRegisterRead(0x27)
	if err != nil {
		return Spectrum{}, err
	}
	ocal32 := binary.BigEndian.Uint32([]byte{ocal0, ocal1, ocal2, ocal3})
	ocal := math.Float32frombits(ocal32)

	rcal0, err := a.virtualRegisterRead(0x28)
	if err != nil {
		return Spectrum{}, err
	}
	rcal1, err := a.virtualRegisterRead(0x29)
	if err != nil {
		return Spectrum{}, err
	}
	rcal2, err := a.virtualRegisterRead(0x2A)
	if err != nil {
		return Spectrum{}, err
	}
	rcal3, err := a.virtualRegisterRead(0x2B)
	if err != nil {
		return Spectrum{}, err
	}
	rcal32 := binary.BigEndian.Uint32([]byte{rcal0, rcal1, rcal2, rcal3})
	rcal := math.Float32frombits(rcal32)

	// Spectrum (450nm, 500nm, 550nm, 570nm, 600nm, 650nm).

	// f := Spectrum{[]Count{
	// 	{Wavelength: 450, Value: float64(vcal), Raw: v},
	// 	{Wavelength: 500, Value: float64(bcal), Raw: b},
	// 	{Wavelength: 550, Value: float64(gcal), Raw: g},
	// 	{Wavelength: 570, Value: float64(ycal), Raw: y},
	// 	{Wavelength: 600, Value: float64(ocal), Raw: o},
	// 	{Wavelength: 650, Value: float64(rcal), Raw: r},
	// }}, nil

	// spectrum := Spectrum{[]Count{}, {}}
	// return Spectrum{[]Count{}{Wavelength: 420, Value: 12.2}}
	// Wavelength: 450,
	// Value: vcal,
	// Raw: v}}{}

	// return Spectrum{v, b, g, y, o, r, vcal, bcal, gcal, ycal, ocal, rcal}, nil
	now := time.Now()
	return Spectrum{Time: now, Unixnano: now.UnixNano(), Counts: []Count{
		{Wavelength: 450, Value: float64(vcal), Raw: v},
		{Wavelength: 500, Value: float64(bcal), Raw: b},
		{Wavelength: 550, Value: float64(gcal), Raw: g},
		{Wavelength: 570, Value: float64(ycal), Raw: y},
		{Wavelength: 600, Value: float64(ocal), Raw: o},
		{Wavelength: 650, Value: float64(rcal), Raw: r},
	}}, nil
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
