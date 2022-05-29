package tags

import (
  "os"
  "math"
  "fmt"
  "github.com/brothertoad/musiclib/common"
)

func check(e error) {
  if e!= nil {
    panic(e)
  }
}

func readFile(path string) []byte {
  b, err := os.ReadFile(path)
  check(err)
  return b
}

func setDuration(duration float64, m common.SongMap) {
  // Round to nearest integer, make it a string, convert to mm:ss.
  totalSeconds := int(math.Round(duration))
  minutes := totalSeconds / 60
  seconds := totalSeconds % 60
  m[common.DurationKey] = fmt.Sprintf("%d:%02d", minutes, seconds)
}

func setMimeAndExtension(mime string, extension string, m common.SongMap) {
  m[common.MimeKey] = mime
  m[common.ExtensionKey] = extension
}
