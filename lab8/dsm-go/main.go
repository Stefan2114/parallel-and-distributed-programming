package main

import (
	"bufio"
	"dsm-go/config"
	"dsm-go/dsm"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	idPtr := flag.Int("id", 0, "Node ID (0, 1, or 2)")
	flag.Parse()
	myID := *idPtr

	me := config.GetProcess(myID)
	if me.Port == "" {
		log.Fatalf("Unknown Node ID: %d", myID)
	}

	fmt.Printf("--- DSM Node %d (Port %s) ---\n", myID, me.Port)

	onUpdate := func(varID, oldVal, newVal int) {
		fmt.Printf("\n>>> CALLBACK: Variable %d changed: %d -> %d\n> ", varID, oldVal, newVal)
	}

	dsmSys := dsm.NewDSM(myID, me.Port, onUpdate)

	time.Sleep(1 * time.Second)

	scanner := bufio.NewScanner(os.Stdin)
	printHelp()

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		switch cmd {
		case "read":
			varID := parseInt(parts, 1)
			val, ok := dsmSys.Get(varID)
			if ok {
				fmt.Printf("Variable %d = %d\n", varID, val)
			} else {
				fmt.Printf("Not subscribed to Variable %d\n", varID)
			}

		case "write":
			varID := parseInt(parts, 1)
			val := parseInt(parts, 2)
			err := dsmSys.Write(varID, val)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Write requested.")
			}

		case "cas":
			varID := parseInt(parts, 1)
			oldVal := parseInt(parts, 2)
			newVal := parseInt(parts, 3)
			success, err := dsmSys.CompareAndExchange(varID, oldVal, newVal)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				if success {
					fmt.Println("CAS Successful!")
				} else {
					fmt.Println("CAS Failed (Value did not match).")
				}
			}

		case "help":
			printHelp()

		case "exit":
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}

func parseInt(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}
	var res int
	fmt.Sscanf(parts[index], "%d", &res)
	return res
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  read <varID>")
	fmt.Println("  write <varID> <val>")
	fmt.Println("  cas <varID> <oldVal> <newVal>")
	fmt.Println("  exit")
}
