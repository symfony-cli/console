package console

import (
	"os"

	"github.com/symfony-cli/terminal"
)

func checkWSL() {
	if fi, err := os.Stat("/proc/version"); fi == nil || err != nil {
		return
	}

	ui := terminal.SymfonyStyle(terminal.Stdout, terminal.Stdin)
	ui.Error("Wrong binary for WSL")
	terminal.Println(`You are trying to run the Windows version of the Symfony CLI on WSL (Linux).
You must use the Linux version to use the Symfony CLI on WSL.

Download it at <href=https://symfony.com/download>https://symfony.com/download</>
`)
	os.Exit(1)
}
