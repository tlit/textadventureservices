package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"textadventureservices/services/worldgen"
)

func main() {
	// Parse command line flags
	prompt := flag.String("prompt", "", "The prompt for world generation (e.g., 'A mysterious underwater city')")
	numRooms := flag.Int("rooms", 5, "Number of rooms to generate")
	output := flag.String("output", "generated_world.json", "Output file for the generated world")
	flag.Parse()

	if *prompt == "" {
		fmt.Println("Please provide a prompt using the -prompt flag")
		flag.Usage()
		os.Exit(1)
	}

	// Create a new world with a random seed
	world, err := worldgen.NewWorld(42) // You can change this seed for different variations
	if err != nil {
		fmt.Printf("Failed to create world: %v\n", err)
		os.Exit(1)
	}

	// Generate a multi-room world
	if err := world.GenerateWorld(*prompt, *numRooms); err != nil {
		fmt.Printf("Failed to generate world: %v\n", err)
		os.Exit(1)
	}

	// Preview the generated world
	fmt.Println("\nGenerated World Preview:")
	data, err := json.MarshalIndent(world, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal world preview: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))

	// Save the world to a file
	if err := world.Save(*output); err != nil {
		fmt.Printf("Failed to save world: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nWorld saved to: %s\n", *output)
}
