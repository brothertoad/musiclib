package common

const FullPathKey = "fullPath"
const BasePathKey = "basePath"
const TitleKey = "title"
const ArtistKey = "artist"
const AlbumKey = "album"
const TrackNumberKey = "trackNumber"
const DiscNumberKey = "discNumber"
const ArtistSortKey = "artistSort"
const AlbumSortKey = "albumSort"
const DurationKey = "duration"
const MimeKey = "mime"
const ExtensionKey = "extension"
const EncodedExtensionKey = "encodedExtension"
const IsEncodedKey = "isEncoded"
const FlagsKey = "flags"

const EncodeFlag = "e"
const PlayerFlag = "p"
const CarFlag = "c"

type Song map[string]string
