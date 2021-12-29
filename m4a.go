package tags

import (
  "fmt"
)

const moov = "moov"
const udta = "udta"
const meta = "meta"
const ilst = "ilst"

const trackkey = "trkn"
const diskkey = "disk"

func M4aTagsFromFile(path string) map[string]string {
  bb := bytebufferfromfile(path)
  m := make(map[string]string)
  moovatom := findatom(bb, moov)
  udtaatom := findatom(moovatom, udta)
  metaatom := findatom(udtaatom, meta)
  // We need to skip four bytes from the meta atom
  metaatom.skip(4)
  ilstatom := findatom(metaatom, ilst)
  readm4atags(ilstatom, m)
  return m
}

func readm4atags(bb *bytebuffer, m map[string]string) {
  keys := [...]string{ "\xa9nam", "\xa9ART", "\xa9alb", "soar", "soal" }
  for bb.remaining() > 0 {
    size := bb.read32BE();
    atomtype := string(bb.read(4))
    found := false
    // Look for text keys first
    for _, key := range keys {
      if atomtype == key {
        // Skip 16 bytes, which is the size of the data, the word "data", the type and locale
        bb.skip(16)
        value := string(bb.read(size - 24))
        m[key] = value
        found = true
      }
    }
    if !found {
      // Handle trkn and disk separate, since they are funky.
      if atomtype == trackkey {
        bb.skip(18)
        track := bb.read16BE()
        bb.skip(4)
        m[trackkey] = fmt.Sprintf("%d", track)
        found = true
      } else if atomtype == diskkey {
        bb.skip(18) // note: skip first byte of trkn
        disk := bb.read16BE()
        bb.skip(2)
        m[diskkey] = fmt.Sprintf("%d", disk)
        found = true
      }
    }
    if !found {
      bb.skip(size - 8)
    }
  }
}

func findatom(bb *bytebuffer, magic string) *bytebuffer {
  for bb.remaining() > 0 {
    size := bb.read32BE()
    atomtype := string(bb.read(4))
    if atomtype == magic {
      return bytebufferfromparent(bb, size - 8)
    }
    bb.skip(size - 8)
  }
  return nil
}
