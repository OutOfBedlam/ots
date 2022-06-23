package geom

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

type LatLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Point orb.Point

func (p LatLon) Point() Point {
	// orb uses (lng, lat) coordinate
	return Point{p.Lon, p.Lat}
}

func LatLonFromPoint(p Point) LatLon {
	return LatLon{Lat: p[1], Lon: p[0]}
}

func (p LatLon) DistanceEuclidean(p2 LatLon) float64 {
	return math.Sqrt(math.Pow(p2.Lon-p.Lon, 2) + math.Pow(p2.Lat-p.Lat, 2))
}

func (p LatLon) Bearing(p2 LatLon) float64 {
	return Bearing(p.Point(), p2.Point())
}

func BearingCoords(p1Lat, p1Lon, p2Lat, p2Lon float64) float64 {
	return geo.Bearing(orb.Point{p1Lon, p1Lat}, orb.Point{p2Lon, p2Lat})
}

func (p LatLon) String() string {
	return fmt.Sprintf("%.5f,%.5f", p.Lat, p.Lon)
}

/*
	직선 AB위에 점C를 수직 투사하여 직교하는 점 D를 구하기 위해서

			* C
			|
			|
			|
	--------+---------
	A       D         B

	Dx = Ax + t(Bx-Ax)
	Dy = Ay + t(By-Ay)
	(Dx-Cx)(Bx-Ax) + (Dy-Cy)(By-Ay) = 0   (벡터 CD와 AB가 직교하므로 두 벡터의 내적은 0)
	알고 있는 값 A,B,C에 대해 t를 풀면 아래와 같다.
	t = [ (Cx-Ax)(Bx-Ax) + (Cy-Ay)(By-Ay) ] / [ (Bx-Ax)^2 + (By-Ay)^2 ]
	이제 t를 대입하여 D(x,y)를 구할 수 있다.

	주의) Euclid공간에서 성립하는 이 규칙을 구체인 위경도상에 적용하기 위해서는 충분이 좁은 지역에서만 가능하다.
*/
func PerpendicularPoint(a, b LatLon, c LatLon) (LatLon, bool) {
	ax, ay := a.Lon, a.Lat
	bx, by := b.Lon, b.Lat
	cx, cy := c.Lon, c.Lat

	t := ((cx-ax)*(bx-ax) + (cy-ay)*(by-ay)) / (math.Pow(bx-ax, 2) + math.Pow(by-ay, 2))
	dx := ax + t*(bx-ax)
	dy := ay + t*(by-ay)

	if dx != dx || dy != dy { /// NaN
		return LatLon{Lon: math.NaN(), Lat: math.NaN()}, false
	}
	// 점 D(x,y)가 부분직선 AB에 포함되는 점인지 외부의 점인지 확인한다.
	inside := dx >= math.Min(ax, bx) && dy >= math.Min(ay, by) && dx <= math.Max(ax, bx) && dy <= math.Max(ay, by)
	return LatLon{Lon: dx, Lat: dy}, inside
}
