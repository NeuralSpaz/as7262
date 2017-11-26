package main

import (
	"fmt"
	"log"

	"github.com/NeuralSpaz/as7262"
	"github.com/NeuralSpaz/as7263"
	"github.com/NeuralSpaz/pca9548a"
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

	// sensor, err := as7262.NewSensor("/dev/i2c-1")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Println(one)

	// sensor2, err := as7262.NewSensor(mux, 1)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// defer sensor2.Close()
	// log.Println(sensor2)

	for {
		// <-time.After(time.Second)
		zeroData, err := zero.ReadAll()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(zeroData)
		zero.LEDoff()

		oneData, err := one.ReadAll()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(oneData)
		one.LEDoff()

		sixData, err := six.ReadAll()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(sixData)
		six.LEDoff()

		sevenData, err := seven.ReadAll()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(sevenData)
		seven.LEDoff()
	}
}
