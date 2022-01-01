package tags

import (
  "log"
  "strings"
  "encoding/binary"
)

// version 1, layer 3 bit rates and sample rates
var v1l3BitRates = []float64{ 0, 32000.0, 40000.0, 48000.0,
  56000.0, 64000.0, 80000.0, 96000.0,
  112000.0, 128000.0, 160000.0, 192000.0,
  224000.0, 256000.0, 320000.0 }
var v1l3SampleRates = []float64{ 44100.0, 48000.0, 32000.0 }

func Mp3TagsFromFile(path string) map[string]string {
  buffer := readFile(path)
  m := make(map[string]string)
  // Look at each byte.  If the byte is 0xff, check to see if the upper three bits
  // of the next byte are set.  If so, it is the start of a frame.  If not, check
  // to see if the byte is 0x49, which represents the letter 'I'.  If so, check
  // to see if it is followed by "D3".  If so, it is the start of the ID3 block.
  // If neither of these is true, then just ignore the byte and move on.
  numFrames := 0
  totalFrameBytes := 0
  duration := 0.0
  var increment int
  for n := 0; n < len(buffer); n += increment {
    increment = 1
    b := buffer[n]
    if b == 0xff {
      if (buffer[n+1] & 0xe0) == 0xe0 {
        numFrames++
        frameSize, frameDuration := mp3ParseFrame(buffer[n:], n)
        totalFrameBytes += frameSize
        increment = frameSize
        duration = duration + frameDuration
      }
    } else if b == 0x49 {
      if buffer[n+1] == 0x44 && buffer[n+2] == 0x33 {
        increment = mp3ParseID3(buffer[n:], m)
      }
    }
  }
  setDuration(duration, m)
  return m
}

// Returns the size of this frame in bytes and the duration of the sound
// in this frame.
func mp3ParseFrame(buffer []byte, offset int) (int, float64) {
  version := (buffer[1] >> 3) & 0x03
  layer := (buffer[1] >> 1) & 0x03
  // We only handle MP3 at this time.
  if version != 3 || layer != 1 {
    log.Fatalf("Got frame with version %d and layer %d at offset %d\n", version, layer, offset)
  }
  protection := (buffer[1] & 0x01) == 0
  bri := buffer[2] >> 4  // bit rate index
  sri := (buffer[2] >> 2) & 0x03  // sample rate index
  padding := (buffer[2] >> 1) & 0x01 == 0x01
  // We now have enough info to calculate the size of the frame.
  bitRate := v1l3BitRates[bri]
  sampleRate := v1l3SampleRates[sri]
  frameSize := int((144.0 * bitRate) / sampleRate)
  if padding {
    frameSize += 1
  }
  if protection {
    frameSize += 2
  }
  return frameSize, 1152.0 / sampleRate
}

func mp3ParseID3(buffer []byte, m map[string]string) int {
  headerSize := 10
  // Check for extended header
  if buffer[5] & 0x40 == 0x40 {
    headerSize += 4 + int(buffer[13])
  }
  // Start after the header, and read tags until we're through.
  // For now, let's just get the size.
  for j := headerSize; j < len(buffer); {
    key := string(buffer[j:j+4])
    size := int(binary.BigEndian.Uint32(buffer[j+4:j+8]))
    if strings.HasPrefix(key, "T") {
      // Ignore encoding for now.
      value := string(buffer[j+11:j+size+10])
      m[key] = value
    }
    j += size + 10
  }
  return mp3GetID3Size(buffer[6:]) + headerSize
}

func mp3GetID3Size(b []byte) int {
  // Read four bytes, use the lower 7 bits of each one to form a 28-bit size.
  var total int = 0
  for j := 0; j < 4; j++ {
    total <<= 7
    total += int(b[j]) & 0x7f
  }
  return total
}
