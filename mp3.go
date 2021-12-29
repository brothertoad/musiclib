package tags

import (
  "log"
  "bytes"
  "encoding/binary"
)

var keysToSave = []string{ "TPE1", "TPE2", "TIT2", "TALB", "TSOA", "TSO2", "TRCK", "TPOS" }

func Mp3TagsFromFile(path string) map[string]string {
  bb := bytebufferfromfile(path)
  if string(bb.read(3)) != "ID3" {
    log.Fatalf("mp3 file %s does not have correct magic number\n", path)
  }
  // Read and ignore major version, minor version and flags
  _ = bb.read(3)
  totalSize := mp3GetTotalSize(bb)
  tbb := bytebufferfromparent(bb, totalSize)
  return mp3BruteForce(tbb)
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

func saveKey(x string) bool {
    for _, n := range keysToSave {
        if x == n {
            return true
        }
    }
    return false
}
