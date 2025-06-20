package main

import (
	"image"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/rs/zerolog/log"

	roverlib "github.com/VU-ASE/roverlib-go/src"
)

var terminated = false

const EMPTY_BATTERY_VOLTAGE = 14.9
const FULL_BATTERY_VOLTAGE = 16.8
const PLUGGED_IN_BATTERY_VOLTAGE = 16.9

// drawStringCentered draws a string in the center of the image
func drawStringCentered(drawer *font.Drawer, img *image1bit.VerticalLSB, str string) {
	// Calculate the width of the text
	textWidth := drawer.MeasureString(str).Round()

	// Calculate the width of the image
	imageWidth := img.Bounds().Dx()

	// Calculate the starting X coordinate to center the text
	startX := (imageWidth - textWidth) / 2

	// Set the X coordinate of the Dot field of the Drawer
	drawer.Dot.X = fixed.I(startX)

	drawer.DrawString(str)
}

func drawString(drawer *font.Drawer, str string) {
	drawer.DrawString(str)
}

func run(service roverlib.Service, config *roverlib.ServiceConfiguration) error {
	log.Info().Msgf("Starting display service v%s", *service.Version)
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open a handle to the first available I²C bus:
	// Display is on i2c-5 for the debix-board
	bus, err := i2creg.Open("/dev/i2c-5")
	if err != nil {
		return err
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	dev, err := ssd1306.NewI2C(bus, &ssd1306.DefaultOpts)
	if err != nil {
		return err
	}

	// Read the rover identity from the /etc/rover file (roverd created), if the file exists
	hostname := "Anonymous Rover"
	etcRover, err := os.ReadFile("/etc/roverd/info.txt")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read /etc/rover")
	} else {
		lines := strings.Split(string(etcRover), "\n")
		if len(lines) < 2 {
			log.Warn().Msg("Invalid /etc/rover file")
		} else {
			// The first line of the file contains the rover index
			index := string(lines[0])
			// Second line contains the rover name
			name := string(lines[1])

			hostname = index + " (" + name + ")"
		}
	}

	//
	// Roverlib will terminate our service when it cannot find the desired read stream
	// so we always initialize the display first, before we try accessing the read stream
	//

	batVoltStr := "Vlt: UNAVAILABLE"
	batPercentStr := "Bat: UNAVAILABLE"
	batVoltUpdate := time.Now() // we don't want to give false information if the battery has not been updated in a while
	fetchBattery := false

	// Keep reading the battery in the background
	go func() {
		for {
			log.Info().Msg("Fetching battery voltage")
			if !fetchBattery {
				log.Info().Msg("Fetching was disabled")
				time.Sleep(1 * time.Second)
				continue
			}

			battery := service.GetReadStream("battery", "voltage")
			if battery == nil {
				log.Warn().Msg("Battery read stream not found")
				batVoltStr = "Vlt: UNAVAILABLE"
				batPercentStr = "Bat: UNAVAILABLE"
			} else {
				bat, err := battery.Read()
				if err != nil {
					log.Error().Err(err).Msg("Error reading battery voltage")
					batVoltStr = "Vlt: ERROR"
					batPercentStr = "Bat: ERROR"
				} else if bat.GetBatteryOutput() == nil {
					log.Warn().Msg("Battery output not found")
					batVoltStr = "Vlt: UNDEFINED"
					batPercentStr = "Bat: UNDEFINED"
				} else {
					voltage := bat.GetBatteryOutput().CurrentOutputVoltage
					log.Info().Float64("voltage", float64(voltage)).Msg("Battery voltage")
					batVoltStr = "Vlt: " + strconv.FormatFloat(float64(voltage), 'f', 2, 64) + "V"
					batPercentStr = "Bat: " + voltageToPercent(voltage)
					batVoltUpdate = time.Now()
				}
			}
			// Do not waste CPU cycles, and let the user see the display
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		//
		// System utilization stats
		//
		cpuPercent, _ := cpu.Percent(0, true)
		cpuStr := "CPU: "
		for _, cpu := range cpuPercent {
			cpuStr += strconv.FormatFloat(cpu, 'f', 0, 64) + ";"
		}
		cpuStr = cpuStr[:len(cpuStr)-1]
		memInfo, _ := mem.VirtualMemory()
		memStr := "Mem: " + strconv.FormatFloat(memInfo.UsedPercent, 'f', 2, 64) + "%"
		img := image1bit.NewVerticalLSB(dev.Bounds())
		f := basicfont.Face7x13
		drawer := font.Drawer{
			Dst:  img,
			Src:  &image.Uniform{image1bit.On},
			Face: f,
			Dot:  fixed.P(0, img.Bounds().Dy()-1-f.Descent),
		}

		//
		// Drawing
		//

		// Initialize starting Y coordinate
		y := int(img.Bounds().Dy() - 1)
		// Draw battery voltage
		drawer.Dot = fixed.P(0, y)
		if terminated {
			drawString(&drawer, "Unplug me!")
		} else {
			// Draw Battery Percentage
			drawString(&drawer, batPercentStr)
			// Decrease Y coordinate by text height
			y -= basicfont.Face7x13.Metrics().Height.Ceil()
			// Draw Battery voltage
			drawer.Dot = fixed.P(0, y)
			drawString(&drawer, batVoltStr)
			// Decrease Y coordinate by text height
			y -= basicfont.Face7x13.Metrics().Height.Ceil()
			// Draw CPU usage
			drawer.Dot = fixed.P(0, y)
			drawString(&drawer, cpuStr)
			// Decrease Y coordinate by text height
			y -= basicfont.Face7x13.Metrics().Height.Ceil()
			// Draw memory usage
			drawer.Dot = fixed.P(0, y)
			drawString(&drawer, memStr)
			y -= basicfont.Face7x13.Metrics().Height.Ceil()
			// Draw hostname
			drawer.Dot = fixed.P(0, y)
			// hostname = "=" + hostname + "="
			drawStringCentered(&drawer, img, hostname)
			if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
				log.Error().Err(err).Msg("Error drawing image")
			}
		}

		fetchBattery = true

		// If more than 30 seconds have passed since the last battery update, show it
		if time.Since(batVoltUpdate) > 30*time.Second {
			batVoltStr = "Vlt: TIMEOUT"
			batPercentStr = "Bat: TIMEOUT"
		}

		// Do not waste CPU cycles, and let the user see the display
		time.Sleep(5 * time.Second)
	}
}

func voltageToPercent(voltage float32) string {
	if(voltage < EMPTY_BATTERY_VOLTAGE){
		return "turning off..."
	} 
	if(voltage > PLUGGED_IN_BATTERY_VOLTAGE){
		return "plugged in"
	}
	if(voltage > FULL_BATTERY_VOLTAGE){
		return "100%"
	}
	// percentage must be between 0-100 here

	percentage := int((voltage - EMPTY_BATTERY_VOLTAGE) / (FULL_BATTERY_VOLTAGE - EMPTY_BATTERY_VOLTAGE) * 100)
	return strconv.Itoa(percentage) + "%"
}

func onTerminate(sig os.Signal) error {
	log.Info().Msg("Terminating display service")
	terminated = true
	time.Sleep(1 * time.Second)
	return nil
}

func main() {
	roverlib.Run(run, onTerminate)
}
