package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/NeuralSpaz/as7262"
	"github.com/NeuralSpaz/as7263"
	"github.com/NeuralSpaz/pca9548a"
	"github.com/nats-io/nats"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	// stop := make(chan bool)
	// go func() {
	// 	for sig := range c {
	// 		// sig is a ^C, handle it
	// 	}
	// }()

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
		// <-time.After(time.Second)
		start := time.Now()
		waitTime := zero.Request()
		time.Sleep(time.Millisecond * 120)
		one.Request()
		time.Sleep(time.Millisecond * 120)
		six.Request()
		time.Sleep(time.Millisecond * 120)
		seven.Request()
		time.Sleep(time.Millisecond * 120)
		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Println("elapsed: ", elapsed)
		fmt.Println("idle: ", waitTime-elapsed)

		<-time.After(waitTime - elapsed)

		zeroreadstart := time.Now()
		zeroData, err := zero.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", zeroData)
		zero.LEDoff()
		zeroreadElapse := time.Now().Sub(zeroreadstart)
		fmt.Println("read duration: ", zeroreadElapse)

		onereadstart := time.Now()
		oneData, err := one.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", oneData)
		one.LEDoff()
		onereadElapse := time.Now().Sub(onereadstart)
		fmt.Println("one read duration: ", onereadElapse)

		sixreadstart := time.Now()
		sixData, err := six.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", sixData)
		six.LEDoff()
		sixreadElapse := time.Now().Sub(sixreadstart)
		fmt.Println("six read duration: ", sixreadElapse)

		sevenreadstart := time.Now()
		sevenData, err := seven.ReadAll()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+#v\n", sevenData)
		seven.LEDoff()
		sevenreadElapse := time.Now().Sub(sevenreadstart)
		fmt.Println("seven read duration: ", sevenreadElapse)

		go func() {
			catholyteVis <- zeroData
			catholyteNIR <- oneData
			anolyteNIR <- sixData
			anolyteVis <- sevenData
		}()

		select {
		case _, ok := <-sig:
			if ok {
				fmt.Printf("Asked to quit, now exiting\n")
				time.Sleep(time.Second * 1)
				zero.Close()
				one.Close()
				six.Close()
				seven.Close()
				mux.Close()
				os.Exit(1)
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			fmt.Println("No value ready, moving on.")
		}

	}
}
