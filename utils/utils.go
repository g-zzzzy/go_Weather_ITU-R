package utils

import (
	"fmt"
	"math"
)

// WGS84 椭球参数
const (
	A = 6378137.0      // 半长轴
	B = 6356752.314245 // 半短轴
	F = (A - B) / A    // 扁率
)

// XYZToLatLonAlt 将 XYZ 坐标转换为经度、纬度和高度。
// x, y, z: 地心 X、Y、Z 坐标（米）
// 返回值：纬度（弧度）、经度（弧度）、高度（米）
func XYZToLatLonAlt(x, y, z float64) (float64, float64, float64) {
	lon := math.Atan2(y, x)

	p := math.Sqrt(x*x + y*y)
	latInit := math.Atan2(z, (1-F)*p) // 初始纬度值

	lat := 0.0
	for i := 0; i < 10; i++ { // 迭代计算纬度
		N := A / math.Sqrt(1-F*F*math.Sin(latInit)*math.Sin(latInit))
		h := p/math.Cos(latInit) - N
		lat = math.Atan2(z, (1-F*(N/(N+h)))*p)
		latInit = lat
	}

	return lat, lon, p/math.Cos(lat) - A/math.Sqrt(1-F*F*math.Sin(lat)*math.Sin(lat))
}

func DegToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

func clip(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func Elevation_angle(h, lat_s, lon_s, lat, lon float64) float64 {
	RE := 6371.0 //km
	rs := RE + h

	lat1 := DegToRad(lat)
	lat2 := DegToRad(lat_s)
	lon1 := DegToRad(lon)
	lon2 := DegToRad(lon_s)

	cosGamma := clip(math.Sin(lat2)*math.Sin(lat1)+math.Cos(lat1)*math.Cos(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1), -1, 1)
	gamma := math.Acos(cosGamma)

	elevation := math.Acos(math.Sin(gamma) / math.Sqrt(1+(RE/rs)*(RE/rs)-2*(RE/rs)*math.Cos(gamma)))
	return elevation
}

func main() {
	x := 6046175.565
	y := 1794851.807
	z := 3214178.232

	lat, lon, alt := XYZToLatLonAlt(x, y, z)

	// 将弧度转换为度
	latDeg := lat * 180 / math.Pi
	lonDeg := lon * 180 / math.Pi

	fmt.Printf("纬度: %.6f°\n", latDeg)
	fmt.Printf("经度: %.6f°\n", lonDeg)
	fmt.Printf("高度: %.2f 米\n", alt)
}
