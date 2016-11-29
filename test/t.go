package main

func main() {
	a := 1
	switch a {
	case 1:
		break
	case 2:
		break
	case 3:
		break
	default:
		break
	}
	a = 1
	for i := 0; i < 20; i++ {
		a += i
	}
	if a < 2 {
		a = 5
	} else {
		a = 6
	}
}
