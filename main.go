package main

import "fmt"

func main() {

	x := map[string]interface{}{
		"aa":   10,
		"bb":   0.12252115,
		"cxcx": []float32{0.1, 0.2},
	}

	a := make(map[string]string)
	for k, v := range x {
		a[k] = fmt.Sprintf("%v", v)
	}
	fmt.Println(a)
}
