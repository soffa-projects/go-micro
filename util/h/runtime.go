package h

func Op[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}

/*
func Op[T any](cb func() (T, error)) T {
  out, err := cb()
  if err != nil {
    panic(err)
  }
  return out
}
*/
