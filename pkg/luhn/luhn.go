package luhn

// Возвращает 0 если число корректно согласно алгоритму луна, 1 если нет и -1 если передали не число
func Validate(s string) int {
	if len(s) == 0 {
		return -1
	}

	parity := len(s) % 2
	sum := 0

	for i, c := range s {
		num := int(c - '0')
		if num > 9 || num < 0 {
			return -1
		}

		if i%2 == parity {
			num *= 2
			if num > 9 {
				num -= 9
			}
		}
		sum += num
	}

	if sum%10 == 0 {
		return 0
	}
	return 1
}
