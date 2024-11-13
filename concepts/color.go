// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"fmt"
	"image/color"
	"math"
)

const Pr = .299
const Pg = .587
const Pb = .114

// public domain function by Darel Rex Finley, 2006
//
// This function expects the passed-in values to be on a scale
// of 0 to 1, and uses that same scale for the return values.
//
// See description/examples at alienryderflex.com/hsp.html
func RGBtoHSP(rgb *Vector3) Vector3 {
	var hsp Vector3

	//  Calculate the Perceived brightness.
	hsp[2] = math.Sqrt(rgb[0]*rgb[0]*Pr + rgb[1]*rgb[1]*Pg + rgb[2]*rgb[2]*Pb)

	//  Calculate the Hue and Saturation.  (This part works
	//  the same way as in the HSV/B and HSL systems???.)
	if rgb[0] == rgb[1] && rgb[0] == rgb[2] {
		hsp[0] = 0
		hsp[1] = 0
		return hsp
	}
	switch {
	case rgb[0] >= rgb[1] && rgb[0] >= rgb[2]: //  R is largest
		if rgb[2] >= rgb[1] {
			hsp[0] = 1.0 - (rgb[2]-rgb[1])/(6.0*(rgb[0]-rgb[1]))
			hsp[1] = 1.0 - rgb[1]/rgb[0]
		} else {
			hsp[0] = (rgb[1] - rgb[2]) / (6.0 * (rgb[0] - rgb[2]))
			hsp[1] = 1.0 - rgb[2]/rgb[0]
		}
	case rgb[1] >= rgb[0] && rgb[1] >= rgb[2]: //  G is largest
		if rgb[0] >= rgb[2] {
			hsp[0] = 2.0/6.0 - (rgb[0]-rgb[2])/(6.0*(rgb[1]-rgb[2]))
			hsp[1] = 1.0 - rgb[2]/rgb[1]
		} else {
			hsp[0] = 2.0/6.0 + (rgb[2]-rgb[0])/(6.0*(rgb[1]-rgb[0]))
			hsp[1] = 1.0 - rgb[0]/rgb[1]
		}
	default: //  B is largest
		if rgb[1] >= rgb[0] {
			hsp[0] = 4.0/6.0 - (rgb[1]-rgb[0])/(6.0*(rgb[2]-rgb[0]))
			hsp[1] = 1.0 - rgb[0]/rgb[2]
		} else {
			hsp[0] = 4.0/6.0 + (rgb[0]-rgb[1])/(6.0*(rgb[2]-rgb[1]))
			hsp[1] = 1.0 - rgb[1]/rgb[2]
		}
	}
	return hsp
}

// public domain function by Darel Rex Finley, 2006
//
// This function expects the passed-in values to be on a scale
// of 0 to 1, and uses that same scale for the return values.
//
// Note that some combinations of HSP, even if in the scale
// 0-1, may return RGB values that exceed a value of 1.  For
// example, if you pass in the HSP color 0,1,1, the result
// will be the RGB color 2.037,0,0.
//
// See description/examples at alienryderflex.com/hsp.html
func HSPtoRGB(hsp *Vector3) Vector3 {
	var rgb Vector3
	var part float64
	minOverMax := 1.0 - hsp[1]

	if minOverMax > 0 {
		if hsp[0] < 1./6. { //  R>G>B
			hsp[0] = 6. * (hsp[0] - 0./6.)
			part = 1. + hsp[0]*(1./minOverMax-1.)

			rgb[2] = hsp[2] / math.Sqrt(Pr/minOverMax/minOverMax+Pg*part*part+Pb)
			rgb[0] = (rgb[2]) / minOverMax
			rgb[1] = (rgb[2]) + hsp[0]*((rgb[0])-(rgb[2]))
		} else if hsp[0] < 2./6. { //  G>R>B
			hsp[0] = 6. * (-hsp[0] + 2./6.)
			part = 1. + hsp[0]*(1./minOverMax-1.)

			rgb[2] = hsp[2] / math.Sqrt(Pg/minOverMax/minOverMax+Pr*part*part+Pb)
			rgb[1] = (rgb[2]) / minOverMax
			rgb[0] = (rgb[2]) + hsp[0]*((rgb[1])-(rgb[2]))
		} else if hsp[0] < 3./6. { //  G>B>R
			hsp[0] = 6. * (hsp[0] - 2./6.)
			part = 1. + hsp[0]*(1./minOverMax-1.)

			rgb[0] = hsp[2] / math.Sqrt(Pg/minOverMax/minOverMax+Pb*part*part+Pr)
			rgb[1] = (rgb[0]) / minOverMax
			rgb[2] = (rgb[0]) + hsp[0]*((rgb[1])-(rgb[0]))
		} else if hsp[0] < 4./6. { //  B>G>R
			hsp[0] = 6. * (-hsp[0] + 4./6.)
			part = 1. + hsp[0]*(1./minOverMax-1.)

			rgb[0] = hsp[2] / math.Sqrt(Pb/minOverMax/minOverMax+Pg*part*part+Pr)
			rgb[2] = (rgb[0]) / minOverMax
			rgb[1] = (rgb[0]) + hsp[0]*((rgb[2])-(rgb[0]))
		} else if hsp[0] < 5./6. { //  B>R>G
			hsp[0] = 6. * (hsp[0] - 4./6.)
			part = 1. + hsp[0]*(1./minOverMax-1.)

			rgb[1] = hsp[2] / math.Sqrt(Pb/minOverMax/minOverMax+Pr*part*part+Pg)
			rgb[2] = (rgb[1]) / minOverMax
			rgb[0] = (rgb[1]) + hsp[0]*((rgb[2])-(rgb[1]))
		} else { //  R>B>G
			hsp[0] = 6. * (-hsp[0] + 1)
			part = 1. + hsp[0]*(1./minOverMax-1.)

			rgb[1] = hsp[2] / math.Sqrt(Pr/minOverMax/minOverMax+Pb*part*part+Pg)

			rgb[0] = (rgb[1]) / minOverMax
			rgb[2] = (rgb[1]) + hsp[0]*((rgb[0])-(rgb[1]))
		}
	} else {
		if hsp[0] < 1./6. { //  R>G>B
			hsp[0] = 6. * (hsp[0] - 0./6.)
			rgb[0] = math.Sqrt(hsp[2] * hsp[2] / (Pr + Pg*hsp[0]*hsp[0]))
			rgb[1] = (rgb[0]) * hsp[0]
			rgb[2] = 0.
		} else if hsp[0] < 2./6. { //  G>R>B
			hsp[0] = 6. * (-hsp[0] + 2./6.)
			rgb[1] = math.Sqrt(hsp[2] * hsp[2] / (Pg + Pr*hsp[0]*hsp[0]))
			rgb[0] = (rgb[1]) * hsp[0]
			rgb[2] = 0.
		} else if hsp[0] < 3./6. { //  G>B>R
			hsp[0] = 6. * (hsp[0] - 2./6.)
			rgb[1] = math.Sqrt(hsp[2] * hsp[2] / (Pg + Pb*hsp[0]*hsp[0]))
			rgb[2] = (rgb[1]) * hsp[0]
			rgb[0] = 0.
		} else if hsp[0] < 4./6. { //  B>G>R
			hsp[0] = 6. * (-hsp[0] + 4./6.)
			rgb[2] = math.Sqrt(hsp[2] * hsp[2] / (Pb + Pg*hsp[0]*hsp[0]))
			rgb[1] = (rgb[2]) * hsp[0]
			rgb[0] = 0.
		} else if hsp[0] < 5./6. { //  B>R>G
			hsp[0] = 6. * (hsp[0] - 4./6.)
			rgb[2] = math.Sqrt(hsp[2] * hsp[2] / (Pb + Pr*hsp[0]*hsp[0]))
			rgb[0] = (rgb[2]) * hsp[0]
			rgb[1] = 0.
		} else { //  R>B>G
			hsp[0] = 6. * (-hsp[0] + 1.0)
			rgb[0] = math.Sqrt(hsp[2] * hsp[2] / (Pr + Pb*hsp[0]*hsp[0]))
			rgb[2] = (rgb[0]) * hsp[0]
			rgb[1] = 0.
		}
	}
	return rgb
}

func ParseHexColor(hex string) (color.NRGBA, error) {
	var r, g, b, factor uint8
	var n int
	var err error
	if len(hex) == 4 {
		n, err = fmt.Sscanf(hex, "#%1x%1x%1x", &r, &g, &b)
		factor = 16
	} else {
		n, err = fmt.Sscanf(hex, "#%2x%2x%2x", &r, &g, &b)
		factor = 1
	}
	if err != nil {
		return color.NRGBA{}, err
	}
	if n != 3 {
		return color.NRGBA{}, fmt.Errorf("color %v is not a hex-color", hex)
	}
	return color.NRGBA{r * factor, g * factor, b * factor, 255}, nil
}

func ColorToInt32PreMul(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
	r = r >> 8
	g = g >> 8
	b = b >> 8
	a = a >> 8

	return ((r & 0xFF) << 24) | ((g & 0xFF) << 16) | ((b & 0xFF) << 8) | (a & 0xFF)
}

func NRGBAToInt32(c color.NRGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}
func RGBAToInt32(c color.RGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}

func Int32ToNRGBA(c uint32) color.NRGBA {
	return color.NRGBA{uint8((c >> 24) & 0xFF), uint8((c >> 16) & 0xFF), uint8((c >> 8) & 0xFF), uint8(c & 0xFF)}
}

func Int32ToRGBA(c uint32) color.RGBA {
	return color.RGBA{uint8((c >> 24) & 0xFF), uint8((c >> 16) & 0xFF), uint8((c >> 8) & 0xFF), uint8(c & 0xFF)}
}
