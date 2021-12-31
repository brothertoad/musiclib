package tags

import (
  "log"
  "strings"
)

const magic = 0x664c6143
const streaminfotype byte = 0
const commenttype byte = 4

func FlacTagsFromFile(path string) map[string]string {
  bb := bytebufferfromfile(path)
  if bb.read32BE() != magic {
    log.Fatalf("flac file %s does not have correct magic number\n", path)
  }
  m := make(map[string]string)
  for {
    blocktype, lastone, size := nextmetablock(bb)
    if blocktype == commenttype {
      cbb := bytebufferfromparent(bb, size)
      getFlacComments(cbb, m)
    } else if blocktype == streaminfotype {
      sibb := bytebufferfromparent(bb, size)
      getFlacDuration(sibb, m)
    } else {
      bb.skip(size)
    }
    if lastone {
      break
    }
  }
  return m
}

func getFlacComments(cbb *bytebuffer, m map[string]string) {
  vendorsize := cbb.read32LE()
  cbb.skip(vendorsize)
  num := cbb.read32LE()
  for j:= 0; j < int(num); j++ {
    size := cbb.read32LE()
    comment := string(cbb.read(size))
    parts := strings.Split(comment, "=")
    m[parts[0]] = parts[1]
  }
}

func getFlacDuration(bb *bytebuffer, m map[string]string) {
  bb.skip(10)
  // We're going to do a shortcut, and assume the upper four bits of the
  // total samples are zero.  This is good to over 750 minutes.
  sampleSize := float64(bb.read32BE() >> 12)
  numSamples := float64(bb.read32BE())
  setDuration(numSamples / sampleSize, m)
}

func nextmetablock(bb *bytebuffer) (byte, bool, uint32) {
  blocktype := bb.peek()
  lastone := blocktype > 127
  if lastone {
    blocktype -= 128
  }
  size := bb.read32BE()
  size &= 0x00ffffff
  return blocktype, lastone, size
}
