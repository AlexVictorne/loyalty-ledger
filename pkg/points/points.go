package points

import "math"

// PointsPrecision определяет, сколько "сотых" в одной целой единице баллов.
const PointsPrecision = 100

// ToInternal переводит дробное значение баллов (API) во внутреннее целое представление (хранение).
func ToInternal(val float64) int64 {
	return int64(math.Ceil(val * PointsPrecision))
}

// ToAPI переводит внутреннее целое представление баллов в дробное для API/отображения.
func ToAPI(val int64) float64 {
	return float64(val) / PointsPrecision
}
