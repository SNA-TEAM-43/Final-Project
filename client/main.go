package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func readLine(scanner *bufio.Scanner, text string) string {
	fmt.Print(text)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func printMenu(url string) {
	fmt.Println()
	fmt.Println("=== Load Client ===")
	fmt.Printf("Server: %s\n\n", url)
	fmt.Println("1) CPU load")
	fmt.Println("2) DB write")
	fmt.Println("3) DB read")
	fmt.Println("4) DB cleanup")
	fmt.Println("5) STRESS (all modes parallel)")
	fmt.Println("6) KILL (max load)")
	fmt.Println("7) CHAOS (inject DB failures for 30s)")
	fmt.Println("8) HEAL  (cancel chaos immediately)")
	fmt.Println("0) Exit")
	fmt.Println()
}

func main() {
	godotenv.Load()

	baseURL := "http://localhost:8080"
	if env := os.Getenv("SERVER_URL"); env != "" {
		baseURL = env
	}
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	scanner := bufio.NewScanner(os.Stdin)
	client := &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 10 * time.Minute},
	}

	for {
		printMenu(client.BaseURL)
		choice := readLine(scanner, "Choice: ")
		switch choice {
		case "1":
			if err := runCPU(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "2":
			if err := runDBWrite(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "3":
			if err := runDBRead(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "4":
			if err := runCleanup(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "5":
			if err := runStress(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "6":
			confirm := readLine(scanner, "KILL will load server for several minutes. Continue? [y/N]: ")
			if err := runKill(client, confirm); err != nil {
				fmt.Println("Error: ", err)
			}
		case "7":
			if err := runChaos(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "8":
			if err := runChaosHeal(client); err != nil {
				fmt.Println("Error: ", err)
			}
		case "0":
			return
		default:
			fmt.Println("Unknown term")
		}
	}
}
