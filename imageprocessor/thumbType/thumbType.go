package thumbType

type ThumbType int
const (
  JPG ThumbType = iota
  PNG
  GIF
)

func (this ThumbType) ToString() string {
  if this == PNG {
    return "PNG"
  } else if this == GIF {
    return "GIF"
  } else {
    return "JPG"
  }
}
