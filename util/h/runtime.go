package h

func F[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}

func RaiseAny(err error) {
	if err != nil {
		panic(err)
	}
}
func RaiseIf(test bool, err error) {
	if test && err != nil {
		panic(err)
	}
}

/*
func F[T any](cb func() (T, error)) T {
  out, err := cb()
  if err != nil {
    panic(err)
  }
  return out
}
*/
