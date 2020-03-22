// The color engine for the go-log library
// Copyright (c) 2017 Fadhli Dzil Ikram

package colorful

import "github.com/withmandala/go-log/buffer"

// ColorBuffer add color option to buffer append
type ColorBuffer struct {
	buffer.Buffer
}

// color pallete map
var (
	colorOff    = []byte("\033[0m")
	colorRed    = []byte("\033[0;31m")
	colorGreen  = []byte("\033[0;32m")
	colorOrange = []byte("\033[0;33m")
	colorBlue   = []byte("\033[0;34m")
	colorPurple = []byte("\033[0;35m")
	colorCyan   = []byte("\033[0;36m")
	colorGray   = []byte("\033[0;37m")
)

// Off apply no color to the data
func (cb *ColorBuffer) Off() {
	cb.Append(colorOff)
}

// Red apply red color to the data
func (cb *ColorBuffer) Red() {
	cb.Append(colorRed)
}

// Green apply green color to the data
func (cb *ColorBuffer) Green() {
	cb.Append(colorGreen)
}

// Orange apply orange color to the data
func (cb *ColorBuffer) Orange() {
	cb.Append(colorOrange)
}

// Blue apply blue color to the data
func (cb *ColorBuffer) Blue() {
	cb.Append(colorBlue)
}

// Purple apply purple color to the data
func (cb *ColorBuffer) Purple() {
	cb.Append(colorPurple)
}

// Cyan apply cyan color to the data
func (cb *ColorBuffer) Cyan() {
	cb.Append(colorCyan)
}

// Gray apply gray color to the data
func (cb *ColorBuffer) Gray() {
	cb.Append(colorGray)
}

// mixer mix the color on and off byte with the actual data
func mixer(data []byte, color []byte) []byte {
	var result []byte
	return append(append(append(result, color...), data...), colorOff...)
}

// Red apply red color to the data
func Red(data []byte) []byte {
	return mixer(data, colorRed)
}

// Green apply green color to the data
func Green(data []byte) []byte {
	return mixer(data, colorGreen)
}

// Orange apply orange color to the data
func Orange(data []byte) []byte {
	return mixer(data, colorOrange)
}

// Blue apply blue color to the data
func Blue(data []byte) []byte {
	return mixer(data, colorBlue)
}

// Purple apply purple color to the data
func Purple(data []byte) []byte {
	return mixer(data, colorPurple)
}

// Cyan apply cyan color to the data
func Cyan(data []byte) []byte {
	return mixer(data, colorCyan)
}

// Gray apply gray color to the data
func Gray(data []byte) []byte {
	return mixer(data, colorGray)
}
