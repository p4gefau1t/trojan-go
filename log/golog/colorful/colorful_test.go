// The color engine for the go-log library
// Copyright (c) 2017 Fadhli Dzil Ikram
//
// Test file

package colorful

import (
	"testing"

	"github.com/p4gefau1t/trojan-go/log/golog/buffer"
	. "github.com/smartystreets/goconvey/convey"
)

func TestColorBuffer(t *testing.T) {
	Convey("Given empty color buffer and test data", t, func() {
		var cb ColorBuffer
		var result buffer.Buffer

		// Add color to the result buffer
		result.Append(colorRed)
		result.Append(colorGreen)
		result.Append(colorOrange)
		result.Append(colorBlue)
		result.Append(colorPurple)
		result.Append(colorCyan)
		result.Append(colorGray)
		result.Append(colorOff)

		Convey("When appended with color", func() {
			cb.Red()
			cb.Green()
			cb.Orange()
			cb.Blue()
			cb.Purple()
			cb.Cyan()
			cb.Gray()
			cb.Off()

			Convey("It should have same content with the test data", func() {
				So(result.Bytes(), ShouldResemble, cb.Bytes())
			})
		})
	})
}

func TestColorMixer(t *testing.T) {
	Convey("Given mixer test result data", t, func() {
		var (
			data         = []byte("Hello")
			resultRed    buffer.Buffer
			resultGreen  buffer.Buffer
			resultOrange buffer.Buffer
			resultBlue   buffer.Buffer
			resultPurple buffer.Buffer
			resultCyan   buffer.Buffer
			resultGray   buffer.Buffer
		)

		// Add result to buffer
		resultRed.Append(colorRed)
		resultRed.Append(data)
		resultRed.Append(colorOff)

		resultGreen.Append(colorGreen)
		resultGreen.Append(data)
		resultGreen.Append(colorOff)

		resultOrange.Append(colorOrange)
		resultOrange.Append(data)
		resultOrange.Append(colorOff)

		resultBlue.Append(colorBlue)
		resultBlue.Append(data)
		resultBlue.Append(colorOff)

		resultPurple.Append(colorPurple)
		resultPurple.Append(data)
		resultPurple.Append(colorOff)

		resultCyan.Append(colorCyan)
		resultCyan.Append(data)
		resultCyan.Append(colorOff)

		resultGray.Append(colorGray)
		resultGray.Append(data)
		resultGray.Append(colorOff)

		Convey("It should have same result when data appended with Red color", func() {
			So(Red(data), ShouldResemble, resultRed.Bytes())
		})

		Convey("It should have same result when data appended with Green color", func() {
			So(Green(data), ShouldResemble, resultGreen.Bytes())
		})

		Convey("It should have same result when data appended with Orange color", func() {
			So(Orange(data), ShouldResemble, resultOrange.Bytes())
		})

		Convey("It should have same result when data appended with Blue color", func() {
			So(Blue(data), ShouldResemble, resultBlue.Bytes())
		})

		Convey("It should have same result when data appended with Purple color", func() {
			So(Purple(data), ShouldResemble, resultPurple.Bytes())
		})

		Convey("It should have same result when data appended with Cyan color", func() {
			So(Cyan(data), ShouldResemble, resultCyan.Bytes())
		})

		Convey("It should have same result when data appended with Gray color", func() {
			So(Gray(data), ShouldResemble, resultGray.Bytes())
		})
	})
}
