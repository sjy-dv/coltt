package main

import (
	"log"

	"github.com/sjy-dv/coltt/pkg/inverted"
)

func main() {
	var idx *inverted.BitmapIndex
	idx = inverted.NewBitmapIndex()

	_ = idx.Add(00001, map[string]interface{}{"age": 22, "name": "A", "gender": true})
	_ = idx.Add(00002, map[string]interface{}{"age": 21, "name": "B", "gender": true})
	_ = idx.Add(00003, map[string]interface{}{"age": 20, "name": "C", "gender": true})
	_ = idx.Add(00003, map[string]interface{}{"age": 27, "name": "C", "gender": false})
	_ = idx.Add(00005, map[string]interface{}{"age": 25, "name": "C", "gender": true})
	_ = idx.Add(00006, map[string]interface{}{"age": 19, "name": "D", "gender": true})
	_ = idx.Add(00007, map[string]interface{}{"age": 30, "name": "E", "gender": true})

	onlyAgeFilter := inverted.NewFilter("age", inverted.OpGreaterThan, 50)
	resSingle, err := idx.SearchSingleFilter(onlyAgeFilter)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resSingle)

	onlyAgeFilterE := inverted.NewFilter("age", inverted.OpGreaterThanEqual, 22)
	res2Single, err := idx.SearchSingleFilter(onlyAgeFilterE)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res2Single)
	onlyNameFilter := inverted.NewFilter("name", inverted.OpEqual, "A")
	nameSingle, err := idx.SearchSingleFilter(onlyNameFilter)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(nameSingle)

	ageFilterM := inverted.NewFilter("age", inverted.OpGreaterThanEqual, 20)
	nameFilterM := inverted.NewFilter("name", inverted.OpEqual, "C")
	resMulti, err := idx.SearchMultiFilter([]*inverted.Filter{
		ageFilterM, nameFilterM,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resMulti)

	ageFilterM = inverted.NewFilter("age", inverted.OpGreaterThan, 10)
	nameFilterM = inverted.NewFilter("name", inverted.OpEqual, "C")
	// genderFilterM := inverted.NewFilter("gender", inverted.OpNotEqual, true)
	resMulti, err = idx.SearchMultiFilter([]*inverted.Filter{
		ageFilterM, nameFilterM,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resMulti, 111)
}
