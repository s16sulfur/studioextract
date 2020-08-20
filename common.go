package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/sulfur/bbio"
)

func printError(err error) {
	msg := fmt.Sprint(err)
	fmt.Println("\033[97;101m ERROR \033[0m", "\033[31m", msg, "\033[0m")
}

func createPng(width int, height int, sex int) ([]byte, error) {
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	bkgColorMale := color.RGBA{0x0, 0x0, 0xff, 0xff}
	bkgColorFemale := color.RGBA{0xff, 0x80, 0xff, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if sex == 0 {
				img.Set(x, y, bkgColorMale)
			} else {
				img.Set(x, y, bkgColorFemale)
			}
		}
	}

	buf := new(bytes.Buffer)
	bufr := bufio.NewWriter(buf)
	err := png.Encode(bufr, img)
	if err != nil {
		return nil, err
	}

	ferr := bufr.Flush()
	if ferr != nil {
		return nil, err
	}

	pngBytes := buf.Bytes()
	return pngBytes, nil
}

func getPngSize(reader *bbio.Reader) int64 {
	pngEndChunk := []byte{0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82}
	pngEndIdx := reader.Index(pngEndChunk)
	return int64(pngEndIdx + len(pngEndChunk))
}

func checkPngData(r io.ReadSeeker) (size int64, err error) {
	pngStartChunk := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	size = 0
	// sr := io.NewSectionReader(r)
	reader := bufio.NewReader(r)

	startLen := len(pngStartChunk)
	bufStart := make([]byte, startLen)

	n, sErr := reader.Read(bufStart)
	if sErr != nil {
		err = sErr
		return
	}
	if n != startLen {
		err = errors.New("Too small for png")
		return
	}
	for i := 0; i < n; i++ {
		if pngStartChunk[i] != bufStart[i] {
			err = errors.New("Png start not found")
			return
		}
	}

	var rn, pos int
	var rErr error

	for {
		readBuf := make([]byte, 4)
		rn, rErr = reader.Read(bufStart)
		if rErr != nil {
			err = rErr
			return
		}
		if rn != 4 {
			err = io.EOF
			return
		}

		first := binary.BigEndian.Uint32(readBuf)
		rn, rErr = reader.Read(bufStart)
		if rErr != nil {
			err = rErr
			return
		}

		second := binary.LittleEndian.Uint32(readBuf)
		if second != 1145980233 {
			break
		}

		offset := int(first + 4)
		if offset > reader.Size() {
			return
		}

		_, snErr := r.Seek(int64(offset), io.SeekCurrent)
		if snErr != nil {
			return
		}

		pos += offset
	}

	size = int64(pos)
	return
}
