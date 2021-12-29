package main

import (
  "os"
  "bufio"
  "fmt"
  "strings"
  "log"
  "github.com/brothertoad/tags"
)

// Test program for the tag module.  Input can be either a flac file, an mp3 file,
// an m4a file, or a list file.  A list file contains a list of audio files,
// one per line, such as generated by the find command.

func main() {
  for j:= 1; j < len(os.Args); j++ {
    path := os.Args[j]
    if strings.HasSuffix(path, "list") {
      file, err := os.Open(path)
      if err != nil {
        log.Fatal(err)
      }
      defer file.Close()
      scanner := bufio.NewScanner(file)
      for scanner.Scan() {
          dumpFile(scanner.Text())
      }
      if err := scanner.Err(); err != nil {
          log.Fatal(err)
      }    } else {
      dumpFile(path)
    }
  }
}

func dumpFile(path string) {
  var m map[string]string
  if strings.HasSuffix(path, "flac") {
    m = tags.FlacTagsFromFile(path)
  } else if strings.HasSuffix(path, "m4a") {
    m = tags.M4aTagsFromFile(path)
  } else if strings.HasSuffix(path, "mp3") {
    m = tags.Mp3TagsFromFile(path)
  } else {
    fmt.Printf("Unknown suffix on file %s\n", path)
    return
  }
  fmt.Printf("file %s has %d tags\n", path, len(m))
  for key, value := range m {
    fmt.Printf("%s: %s\n", key, value)
  }
}
