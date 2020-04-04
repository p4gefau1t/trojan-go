// Buffer-like byte slice
// Copyright (c) 2017 Fadhli Dzil Ikram
//
// Test file for buffer

package buffer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBufferAllocation(t *testing.T) {
	Convey("Given new unallocated buffer", t, func() {
		var buf Buffer

		Convey("When appended with data", func() {
			data := []byte("Hello")
			buf.Append(data)

			Convey("It should have same content as the original data", func() {
				So(buf.Bytes(), ShouldResemble, data)
			})
		})

		Convey("When appended with single byte", func() {
			data := byte('H')
			buf.AppendByte(data)

			Convey("It should have 1 byte length", func() {
				So(len(buf), ShouldEqual, 1)
			})

			Convey("It should have same content", func() {
				So(buf.Bytes()[0], ShouldEqual, data)
			})
		})

		Convey("When appended with integer", func() {
			data := 12345
			repr := []byte("012345")
			buf.AppendInt(data, len(repr))

			Convey("Should have same content with the integer representation", func() {
				So(buf.Bytes(), ShouldResemble, repr)
			})
		})
	})
}

func TestBufferReset(t *testing.T) {
	Convey("Given allocated buffer", t, func() {
		var buf Buffer
		data := []byte("Hello")
		replace := []byte("World")
		buf.Append(data)

		Convey("When buffer reset", func() {
			buf.Reset()

			Convey("It should have zero length", func() {
				So(len(buf), ShouldEqual, 0)
			})
		})

		Convey("When buffer reset and replaced with another append", func() {
			buf.Reset()
			buf.Append(replace)

			Convey("It should have same content with the replaced data", func() {
				So(buf.Bytes(), ShouldResemble, replace)
			})
		})
	})
}
