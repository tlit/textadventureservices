package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"textadventureservices/services/worldgen"
)

func main() {
	// Parse command line flags
	prompt := flag.String("prompt", "", "The prompt for world generation (e.g., 'A mysterious underwater city')")
	exits := flag.String("exits", "north,south,east,west", "Comma-separated list of exits (e.g., 'north,south')")
	output := flag.String("output", "generated_world.json", "Output file for the generated world")
	flag.Parse()

	if *prompt == "" {
		fmt.Println("Please provide a prompt using the -prompt flag")
		flag.Usage()
		os.Exit(1)
	}

	// Parse exits
	exitList := strings.Split(*exits, ",")
	for i, exit := range exitList {
		exitList[i] = strings.TrimSpace(exit)
	}

	// Create a new world with a random seed
	world, err := worldgen.NewWorld(42) // You can change this seed for different variations
	if err != nil {
		fmt.Printf("Failed to create world: %v\n", err)
		os.Exit(1)
	}

	// Generate a room with the given prompt
	room, err := worldgen.GenerateRoom(*prompt, exitList)
	if err != nil {
		fmt.Printf("Failed to generate room: %v\n", err)
		os.Exit(1)
	}

	// Add the room to the world
	world.AddRoom(room)

	// Save the world to a file
	if err := world.Save(*output); err != nil {
		fmt.Printf("Failed to save world: %v\n", err)
		os.Exit(1)
	}

	// Print the generated world for preview
	data, err := json.MarshalIndent(world, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal world: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGenerated World Preview:\n%s\n", string(data))
	fmt.Printf("\nWorld saved to: %s\n", *output)
}
