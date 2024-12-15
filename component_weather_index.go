package go_Weather_ITUR

type WeatherIndexComponent struct {
	T 				float64		// 2m temperature (K)
	P 				float64		// surface pressure	(hPa)
	V_t 			float64		// total column water vapour (kg/m2)
	rho				float64		// 
	precipitation 	float64 	// rain (mm/s)
	hr 				float64		// rain height (km)
}
