package tiles

type Layer = int

const (
	LayerBackground = Layer(0x00000000)
	LayerNature     = Layer(0x000000F0)
	LayerLanduse    = Layer(0x00000F00)
	LayerPlace      = Layer(0x0000F000)
	LayerAmenity    = Layer(0x000F0000)
	LayerRoad       = Layer(0x000FF000)
	LayerBuilding   = Layer(0x00F00000)
	LayerRoute      = Layer(0x0F000000)
	LayerBorder     = Layer(0x0F000000)
	LayerAero       = Layer(0xF0000000)
	LayerLabel      = Layer(0xFFFFFF00)
	LayerWatermark  = Layer(0xFFFFFFFF)
)

//// true: z-order 아래로, false: z-order 위로
func LayerCompareOrder(lo Object, ro Object) bool {
	if label, ok := lo.(*Label); ok {
		switch other := ro.(type) {
		default:
			return false
		case *Label:
			return label.Layer() < other.Layer()
		}
	} else if obj, ok := lo.(*PolygonObject); ok {
		switch other := ro.(type) {
		case *Label:
			return true // true: z-order 아래로,   false: z-order 위로
		case *PolygonObject:
			if (obj.layer < LayerPlace || other.layer < LayerPlace) && obj.Layer() != other.Layer() {
				return obj.Layer() < other.Layer()
			}

			if obj.fillColor != nil && other.fillColor != nil {
				return obj.Area() > other.Area()
			} else if obj.fillColor != nil {
				return true // true: z-order 아래로,   false: z-order 위로
			} else {
				return false // true: z-order 아래로,   false: z-order 위로
			}
		case *MultiPolygonObject:
			if (obj.layer < LayerPlace || other.layer < LayerPlace) && obj.Layer() != other.Layer() {
				return obj.Layer() < other.Layer()
			}

			if obj.fillColor != nil && other.fillColor != nil {
				return obj.Area() > other.Area()
			} else if obj.fillColor != nil {
				return true // true: z-order 아래로,   false: z-order 위로
			} else {
				return false // true: z-order 아래로,   false: z-order 위로
			}
		default:
			return obj.Layer() < other.Layer()
		}
	} else if obj, ok := lo.(*MultiPolygonObject); ok {
		switch other := ro.(type) {
		case *Label:
			return true // true: z-order 아래로,   false: z-order 위로
		case *PolygonObject:
			if (obj.layer < LayerPlace || other.layer < LayerPlace) && obj.Layer() != other.Layer() {
				return obj.Layer() < other.Layer()
			}

			if obj.fillColor != nil && other.fillColor != nil {
				return obj.Area() > other.Area()
			} else if obj.fillColor != nil {
				return true // true: z-order 아래로,   false: z-order 위로
			} else {
				return false // true: z-order 아래로,   false: z-order 위로
			}
		case *MultiPolygonObject:
			if (obj.layer < LayerPlace || other.layer < LayerPlace) && obj.Layer() != other.Layer() {
				return obj.Layer() < other.Layer()
			}

			if obj.fillColor != nil && other.fillColor != nil {
				return obj.Area() > other.Area()
			} else if obj.fillColor != nil {
				return true // true: z-order 아래로,   false: z-order 위로
			} else {
				return false // true: z-order 아래로,   false: z-order 위로
			}
		default:
			return obj.Layer() < other.Layer()
		}
	}
	return false
}
