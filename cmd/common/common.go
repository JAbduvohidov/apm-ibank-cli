package common

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ClearConsole() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func GetCommand() string {
	var cmd string
	_, err := fmt.Scan(&cmd)
	if err != nil {
		log.Fatalf("can't read input: %v", err)
	}
	return strings.TrimSpace(cmd)
}

func GetIntegerInput() int64 {
	var input int64
	_, err := fmt.Scan(&input)
	if err != nil {
		log.Fatalf("can't read input: %v", err)
	}
	return input
}

func GetStringInput() string {
	var input string
	_, err := fmt.Scan(&input)
	if err != nil {
		log.Fatalf("can't read input: %v", err)
	}
	return input
}