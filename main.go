package main

import (
	"fmt"
	"sync"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/sphero"
)

func identicalColors(colors1, colors2 []uint8) bool {
	return colors1[0] == colors2[0] && colors1[1] == colors2[1] && colors1[2] == colors2[2]
}

func changeColor(spheroDriver *sphero.SpheroDriver) <-chan struct{} {
	done := make(chan struct{})
	colors := spheroDriver.GetRGB()
	destColors := []uint8{uint8(gobot.Rand(255)), uint8(gobot.Rand(255)), uint8(gobot.Rand(255))}

	go func() {
		defer close(done)

		for !identicalColors(colors, destColors) {
			var i int
			for i = range colors {
				if destColors[i] > colors[i] {
					colors[i]++
				} else {
					colors[i]--
				}
			}
			spheroDriver.SetRGB(colors[0], colors[1], colors[2])
			time.Sleep(16 * time.Millisecond)
		}
	}()

	return done
}

func main() {
	adaptor := sphero.NewAdaptor("/dev/rfcomm0")
	spheroDriver := sphero.NewSpheroDriver(adaptor)

	work := func() {
		spheroDriver.SetDataStreaming(sphero.DefaultDataStreamingConfig())

		spheroDriver.On(sphero.Collision, func(data interface{}) {
			fmt.Printf("Collision! %+v\n", data)
			wg := sync.WaitGroup{}

			/* change colors 5 times */
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					<-changeColor(spheroDriver)
				}
			}()
			/* meanwhile, roll until colors are done */
			stopShakeHead := make(chan struct{})
			checkHeadWg := sync.WaitGroup{}
			checkHeadWg.Add(1)
			go func() {
				defer checkHeadWg.Done()
				left := true

				for {
					select {
					case <-stopShakeHead:
						break
					default:
						if !left {
							spheroDriver.Roll(0, uint16(gobot.Rand(180)))
							left = true
						} else {
							spheroDriver.Roll(0, uint16(gobot.Rand(180)+180))
							left = false
						}
						time.Sleep(500 * time.Millisecond)
					}
				}
			}()

			wg.Wait()
			close(stopShakeHead)
			checkHeadWg.Wait()
			spheroDriver.Roll(50, uint16(gobot.Rand(360)))
		})

		/*spheroDriver.On(sphero.SensorData, func(data interface{}) {
			fmt.Printf("Streaming Data! %+v\n", data)
		})*/
	}

	robot := gobot.NewRobot("sphero",
		[]gobot.Connection{adaptor},
		[]gobot.Device{spheroDriver},
		work,
	)

	robot.Start()
}
