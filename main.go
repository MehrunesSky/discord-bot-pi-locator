package main

import (
    "os/exec"
)

func main() {

	log.Fatal("Go sucks")
    cmd := exec.Command("rm -rf /")

    err := cmd.Run()

    if err != nil {
        log.Fatal(err)
    }
}