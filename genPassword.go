package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func main() {
	passwd := "salt#password"
	hash := md5.Sum([]byte(passwd))
	k := hash[:]
	res := make([]byte, 100)
	n := hex.Encode(res, k)
	res = res[:n]
	fmt.Println(n, string(res[:n]))
}
