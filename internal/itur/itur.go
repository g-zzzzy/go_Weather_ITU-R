package itur

import (
	"log"
)

// 初始化计算雨衰的变量的值
// 除了 lat, lon, el, f := 22.5, D := 1.2, p := 0.1
// 即 hs:地面站高度， P：压强， V_t, rho
func Atmospheric_attenuation_slant_path(
	lat, lon, f, el, p, D, hs, rho, R001, eta, T, H, P, hL, Ls, tau, V_t float64,
	mode, returnContributions, includeRain, includeGas, includeScintillation, includeClouds bool,
) float64 {
	// f : (GHz)
	// el : (degree)
	// p : Percentage of the time the rain attenuation value is exceeded.
	// D : diameter of the earth-station antenna (m)
	// hs: 地面站海拔(km)
	// rho: water vapor density (g/m3)
	// R001: precipitation (mm/h)
	// eta: antenna efficiency. Default 0.5
	// T: (°C)
	// H: Average surface relative humidity (%)
	// P: surface pressure (hPa)
	// hL: Height of the turbulent layer (m). Default value 1000 m
	// Ls: Slant path length (km). If not provided, it will be computed using the
	// 		rain height and the elevation angle.
	// tau: Polarization tilt angle. Default value is 45
	// V_t: Integrated water vapour content along the path (kg/m2 or mm)
	// mode : Mode for the calculation of gaseous attenuation. 'approx', 'exact'
	if p < 0.001 || p > 50 {
		log.Println("Warning: The method to compute the total atmospheric attenuation is only recommended for p between 0.001% and 50%.")
	}
	// p_c_g := math.Max(1, p)

	if hs == 0 {
		hs = 100.0 //m
	}

	if P == 0 {
		P = 1013.0 //hPa
	}

	if V_t == 0 {
		V_t = 10.0
	}

	if rho == 0 {
		rho = 0.1
	}

	var Ar float64

	if includeRain {
		//地面的经纬度lat, lon, 频率，倾角，地面站高度，不可用度，极化倾角，路径长度
		//要传入R001
		Ar = RainAttenuation(lat, lon, f, el, hs, p, R001, tau, Ls)
	}
	return Ar
}
