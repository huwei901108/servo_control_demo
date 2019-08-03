package main

import (
	"fmt"
	"gocv.io/x/gocv"
)

func main() {
	fmt.Println("start")
	webcam, _ := gocv.VideoCaptureDevice(0)
	window := gocv.NewWindow("hello")
	img := gocv.NewMat()
	for {
		fmt.Println("reading")
		webcam.Read(&img)
		window.IMShow(img)
		window.WaitKey(1)

	}
}
