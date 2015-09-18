package thumbType

type ThumbType int

const (
	UNKNOWN ThumbType = iota
	JPG
	PNG
	GIF
	WEBP
)

func (this ThumbType) ToString() string {
	switch this {
	case JPG:
		return "JPG"
	case PNG:
		return "PNG"
	case GIF:
		return "GIF"
	case WEBP:
		return "WEBP"
	default:
		return "UNKNOWN"
	}
}

func FromMime(mime string) ThumbType {
	switch mime {
	case "image/jpeg":
		return JPG
	case "image/png":
		return PNG
	case "image/gif":
		return GIF
	case "image/webp":
		return WEBP
	default:
		return UNKNOWN
	}
}

func FromString(format string) ThumbType {
	switch format {
	case "jpg":
		return JPG
	case "jpeg":
		return JPG
	case "png":
		return PNG
	case "gif":
		return GIF
	case "webp":
		return WEBP
	default:
		return UNKNOWN
	}
}
