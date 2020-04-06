package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	f, _ := os.OpenFile("./data.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	w := csv.NewWriter(f)
	w.Write([]string{"id", "password", "nickname", "pic_path"})
	cols := make([]string, 4)
	cols[1] = "7542570b4fbb8cbac314d3df00bf834e"
	cols[2] = "apple"
	cols[3] = "/bc.png"
	for i := 0; i < 10000000; i++ {
		cols[0] = fmt.Sprintf("U%d", i)
		w.Write(cols)
	}
	w.Flush()
	f.Close()
}
