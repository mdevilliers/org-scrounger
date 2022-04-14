package util 

func NewSet[T comparable]() Set[T] {
    return Set[T]{}
}

type Set[T comparable]  map[T]int

func(s Set[T]) Add(c T) {
	_, found := s[c]
	if found {
		s[c] = s[c] + 1
	} else{
		s[c] = 1
	}
}


