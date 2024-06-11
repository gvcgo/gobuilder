package main

import (
	"fmt"

	"github.com/gvcgo/gobuilder/internal/builder"
)

func main() {
	imgName := builder.FindXgoDockerImage()
	fmt.Println(imgName)
}
