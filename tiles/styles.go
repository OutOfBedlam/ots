package tiles

import (
	"image/color"
)

type Style struct {
	FillColor       color.Color
	LineColor       color.Color
	LineWidth       float64
	LineDash        []float64
	MarkerColor     color.Color
	Marker          Icon
	MarkerZoomLimit int
	BaseLayer       Layer
}

type StyleParam struct {
	Tags   map[string]string
	Closed bool
}

type StyleFunc func(style *Style, p *StyleParam)

func (s *Style) MarkerVisible(zoom int) bool {
	if s.MarkerZoomLimit == 0 {
		return true
	}
	return zoom >= s.MarkerZoomLimit
}

type style_of_func struct {
	BaseTag string
	Func    func(*Style, string, *StyleParam)
}

//// osm tags: https://wiki.openstreetmap.org/wiki/Beginners_Guide_1.3
var styleFuncs = []style_of_func{
	{BaseTag: "type", Func: styleOfRelationType},          // 'type' is tag of Relation
	{BaseTag: "shop", Func: styleOfShop},                  // Used for tagging shops that you buy products from.
	{BaseTag: "building", Func: styleOfBuilding},          // Used for tagging buildings.
	{BaseTag: "building:part", Func: styleOfBuildingPart}, // Used for tagging buildings.
	{BaseTag: "amenity", Func: styleOfAmenity},            // Used for tagging useful amenities like restaurants, drinking water spots, parking lots, etc.
	{BaseTag: "place", Func: styleOfPlace},                // Used for tagging countries, cities, towns, villages, etc.
	{BaseTag: "highway", Func: styleOfHighway},            // For tagging highways, roads, paths, footways, cycleways, bus stops, etc.
	{BaseTag: "landuse", Func: styleOfLanduse},            // Used for tagging land being used by humans.
	{BaseTag: "natural", Func: styleOfNatural},            // Used for tagging natural land like woods.
	{BaseTag: "leisure", Func: styleOfLeisure},
	//// sub tags
	{BaseTag: "route", Func: styleOfRoute},
	{BaseTag: "man_made", Func: styleOfManMade},
	{BaseTag: "railway", Func: styleOfRailway},
	{BaseTag: "waterway", Func: styleOfWaterway},
	{BaseTag: "boundary", Func: styleOfBoundary},
	{BaseTag: "barrier", Func: styleOfBarrier},
	//// additional tags
	{BaseTag: "power", Func: styleOfPower},
}

func styleFromTags(p *StyleParam, customs ...StyleFunc) *Style {
	style := &Style{
		FillColor:   nil,
		LineColor:   nil,
		LineWidth:   1.0,
		LineDash:    nil,
		MarkerColor: Brown900,
		Marker:      nil,
		BaseLayer:   LayerBackground + 1,
	}
	for _, sf := range styleFuncs {
		if v, b := p.Tags[sf.BaseTag]; b {
			sf.Func(style, v, p)
		}
	}

	for _, custom := range customs {
		if custom == nil {
			continue
		}
		custom(style, p)
	}
	return style
}

//// https://wiki.openstreetmap.org/wiki/Types_of_relation
func styleOfRelationType(style *Style, typ string, p *StyleParam) {
	switch typ {
	default:
		//// Established relations
	case "multipolygon":
		//// https://wiki.openstreetmap.org/wiki/Relation:multipolygon
		style.FillColor = nil
		style.LineColor = Brown800
		if natural, b := p.Tags["natural"]; b {
			styleOfNatural(style, natural, p)
		} else if landuse, b := p.Tags["landuse"]; b {
			styleOfLanduse(style, landuse, p)
		} else if building, b := p.Tags["building"]; b {
			styleOfBuilding(style, building, p)
		} else if man_made, b := p.Tags["man_made"]; b {
			styleOfManMade(style, man_made, p)
		} else if amenity, b := p.Tags["amenity"]; b {
			styleOfAmenity(style, amenity, p)
		} else if leisure, b := p.Tags["leisure"]; b {
			styleOfLeisure(style, leisure, p)
		} else if highway, b := p.Tags["highway"]; b {
			styleOfHighway(style, highway, p)
		} else if waterway, b := p.Tags["waterway"]; b {
			styleOfWaterway(style, waterway, p)
		} else if water, b := p.Tags["water"]; b {
			styleOfWater(style, water, p)
		}
	case "route":
		route, _ := p.Tags["route"]
		styleOfRoute(style, route, p)
	case "route_master":
		style.FillColor = nil
	case "restriction":
		style.FillColor = nil
	case "boundary":
		boundary, _ := p.Tags["boundary"]
		styleOfBoundary(style, boundary, p)
	case "public_transport":
		style.FillColor = nil
	case "destination_sign":
		style.FillColor = nil
	case "waterway":
		style.FillColor = nil
	case "enforcement":
		style.FillColor = nil
	case "connectivity":
		style.FillColor = nil
	case "leagal":
		style.FillColor = nil
		style.LineColor = Purple400
		style.LineDash = []float64{4.0}
		style.BaseLayer = LayerBorder + 1

		//// Uncommon relations
	case "associatedStreet":
		style.FillColor = nil
	case "superroute":
		style.FillColor = nil
	case "site":
		style.FillColor = nil
	case "network":
		style.FillColor = nil
	case "building":
		if building, b := p.Tags["building"]; b {
			styleOfBuilding(style, building, p)
		}
	case "street":
		style.FillColor = nil
	case "bridge":
		style.FillColor = nil
	case "tunnel":
		style.FillColor = nil
	}
}

func styleOfHighway(style *Style, highway string, p *StyleParam) {
	style.FillColor = nil
	style.MarkerColor = BlueGray900
	style.BaseLayer = LayerRoad

	switch highway {
	default:
		style.LineWidth = 2.0
		style.LineColor = Red400
	case "path":
		fallthrough
	case "steps":
		style.LineWidth = 2.0
		style.LineDash = []float64{3}
		style.LineColor = DeepPurple400
		style.MarkerZoomLimit = 17
	case "pedestrian": // 보행자전용
		style.LineWidth = 1.0
		style.LineDash = []float64{3}
		style.LineColor = BlueGray400
		style.MarkerZoomLimit = 17
	case "footway":
		style.LineWidth = 1.0
		style.LineDash = []float64{3}
		style.LineColor = BlueGray400
		style.MarkerZoomLimit = 17
	case "service":
		style.LineWidth = 2.0
		style.LineDash = nil
		style.LineColor = Red400
		style.MarkerZoomLimit = 17
	case "residential":
		style.LineWidth = 2.0
		style.LineDash = nil
		style.LineColor = Red400
		style.MarkerZoomLimit = 17
	case "tertiary":
		style.LineWidth = 2.0
		style.LineDash = nil
		style.LineColor = Red400
		style.MarkerZoomLimit = 16
	case "secondary":
		style.LineWidth = 3.0
		style.LineDash = nil
		style.LineColor = Red400
		style.MarkerZoomLimit = 15
	case "secondary_link":
		style.LineWidth = 3.0
		style.LineDash = nil
		style.LineColor = Red400
		style.MarkerZoomLimit = 15
	case "trunk":
		style.LineWidth = 3.0
		style.LineDash = nil
		style.LineColor = Red400
	case "trunk_link":
		style.LineWidth = 3.0
		style.LineDash = nil
		style.LineColor = Red400
	case "primary_link":
		style.LineWidth = 3.0
		style.LineDash = nil
		style.LineColor = Red400
	case "primary":
		style.LineWidth = 5.0
		style.LineDash = nil
		style.LineColor = Red400
	}
}

func styleOfAmenity(style *Style, amenity string, p *StyleParam) {
	style.BaseLayer = LayerAmenity
	//// https://wiki.openstreetmap.org/wiki/Key:amenity
	switch amenity {
	default:
		//// Sustenance
	case "bar":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "biergarten":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "cafe":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "fast_food":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "food_court":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "ice_cream":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "pub":
		style.FillColor = Orange300
		style.LineColor = Orange700
	case "restaurant":
		style.FillColor = Orange300
		style.LineColor = Orange700
		//// Education
	case "college":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
		style.Marker = fa_school
		style.MarkerZoomLimit = 15
	case "driving_school":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
	case "kindergarten":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
		style.Marker = fa_child
		style.MarkerZoomLimit = 16
	case "language_school":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
	case "library":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
		style.Marker = fa_book
		style.MarkerZoomLimit = 16
	// case "toy_library":
	// case "music_school":
	case "school":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
		style.Marker = fa_school
		style.MarkerZoomLimit = 16
	case "university":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
		style.Marker = fa_university
		style.MarkerZoomLimit = 15
		//// TODO: Transportation
	case "parking":
		style.FillColor = BlueGray50
		style.LineColor = BlueGray300
		style.Marker = fa_parking
		style.MarkerZoomLimit = 16
		//// TODO: Financial
		//// Healthcare
	case "hospital":
		style.FillColor = Red50
		style.LineColor = Red200
		style.Marker = fa_hostpital_alt
		style.MarkerZoomLimit = 15
		//// TODO: Entertainment, Aarts & Curture
		//// TODO: Public Service
	case "police":
		style.FillColor = Red50
		style.LineColor = Red200
		style.Marker = fa_user_shield
	case "fire_station":
		style.FillColor = Red50
		style.LineColor = Red200
		style.Marker = fa_fire
		//// TODO: Facilities
		//// TODO: Waste Management
		//// TODO: Others
	}
}

func styleOfBuilding(style *Style, building string, p *StyleParam) {
	style.BaseLayer = LayerBuilding

	switch building {
	default:
		style.FillColor = Gray400
		style.LineColor = Gray600
	case "stadium":
		style.FillColor = Lime100
		style.LineColor = Lime500
	}
}

func styleOfBuildingPart(style *Style, building string, p *StyleParam) {
	style.BaseLayer = LayerBuilding
	style.FillColor = Gray400
	style.LineColor = Gray600
}

func styleOfPlace(style *Style, place string, p *StyleParam) {
	style.BaseLayer = LayerPlace

	switch place {
	default:
	case "square":
		style.FillColor = BlueGray50
		style.LineColor = BlueGray300
	case "village":
		style.FillColor = Orange50
		style.LineColor = Orange100
	}
}

func styleOfLanduse(style *Style, landuse string, p *StyleParam) {
	style.BaseLayer = LayerLanduse

	switch landuse {
	default:
	case "residential":
		style.FillColor = Orange50
		style.LineColor = Orange100
	case "commercial":
		style.FillColor = Indigo50
		style.LineColor = Indigo100
	case "military":
		style.FillColor = Brown50
		style.LineColor = Brown100
	case "forest":
		style.FillColor = Green700
		style.LineColor = nil
	case "grass":
		style.FillColor = LightGreen200
		style.LineColor = LightGreen400
	case "farmland":
		style.FillColor = LightGreen100
		style.LineColor = LightGreen200
	case "stadium":
		style.FillColor = Lime100
		style.LineColor = Lime500
	case "education":
		style.FillColor = Cyan50
		style.LineColor = Cyan400
		style.Marker = fa_school
	case "railway":
		style.FillColor = Gray200
		style.LineColor = nil
	}
}

func styleOfNatural(style *Style, natural string, p *StyleParam) {
	style.LineColor = nil
	style.BaseLayer = LayerNature

	switch natural {
	default:
		//// Vegetation
	case "fell":
		style.FillColor = LightBlue100
	case "grassland":
		style.FillColor = LightGreen200
	case "heath":
		style.FillColor = LightGreen50
	case "scrub":
		style.FillColor = LightGreen300
	case "wood":
		style.FillColor = Green500
		style.LineColor = Green700
		//// Water related
	case "water":
		style.FillColor = LightBlue100
	case "bay":
		style.FillColor = LightBlue100
	case "beach":
		style.FillColor = Amber100
	case "wetland":
		style.FillColor = Gray300
	case "coastline":
		style.LineWidth = 2
		style.LineColor = Blue600
		style.FillColor = nil
		//// Geology related
	case "sand":
		style.FillColor = Amber100
	}
}

func styleOfShop(style *Style, shop string, p *StyleParam) {
	style.BaseLayer = LayerBuilding
	style.FillColor = Gray400
	style.LineColor = Gray600
	// app == "supermarket"
}

func styleOfLeisure(style *Style, leisure string, p *StyleParam) {
	switch leisure {
	default:
	case "stadium":
		style.FillColor = Lime100
		style.LineColor = Lime500
		style.BaseLayer = LayerLanduse
	case "sports_centre":
		style.FillColor = Lime100
		style.LineColor = Lime500
		style.BaseLayer = LayerLanduse
	case "track":
		style.FillColor = Lime300
		style.LineColor = Lime700
		style.BaseLayer = LayerLanduse
	case "schoolyard":
		style.FillColor = LightGreen100
		style.LineColor = LightGreen400
		style.BaseLayer = LayerLanduse
	case "park":
		style.FillColor = LightGreen200
		style.LineColor = LightGreen500
		style.BaseLayer = LayerLanduse
	case "garden":
		style.FillColor = LightGreen400
		style.LineColor = LightGreen800
		style.BaseLayer = LayerLanduse
	case "pitch":
		style.FillColor = Teal300
		style.LineColor = Teal500
		style.BaseLayer = LayerLanduse
	case "commercial":
		style.FillColor = Indigo50
		style.LineColor = Indigo100
	}
}

func styleOfManMade(style *Style, man_made string, p *StyleParam) {
	switch man_made {
	default:
	case "bridge":
		style.FillColor = Gray200
		style.LineColor = Gray500
		style.LineWidth = 4.0
		style.BaseLayer = LayerBuilding
	case "wastewater_plant":
		style.FillColor = Gray200
		style.LineColor = Gray500
		style.BaseLayer = LayerLanduse
	}
}

func styleOfWaterway(style *Style, waterway string, p *StyleParam) {
	style.BaseLayer = LayerRoute
	style.FillColor = nil
	style.LineColor = Blue800
	switch waterway {
	default:
	}
}

func styleOfWater(style *Style, waterway string, p *StyleParam) {
	style.BaseLayer = LayerNature
	switch waterway {
	default:
	case "river":
		style.FillColor = Blue100
	}
}

func styleOfRoute(style *Style, route string, p *StyleParam) {
	//// https://wiki.openstreetmap.org/wiki/Key:route
	style.FillColor = nil
	style.LineColor = Brown800
	style.BaseLayer = LayerRoute
	switch route {
	case "ferry":
		style.LineColor = Blue900
		style.LineDash = []float64{4.0}
	}
}

func styleOfRailway(style *Style, railway string, p *StyleParam) {
	style.FillColor = nil
	style.LineColor = Brown800
	style.BaseLayer = LayerRoute
	switch railway {
	case "construction":
		style.LineColor = Brown400
		style.LineDash = []float64{10.0, 10.0}
	}
}

func styleOfPower(style *Style, power string, p *StyleParam) {
	style.FillColor = nil
	style.LineColor = Brown600
	style.BaseLayer = LayerRoute
	style.LineDash = []float64{2.0, 8.0}
}

func styleOfBoundary(style *Style, boundary string, p *StyleParam) {
	style.FillColor = nil
	style.LineColor = DeepPurple400
	style.BaseLayer = LayerBorder
	switch boundary {
	default:
	case "postal_code":
		//// example  REL:13682339,13682340
		style.LineColor = BlueGray200
	case "administrative":
		style.LineWidth = 4.0
		style.LineDash = []float64{8.0, 12.0, 2.0, 12.0}
	}
}

func styleOfBarrier(style *Style, boundary string, p *StyleParam) {
	style.FillColor = nil
	style.LineColor = Brown400
	style.LineWidth = 1.0
	style.LineDash = []float64{4.0}
	style.BaseLayer = LayerBorder

	// barrier=fence
}
