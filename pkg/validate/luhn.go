package validate

// IsValidLuhn проверяет строку на соответствие алгоритму Луна и только цифры
func IsValidLuhn(number string) bool {
	if len(number) <= 1 {
		return false
	}
	var sum int
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if d < 0 || d > 9 {
			return false
		}
		if double {
			d = d * 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}
