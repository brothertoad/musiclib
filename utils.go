package tags

import (
  "os"
  "math"
  "strconv"
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
  m["duration"] = strconv.Itoa(int(math.Round(duration)))
}
