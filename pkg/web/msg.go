package web

type Msg[T any] struct {
	Data T      `json:"data"`
	Err  string `json:"err,omitempty"`
}

func NewMsg[T any](data T, err ...error) Msg[T] {
	m := Msg[T]{
		Data: data,
		Err:  "",
	}
	for _, e := range err {
		if e != nil {
			m.Err = err[0].Error()
			break
		}
	}
	return m
}

type Error struct {
	Err string `json:"err,omitempty"`
}

func NewError(err error) Error {
	e := Error{}
	if err != nil {
		e.Err = err.Error()
	}
	return e
}
