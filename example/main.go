package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NeuralSpaz/as7262"
	"github.com/NeuralSpaz/as7263"
	"github.com/NeuralSpaz/pca9548a"
	"github.com/nats-io/nats"
)

func main() {
	fmt.Println("starting sensor")
	mux, err := pca9548a.NewMux("/dev/i2c-1")
	defer mux.Close()
	if err != nil {
		log.Panic(err)
	}
	zero, err := as7262.NewSensor(mux, 0)
	if err != nil {
		log.Panic(err)
	}
	defer zero.Close()

	one, err := as7263.NewSensor(mux, 1)
	if err != nil {
		log.Panic(err)
	}
	defer one.Close()

	six, err := as7263.NewSensor(mux, 6)
	if err != nil {
		log.Panic(err)
	}
	defer six.Close()

	seven, err := as7262.NewSensor(mux, 7)
	if err != nil {
		log.Panic(err)
	}
	defer seven.Close()

	servers := "nats://127.0.0.1:4222"
	hostname, _ := os.Hostname()
	name := nats.Name(hostname)
	nc, err := nats.Connect(servers, name)
	if err != nil {
		log.Fatalln(err)
	}
	c, _ := nats.NewEncodedConn(nc, "json")
	defer c.Close()

	catholyteVis := make(chan as7262.Spectrum, 10)
	catholyteNIR := make(chan as7263.Spectrum, 10)

	anolyteVis := make(chan as7262.Spectrum, 10)
	anolyteNIR := make(chan as7263.Spectrum, 10)

	c.BindSendChan("catholyte.vis", catholyteVis)
	c.BindSendChan("catholyte.nir", catholyteNIR)
	c.BindSendChan("anolyte.vis", anolyteVis)
	c.BindSendChan("anolyte.nir", anolyteNIR)

	for {
		<-time.After(time.Second)
		zeroData, err := zero.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", zeroData)
		zero.LEDoff()
		catholyteVis <- zeroData

		oneData, err := one.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", oneData)
		one.LEDoff()
		catholyteNIR <- oneData

		sixData, err := six.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", sixData)
		six.LEDoff()
		anolyteNIR <- sixData

		sevenData, err := seven.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", sevenData)
		seven.LEDoff()
		anolyteVis <- sevenData

	}
}
