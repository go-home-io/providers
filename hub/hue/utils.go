package main

import (
	"math"

	"go-home.io/x/server/plugins/common"
)

// Converts RGB color into CIE.
func rgb2cie(color common.Color) (float32, float32) {
	r := rgb2cieMagic(color.R)
	g := rgb2cieMagic(color.G)
	b := rgb2cieMagic(color.B)

	x := r*0.649926 + g*0.103455 + b*0.197109
	y := r*0.234327 + g*0.743075 + b*0.022598
	z := r*0.0000000 + g*0.053077 + b*1.035763

	X := x / (x + y + z)
	Y := y / (x + y + z)

	return float32(X), float32(Y)
}

// Magic numbers around HUE implementation, while converting RGB into CIE.
func rgb2cieMagic(c uint8) float32 {
	correctedValue := float32(float32(c) / float32(brightnessMax))

	if correctedValue > 0.04045 {
		return float32(math.Pow((float64(correctedValue)+0.055)/(1.0+0.055), 2.4))
	}

	return correctedValue / 12.92
}

// Magic numbers around HUE implementation, while converting CIE into RGB.
func cie2rgbMagic(c float32) float32 {
	if c < 0.0031308 {
		return 12.92 * c
	}

	return float32((1.0+0.055)*math.Pow(float64(c), 1.0/2.4) - 0.055)
}

// Converts CIE color into RGB.
// nolint: gocyclo
func cie2rgb(x float32, y float32, brightness float32) common.Color {
	if 0 == brightness {
		brightness = brightnessMax
	}

	z := 1.0 - x - y
	Y := brightness / 254
	X := (Y / y) * x
	Z := (Y / y) * z

	r := X*1.4628067 - Y*0.1840623 - Z*0.2743606
	g := -X*0.5217933 + Y*1.4472381 + Z*0.0677227
	b := X*0.0349342 - Y*0.0968930 + Z*1.2884099

	if r > b && r > g && r > 1.0 {
		g = g / r
		b = b / r
		r = 1.0
	} else if g > b && g > r && g > 1.0 {
		r = r / g
		b = b / g
		g = 1.0
	} else if b > r && b > g && b > 1.0 {
		r = r / b
		g = g / b
		b = 1.0
	}

	r = cie2rgbMagic(r)
	g = cie2rgbMagic(g)
	b = cie2rgbMagic(b)

	r = float32(math.Round(float64(r) * 255))
	g = float32(math.Round(float64(g) * 255))
	b = float32(math.Round(float64(b) * 255))

	return common.Color{R: uint8(r * brightnessMax), G: uint8(g * brightnessMax), B: uint8(b * brightnessMax)}
}
