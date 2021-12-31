package tags

import (
  "os"
  "fmt"
  "io"
  "log"
  "strings"
  "bytes"
  "encoding/binary"
  "github.com/tcolgate/mp3"
)

var keysToSave = []string{ "TPE1", "TPE2", "TIT2", "TALB", "TSOA", "TSO2", "TRCK", "TPOS" }

// version 1, layer 3 bit rates and sample rates
var v1l3BitRates = []float64{ 0, 32000.0, 40000.0, 48000.0,
  56000.0, 64000.0, 80000.0, 96000.0,
  112000.0, 128000.0, 160000.0, 192000.0,
  224000.0, 256000.0, 320000.0 }
var v1l3SampleRates = []float64{ 44100.0, 48000.0, 32000.0 }

func Mp3TagsFromFile(path string) map[string]string {
  mp3ParseFile(path)
  bb := bytebufferfromfile(path)
  if string(bb.read(3)) != "ID3" {
    log.Fatalf("mp3 file %s does not have correct magic number\n", path)
  }
  // Read and ignore major version, minor version and flags
  _ = bb.read(3)
  totalSize := mp3GetTotalSize(bb)
  tbb := bytebufferfromparent(bb, totalSize)
  m := mp3BruteForce(tbb)
  getMp3Duration(path, m)
  return m
}

func mp3ParseFile(path string) {
  buffer := readFile(path)
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
        increment = mp3ParseID3(buffer[n:])
      }
    }
  }
  fmt.Printf("Found %d frames, totaling %d bytes, with duration %f.\n", numFrames, totalFrameBytes, duration)
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

func mp3ParseID3(buffer []byte) int {
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
      fmt.Printf("Should add key %s, value %s to map\n", key, value)
    }
    j += size + 10
  }
  return mp3GetID3Size(buffer[6:]) + headerSize
}

func mp3BruteForce(bb *bytebuffer) map[string]string {
  // Search for tags desired.  If a tag is found, get its length, skip flags,
  // check encoding, then get value.
  m := make(map[string]string)
  for _, key := range keysToSave {
    n := bytes.Index(bb.b, []byte(key))
    if n >= 0 {
      size := binary.BigEndian.Uint32(bb.b[n+4:n+8])
      m[key] = string(bb.b[n+11:n+int(size)+10])
    }
  }
  return m
}

func mp3GetTotalSize(bb *bytebuffer) uint32 {
  // Read four bytes, use the lower 7 bits of each one to form a 28-bit size.
  var total uint32 = 0
  for j := 0; j < 4; j++ {
    total <<= 7
    total += bb.readByte() & 0x7f
  }
  return total
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

func saveKey(x string) bool {
    for _, n := range keysToSave {
        if x == n {
            return true
        }
    }
    return false
}

func getMp3Duration(path string, m map[string]string) {
  duration := 0.0
  r, err := os.Open(path)
  if err != nil {
    return
  }
  d := mp3.NewDecoder(r)
  var f mp3.Frame
  skipped := 0
  for {
    if err := d.Decode(&f, &skipped); err != nil {
      if err == io.EOF {
        break
      }
      log.Fatalf("Error getting duration from %s: %s\n", path, err.Error())
      break
    }
    duration = duration + f.Duration().Seconds()
  }
  setDuration(duration, m)
}
