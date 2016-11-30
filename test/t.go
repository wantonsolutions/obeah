package main

func main() {
	a := 1
	switch a {
	case 1:
		a++
		break
	case 2:
		a--
		break
	case 3:
		a = a * a
		break
	default:
		a = a
		break
	}
	a = 1
	for i := 0; i < 20; i++ {
		a += i
	}
	if a < 2 {
		a = 5
	} else if a > 6 {
		a = 6
	} else {
		a = 7
	}
}
