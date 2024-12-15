package go_Weather_ITUR

import "github.com/joshuaferrara/go-satellite"

type TLEComponent struct {
	line1     string
	line2     string
	gravConst satellite.Gravity
}
