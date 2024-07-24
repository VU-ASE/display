package main

import (
	"image"
	"os"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"google.golang.org/protobuf/proto"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	zmq "github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"

	pb_module_outputs "github.com/VU-ASE/pkg-CommunicationDefinitions/v2/packages/go/outputs"
)

const batterysrvr = "tcp://localhost:6000"

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

func main() {
	if _, err := host.Init(); err != nil {
		log.Err(err)
	}

	// Open a handle to the first available I²C bus:
	// Display is on i2c-5 for the debix-board
	bus, err := i2creg.Open("/dev/i2c-5")
	if err != nil {
		log.Err(err)
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	dev, err := ssd1306.NewI2C(bus, &ssd1306.DefaultOpts)
	if err != nil {
		log.Err(err)
	}

	// Start zmq subscriber
	subscriber, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		log.Err(err)
	}
	defer subscriber.Close()

	err = subscriber.Connect(batterysrvr)
	if err != nil {
		log.Err(err)
	}

	err = subscriber.SetSubscribe("") // Subscribe to all messages
	if err != nil {
		log.Err(err)
	}

	for {
		// Main receiver loop
		msg, err := subscriber.RecvBytes(0)
		// Don't exit on errors but log them
		if err != nil {
			log.Error().Err(err).Msg("Error receiving bytes")
			continue
		}

		// Decode the message -- first is the sensor wrapper message
		wrapperData := pb_module_outputs.SensorOutput{}
		err = proto.Unmarshal(msg, &wrapperData)
		if err != nil {
			log.Error().Err(err).Msg("Error unmarshalling message")
			continue
		}

		// Get the distance output from the message
		batData := wrapperData.GetBatteryOutput()
		batVolt := float64(batData.CurrentOutputVoltage)

		// Convert battery voltage to a string and write it to the device
		batVoltStr := "Bat: " + strconv.FormatFloat(batVolt, 'f', 2, 64) + "V"

		// Get system utilization
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

		// Initialize starting Y coordinate
		y := int(img.Bounds().Dy() - 1 - basicfont.Face7x13.Metrics().Height.Ceil())

		// Draw battery voltage
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
		hostname, _ := os.Hostname()

		// Add "=" on either side of the hostname
		hostname = "=" + hostname + "="
		drawStringCentered(&drawer, img, hostname)

		if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
			log.Error().Err(err).Msg("Error drawing image")
		}
	}
}
