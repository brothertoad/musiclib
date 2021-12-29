package tags

import (
  "os"
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
