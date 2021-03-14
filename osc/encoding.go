package osc

import (
	"bytes"
	"encoding/binary"
)

////
// De/Encoding functions
////

// readBlob reads an OSC blob from the blob byte array. Padding bytes are
// removed from the reader and not returned.
func readBlob(reader *bytes.Buffer) ([]byte, int, error) {
	// First, get the length
	var blobLen int32
	if err := binary.Read(reader, binary.BigEndian, &blobLen); err != nil {
		return nil, 0, err
	}
	n := 4 + int(blobLen)

	// Read the data
	blob := make([]byte, blobLen)
	if _, err := reader.Read(blob); err != nil {
		return nil, 0, err
	}

	// Remove the padding bytes
	numPadBytes := padBytesNeeded(int(blobLen))
	if numPadBytes > 0 {
		n += numPadBytes
		dummy := make([]byte, numPadBytes)
		if _, err := reader.Read(dummy); err != nil {
			return nil, 0, err
		}
	}

	return blob, n, nil
}

// writeBlob writes the data byte array as an OSC blob into buff. If the length
// of data isn't 32-bit aligned, padding bytes will be added.
func writeBlob(data []byte, buf *bytes.Buffer) (int, error) {
	// Add the size of the blob
	dlen := int32(len(data))
	if err := binary.Write(buf, binary.BigEndian, dlen); err != nil {
		return 0, err
	}

	// Write the data
	if _, err := buf.Write(data); err != nil {
		return 0, nil
	}

	// Add padding bytes if necessary
	numPadBytes := padBytesNeeded(len(data))
	if numPadBytes > 0 {
		padBytes := make([]byte, numPadBytes)
		n, err := buf.Write(padBytes)
		if err != nil {
			return 0, err
		}
		numPadBytes = n
	}

	return 4 + len(data) + numPadBytes, nil
}

// readPaddedString reads a padded string from the given reader. The padding
// bytes are removed from the reader.
func readPaddedString(reader *bytes.Buffer) (string, int, error) {
	// Read the string from the reader
	str, err := reader.ReadString(0)
	if err != nil {
		return "", 0, err
	}
	n := len(str)

	// Remove the string delimiter, in order to calculate the right amount
	// of padding bytes
	str = str[:len(str)-1]

	// Remove the padding bytes
	padLen := padBytesNeeded(len(str)) - 1
	if padLen > 0 {
		n += padLen
		padBytes := make([]byte, padLen)
		if _, err = reader.Read(padBytes); err != nil {
			return "", 0, err
		}
	}

	return str, n, nil
}

// writePaddedString writes a string with padding bytes to the a buffer.
// Returns, the number of written bytes and an error if any.
func writePaddedString(str string, buf *bytes.Buffer) (int, error) {
	// Write the string to the buffer
	n, err := buf.WriteString(str)
	if err != nil {
		return 0, err
	}

	// Calculate the padding bytes needed and create a buffer for the padding bytes
	numPadBytes := padBytesNeeded(len(str))
	if numPadBytes > 0 {
		padBytes := make([]byte, numPadBytes)
		// Add the padding bytes to the buffer
		n, err := buf.Write(padBytes)
		if err != nil {
			return 0, err
		}
		numPadBytes = n
	}

	return n + numPadBytes, nil
}

// padBytesNeeded determines how many bytes are needed to fill up to the next 4
// byte length.
func padBytesNeeded(elementLen int) int {
	return 4*(elementLen/4+1) - elementLen
}
