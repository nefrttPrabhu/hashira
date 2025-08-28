package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"strconv"
)

type Point struct {
	X *big.Int
	Y *big.Int
}

func lagrangeInterpolateAtZero(points []Point) *big.Int {
	secret := new(big.Rat).SetInt64(0)
	k := len(points)

	xRats := make([]*big.Rat, k)
	yRats := make([]*big.Rat, k)
	for i := 0; i < k; i++ {
		xRats[i] = new(big.Rat).SetInt(points[i].X)
		yRats[i] = new(big.Rat).SetInt(points[i].Y)
	}

	for i := 0; i < k; i++ {
		numerator := new(big.Rat).SetInt64(1)
		denominator := new(big.Rat).SetInt64(1)

		for j := 0; j < k; j++ {
			if i == j {
				continue
			}
			numerator.Mul(numerator, xRats[j])
			diff := new(big.Rat).Sub(xRats[i], xRats[j])
			denominator.Mul(denominator, diff)
		}

		lagrangeBasis := new(big.Rat).Quo(numerator, denominator)
		term := new(big.Rat).Mul(yRats[i], lagrangeBasis)
		secret.Add(secret, term)
	}

	if !secret.IsInt() {
		log.Fatal("Error: The calculated secret is not an integer. Check the input points.")
	}

	return secret.Num()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run main.go <path_to_json_file>")
	}
	filePath := os.Args[1]

	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(file, &rawData); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	var keys struct {
		K int `json:"k"`
	}
	if err := json.Unmarshal(rawData["keys"], &keys); err != nil {
		log.Fatalf("Error parsing 'keys' object: %v", err)
	}
	k := keys.K

	var points []Point
	for key, rawValue := range rawData {
		if key == "keys" {
			continue
		}

		xVal, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			log.Printf("Warning: could not parse key '%s' as an integer. Skipping.", key)
			continue
		}

		var val struct {
			Base  string `json:"base"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(rawValue, &val); err != nil {
			log.Fatalf("Error parsing point data for key '%s': %v", key, err)
		}

		base, err := strconv.Atoi(val.Base)
		if err != nil {
			log.Fatalf("Error converting base '%s' to integer: %v", val.Base, err)
		}

		yVal, success := new(big.Int).SetString(val.Value, base)
		if !success {
			log.Fatalf("Error decoding value '%s' with base %d", val.Value, base)
		}

		points = append(points, Point{X: big.NewInt(xVal), Y: yVal})
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].X.Cmp(points[j].X) < 0
	})

	if len(points) < k {
		log.Fatalf("Error: Not enough points in JSON (%d) to meet requirement k=%d", len(points), k)
	}
	pointsToUse := points[:k]

	secret := lagrangeInterpolateAtZero(pointsToUse)

	fmt.Println("Successfully decoded points and calculated the secret.")
	fmt.Println("-----------------------------------------------------")
	fmt.Printf("Secret (C): %s\n", secret.String())
	fmt.Println("-----------------------------------------------------")
}
