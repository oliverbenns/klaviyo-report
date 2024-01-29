package conv

func Ptr[V any](v V) *V {
	return &v
}

func Val[V any](v *V) V {
	if v == nil {
		var zero V
		return zero
	}
	return *v
}
