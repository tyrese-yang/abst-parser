package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/tyrese/HDS/f4v"
)

func main() {
	fileName := os.Args[1]
	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer f.Close()
	abstByte, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err.Error())
		return
	}
	abst, err := f4v.ParseAbst(abstByte)
	if err != nil {
		log.Println(err.Error())
		return
	}
	abst.Print()
}
