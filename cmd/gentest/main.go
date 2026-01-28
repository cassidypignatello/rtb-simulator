package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cass/rtb-simulator/internal/generator"
	"github.com/cass/rtb-simulator/internal/generator/scenarios"
)

func main() {
	scenario := scenarios.NewMobileApp()
	gen := generator.New(scenario)

	fmt.Printf("Scenario: %s\n\n", gen.ScenarioName())

	req := gen.Generate()

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
}
