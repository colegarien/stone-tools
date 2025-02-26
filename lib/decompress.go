package lib

import (
	"io"
)

func Decompress(reader io.ReadSeeker, compressedDataSize uint32) ([]byte, error) {
	var (
		err                               error
		mostSigBit, combinedWord          uint32
		offsetToCopyFrom, byteCountToCopy uint32
		currentDataByte, byteToCopy       byte
		leastSigBit, indicatorWord        uint32
		bytesCopied, bufferIndexToCopy    int32
		nextOffsetFreeInCircularBuffer    uint32 = 1
		bitsLeftToShift                   uint8
		dataSizeLeft                      int32
	)

	outputBuffer := make([]byte, 0, 0x8000) // Start with 32KB capacity
	circularCopyBuffer := make([]byte, 0x400)

	for {
		for {
			if bitsLeftToShift == 0 {
				bitsLeftToShift = 8

				if compressedDataSize == 0 {
					indicatorWord = 0xffffffff
				} else {
					var buf [1]byte
					_, err = reader.Read(buf[:])
					if err != nil {
						return nil, err
					}

					indicatorWord = uint32(buf[0])
					compressedDataSize--
				}
			}

			bitsLeftToShift--
			dataSizeLeft = int32(compressedDataSize)

			preshiftIndicator := indicatorWord
			indicatorWord >>= 1
			if preshiftIndicator&1 == 0 {
				break
			}

			if dataSizeLeft == 0 {
				currentDataByte = 0xff
			} else {
				var buf [1]byte
				_, err = reader.Read(buf[:])
				if err != nil {
					return nil, err
				}

				currentDataByte = buf[0]
				compressedDataSize--
			}

			outputBuffer = append(outputBuffer, currentDataByte)
			circularCopyBuffer[nextOffsetFreeInCircularBuffer] = currentDataByte
			nextOffsetFreeInCircularBuffer = (nextOffsetFreeInCircularBuffer + 1) & 0x3ff
		}

		if dataSizeLeft == 0 {
			leastSigBit = 0xffffffff
		} else {
			var buf [1]byte
			_, err = reader.Read(buf[:])
			if err != nil {
				return nil, err
			}

			leastSigBit = uint32(buf[0])
			compressedDataSize = uint32(dataSizeLeft - 1)
		}

		if compressedDataSize == 0 {
			mostSigBit = 0xffffffff
		} else {
			var buf [1]byte
			_, err = reader.Read(buf[:])
			if err != nil {
				return nil, err
			}

			mostSigBit = uint32(buf[0])
			compressedDataSize--
		}

		// read two byte and combine them into a word
		combinedWord = mostSigBit*0x100 + leastSigBit
		// get the intended offset to copy from
		offsetToCopyFrom = combinedWord & 0x3ff
		if offsetToCopyFrom == 0 {
			break
		}

		bytesCopied = 0
		byteCountToCopy = ((combinedWord >> 10) & 0x3f) + 2
		bufferIndexToCopy = int32(nextOffsetFreeInCircularBuffer) - int32(offsetToCopyFrom)

		for bytesCopied <= int32(byteCountToCopy) {
			byteToCopy = circularCopyBuffer[((bufferIndexToCopy&0x3ff)+(bytesCopied&0x3ff))&0x3ff]
			outputBuffer = append(outputBuffer, byteToCopy)

			circularCopyBuffer[nextOffsetFreeInCircularBuffer] = byteToCopy
			nextOffsetFreeInCircularBuffer = (nextOffsetFreeInCircularBuffer + 1) & 0x3ff
			bytesCopied++
		}
	}

	return outputBuffer, nil
}
