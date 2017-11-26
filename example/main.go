package main

import (
	"fmt"
	"log"
	"time"

	"github.com/NeuralSpaz/as7262"
)

func main() {
	fmt.Println("starting sensor")
	// mux, err := pca9548a.NewMux("/dev/i2c-1")
	// defer mux.Close()
	// if err != nil {
	// 	log.Panic(err)
	// }
	// sensor, err := as7262.NewSensor(mux, 0)
	// if err != nil {
	// 	log.Panic(err)
	// }
	sensor, err := as7262.NewSensor("/dev/i2c-1")
	if err != nil {
		log.Fatalln(err)
	}
	defer sensor.Close()
	log.Println(sensor)

	// sensor2, err := as7262.NewSensor(mux, 1)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// defer sensor2.Close()
	// log.Println(sensor2)

	for {
		<-time.After(time.Second)
		data, err := sensor.ReadAll()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(data)
		sensor.LEDoff()
	}
}
