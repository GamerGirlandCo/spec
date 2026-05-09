package option

func optional[T any](fallback T, values ...T) T {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}
