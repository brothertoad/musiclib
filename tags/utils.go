package tags

import (
  "os"
  "math"
  "fmt"
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

func setDuration(duration float64, m map[string]string) {
  // Round to nearest integer, then make it a string.
  // Should probably convert to mm:ss here
  totalSeconds := int(math.Round(duration))
  minutes := totalSeconds / 60
  seconds := totalSeconds % 60
  m["duration"] = fmt.Sprintf("%d:%02d", minutes, seconds)
}
