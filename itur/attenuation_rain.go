package itur

import (
	"go_Weather_ITUR/utils"
	"math"
)

type Coefficients struct {
	AJ []float64
	BJ []float64
	CJ []float64
	M  float64
	C  float64
}

func CurveFunction(f, a, b, c float64) float64 {
	return a * math.Exp(-math.Pow((math.Log10(f)-b)/c, 2))
}

func sumCurveFunction(f float64, coef Coefficients) float64 {
	sum := 0.0
	for i := range coef.AJ {
		sum += CurveFunction(f, coef.AJ[i], coef.BJ[i], coef.CJ[i])
	}
	return sum
}

func RainSpecificAttenuationCoefficients(f, el, tau float64) (float64, float64) {
	kh := Coefficients{
		AJ: []float64{-5.33980, -0.35351, -0.23789, -0.94158},
		BJ: []float64{-0.10008, 1.2697, 0.86036, 0.64552},
		CJ: []float64{1.13098, 0.454, 0.15354, 0.16817},
		M:  -0.18961,
		C:  0.71147,
	}

	kv := Coefficients{
		AJ: []float64{-3.80595, -3.44965, -0.39902, 0.50167},
		BJ: []float64{0.56934, -0.22911, 0.73042, 1.07319},
		CJ: []float64{0.81061, 0.51059, 0.11899, 0.27195},
		M:  -0.16398,
		C:  0.63297,
	}

	alphah := Coefficients{
		AJ: []float64{-0.14318, 0.29591, 0.32177, -5.37610, 16.1721},
		BJ: []float64{1.82442, 0.77564, 0.63773, -0.96230, -3.29980},
		CJ: []float64{-0.55187, 0.19822, 0.13164, 1.47828, 3.4399},
		M:  0.67849,
		C:  -1.95537,
	}

	alphav := Coefficients{
		AJ: []float64{-0.07771, 0.56727, -0.20238, -48.2991, 48.5833},
		BJ: []float64{2.3384, 0.95545, 1.1452, 0.791669, 0.791459},
		CJ: []float64{-0.76284, 0.54039, 0.26809, 0.116226, 0.116479},
		M:  -0.053739,
		C:  0.83433,
	}

	// Compute KH and KV
	KH := math.Pow(10, sumCurveFunction(f, kh)+kh.M*math.Log10(f)+kh.C)
	KV := math.Pow(10, sumCurveFunction(f, kv)+kv.M*math.Log10(f)+kv.C)

	// Compute AlphaH and AlphaV
	AlphaH := sumCurveFunction(f, alphah) + alphah.M*math.Log10(f) + alphah.C
	AlphaV := sumCurveFunction(f, alphav) + alphav.M*math.Log10(f) + alphav.C

	// Compute k and alpha
	cosEl := math.Cos(utils.DegToRad(el))
	cos2Tau := math.Cos(utils.DegToRad(2 * tau))

	k := (KH + KV + (KH-KV)*math.Pow(cosEl, 2)*cos2Tau) / 2.0
	alpha := (KH*AlphaH + KV*AlphaV + (KH*AlphaH-KV*AlphaV)*math.Pow(cosEl, 2)*cos2Tau) / (2.0 * k)

	return k, alpha
}

const EPSILON = 1e-9

// 地面的经纬度lat, lon, 频率，倾角，地面站高度，不可用度，极化倾角，路径长度
// itu618
func RainAttenuation(lat, lon, f, el, hs, p, R001, tau, Ls float64) float64 {
	tau = 45
	Re := 8500.0 //地球半径km

	// step 1: hr降雨高度取 4km
	hr := 4.0

	// step 2: Ls
	if el >= 5 {
		Ls = (hr - hs) / math.Sin(utils.DegToRad(el))
	} else {
		sinEl := math.Sin(utils.DegToRad(el))
		Ls = 2 * (hr - hs) / (math.Sqrt(math.Pow(sinEl, 2)+2*(hr-hs)/Re) + sinEl)
	}

	// Step 3: Lg
	Lg := math.Abs(Ls * math.Cos(utils.DegToRad(el)))

	//step 5 : gammar
	//itu838
	k, alpha := RainSpecificAttenuationCoefficients(f, el, tau)
	gammar := k * math.Pow(R001, alpha)

	// step 6: r001
	r001 := 1.0 / (1.0 + 0.78*math.Sqrt(Lg*gammar/f) - 0.38*(1-math.Exp(-2*Lg)))

	//step 7
	eta := utils.DegToRad(math.Atan2(hr-hs, Lg*r001))
	Delta_h := math.Max(hr-hs, EPSILON)
	Lr := 0.0
	if eta > el {
		Lr = Lg * r001 / math.Cos(utils.DegToRad(el))
	} else {
		Lr = Delta_h / math.Sin(utils.DegToRad(el))
	}

	xi := 0.0
	if math.Abs(lat) < 36 {
		xi = 36 - math.Abs(lat)
	}

	v001 := 1.0 / (1.0 + math.Sqrt(math.Sin(utils.DegToRad(el)))*
		(31*(1-math.Exp(-(el/(1+xi))))*math.Sqrt(Lr*gammar)/math.Pow(f, 2)-0.45))

	// step 8:
	Le := Lr * v001

	// step 9:
	A001 := gammar * Le

	// step 10:
	beta := 0.0
	if p >= 1 {
		beta = 0.0
	} else if math.Abs(lat) >= 36 {
		beta = 0.0
	} else if el > 25 {
		beta = -0.005*(math.Abs(lat)-36) + 1.8 - 4.25*math.Sin(utils.DegToRad(el))
	} else {
		beta = -0.005 * (math.Abs(lat) - 36)
	}

	A := A001 * math.Pow(p/0.01, -(0.655+0.033*math.Log(p)-0.045*math.Log(A001)-beta*(1-p)*math.Sin(utils.DegToRad(el))))

	return A

}
