package tags

import (
  "strings"
  "github.com/brothertoad/musiclib/common"
)

func GetTagsFromFile(path string) common.SongMap {
  if strings.HasSuffix(path, "flac") {
    return FlacTagsFromFile(path)
  } else if strings.HasSuffix(path, "mp3") {
    return Mp3TagsFromFile(path)
  } else if strings.HasSuffix(path, "m4a") {
    return M4aTagsFromFile(path)
  }
  return make(common.SongMap)
}
