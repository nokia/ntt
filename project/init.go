package project

import "github.com/subosito/gotenv"

func init() {
	gotenv.Load("ntt.env", "k3.env")
}
