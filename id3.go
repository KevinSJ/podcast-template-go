package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Read the ID3 tag header
func readID3Header(file *os.File) (tagSize int, err error) {
	header := make([]byte, 10)
	_, err = file.Read(header)
	if err != nil {
		return 0, err
	}

	// Ensure the file has an ID3v2 tag
	if string(header[:3]) != "ID3" {
		return 0, fmt.Errorf("no ID3v2 tag found")
	}

	// Calculate the sync-safe size
	tagSize = int(header[6])<<21 | int(header[7])<<14 | int(header[8])<<7 | int(header[9])
	return tagSize, nil
}

// Read a text frame (e.g., TIT2, TPE1)
func readTextFrame(reader io.Reader) (string, error, int) {
	// Read the next 10 bytes for the frame header
	frameHeader := make([]byte, 10)
	_, err := io.ReadFull(reader, frameHeader)
	if err != nil {
		return "", err, -1
	}

	// Extract the frame ID and size
	frameID := string(frameHeader[:4])
	frameSize := int(binary.BigEndian.Uint32(frameHeader[4:8]))

	// Check for a valid text frame (e.g., TIT2, TPE1)
	if frameID != "TIT2" && frameID != "TPE1" {
		// Skip this frame and return
		_, err = io.CopyN(io.Discard, reader, int64(frameSize))
		if err != nil {
			return "", err, -1
		}
		return "", fmt.Errorf("frame %s skipped", frameID), -1
	}

	// Read the frame content
	frameContent := make([]byte, frameSize)
	_, err = io.ReadFull(reader, frameContent)
	if err != nil {
		return "", err, -1
	}

	// Check encoding and decode
	encoding := frameContent[0]
	text := frameContent[1:] // Skip encoding byte
	switch encoding {
	case 0x00: // ISO-8859-1
		return string(text), nil, frameSize
	case 0x03: // UTF-8
		return string(text), nil, frameSize
	default:
		return "", fmt.Errorf("unsupported encoding: %d", encoding), frameSize
	}
}

func ReadID3Tags(filePath string) (title, artist string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	// Read the ID3 header to determine tag size
	tagSize, err := readID3Header(file)
	if err != nil {
		return "", "", err
	}

	// Only parse frames within the tag size
	tagData := make([]byte, tagSize)
	_, err = file.Read(tagData)
	if err != nil {
		return "", "", err
	}

	// Create a buffer for the tag data
	buffer := bytes.NewReader(tagData)

    frameSize := 0

	// Read TIT2 (Title)
	title, _, frameSize = readTextFrame(buffer)

	// Reset the buffer to start reading from the beginning for TPE1
	buffer.Seek(0, frameSize)

	// Read TPE1 (Artist)
	artist, _, _ = readTextFrame(buffer)

	return title, artist, nil
}

