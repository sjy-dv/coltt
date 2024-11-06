package main

import (
	"fmt"

	"github.com/sjy-dv/nnv/backup"
	"github.com/sjy-dv/nnv/backup/document"
	"github.com/sjy-dv/nnv/backup/query"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {

	db, _ := backup.Open("test-db")

	db.CreateCollection("test")

	d := document.NewDocument()
	type A struct {
		ID   string
		Name string
	}
	a := A{}
	a.ID = "aaa"
	a.Name = "name_aaa"
	d.Set("struct", a)
	db.InsertOne("test", d)

	data, _ := db.FindAll(query.NewQuery("test"))
	fmt.Println(data[0].Get("struct"))
	cc, _ := msgpack.Marshal(data[0].Get("struct"))
	a2 := A{}
	msgpack.Unmarshal(cc, &a2)
	fmt.Println(a2.ID, a2.Name)
}
