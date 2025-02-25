package game

var Shells = make([]Shell, 0, 8)

func RemoveFirst[T any](s *[]T) {
	*s = (*s)[1:]
}

var Settings = settings{1, false}
var Hp = make(map[string]int)

type Shell struct {
	Value int
}

func (s Shell) String() string {
	if s.Value == 0 {
		return "live"
	}
	return "blank"
}

type settings struct {
	Damage         int
	CuffedOpponent bool
}
