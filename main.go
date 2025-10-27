package main

import (
	"fmt"

	"github.com/PokeForum/PokeForum/cmd"
	_const "github.com/PokeForum/PokeForum/internal/const"
)

func main() {
	fmt.Printf("PokeForum Version: %s (Hash: %s)", _const.Version, _const.GitCommit)
	cmd.Execute()
}
