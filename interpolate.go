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

// Point represents a decoded (x, y) coordinate for the polynomial.
type Point struct {
	X *big.Int
	Y *big.Int
}

// lagrangeInterpolateAtZero calculates the polynomial's value at x=0 (the secret).
// The formula for the secret C is: C = f(0) = Σ [y_i * L_i(0)]
// where L_i(0) is the Lagrange basis polynomial: L_i(0) = Π [x_j / (x_j - x_i)] for j ≠ i.
func lagrangeInterpolateAtZero(points []Point) *big.Int {
	secret := new(big.Rat).SetInt64(0) // Use rational numbers for precision
	k := len(points)

	// Convert points to big.Rat for calculations
	xRats := make([]*big.Rat, k)
	yRats := make([]*big.Rat, k)
	for i := 0; i < k; i++ {
		xRats[i] = new(big.Rat).SetInt(points[i].X)
		yRats[i] = new(big.Rat).SetInt(points[i].Y)
	}

	// Iterate through each point to calculate its term in the Lagrange sum
	for i := 0; i < k; i++ {
		// Calculate L_i(0)
		numerator := new(big.Rat).SetInt64(1)
		denominator := new(big.Rat).SetInt64(1)

		for j := 0; j < k; j++ {
			if i == j {
				continue
			}
			// Numerator term is x_j
			numerator.Mul(numerator, xRats[j])

			// Denominator term is (x_i - x_j)
			diff := new(big.Rat).Sub(xRats[i], xRats[j])
			denominator.Mul(denominator, diff)
		}

		// Calculate the full term: y_i * (numerator / denominator)
		lagrangeBasis := new(big.Rat).Quo(numerator, denominator)
		term := new(big.Rat).Mul(yRats[i], lagrangeBasis)

		// Add to the total secret
		secret.Add(secret, term)
	}

	// The final secret must be an integer.
	if !secret.IsInt() {
		log.Fatal("Error: The calculated secret is not an integer. Check the input points.")
	}

	return secret.Num() // Return the numerator, which is the integer value
}

func main() {
	// --- Checkpoint 1: Read the Test Case (Input) ---
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

	// Extract 'k' from the 'keys' object
	var keys struct {
		K int `json:"k"`
	}
	if err := json.Unmarshal(rawData["keys"], &keys); err != nil {
		log.Fatalf("Error parsing 'keys' object: %v", err)
	}
	k := keys.K

	// --- Checkpoint 2: Decode the Y Values ---
	var points []Point
	for key, rawValue := range rawData {
		if key == "keys" {
			continue
		}

		// Parse x-coordinate (the key)
		xVal, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			log.Printf("Warning: could not parse key '%s' as an integer. Skipping.", key)
			continue
		}

		// Parse the value object { "base": "b", "value": "v" }
		var val struct {
			Base  string `json:"base"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(rawValue, &val); err != nil {
			log.Fatalf("Error parsing point data for key '%s': %v", key, err)
		}

		// Convert base string to integer
		base, err := strconv.Atoi(val.Base)
		if err != nil {
			log.Fatalf("Error converting base '%s' to integer: %v", val.Base, err)
		}

		// Decode the y-value from the given base using big.Int
		yVal, success := new(big.Int).SetString(val.Value, base)
		if !success {
			log.Fatalf("Error decoding value '%s' with base %d", val.Value, base)
		}

		points = append(points, Point{X: big.NewInt(xVal), Y: yVal})
	}

	// Sort points by X value to ensure consistent order
	sort.Slice(points, func(i, j int) bool {
		return points[i].X.Cmp(points[j].X) < 0
	})

	// We only need 'k' points to solve for the polynomial
	if len(points) < k {
		log.Fatalf("Error: Not enough points in JSON (%d) to meet requirement k=%d", len(points), k)
	}
	pointsToUse := points[:k]

	// --- Checkpoint 3: Find the Secret (C) ---
	secret := lagrangeInterpolateAtZero(pointsToUse)

	fmt.Println("Successfully decoded points and calculated the secret.")
	fmt.Println("-----------------------------------------------------")
	fmt.Printf("Secret (C): %s\n", secret.String())
	fmt.Println("-----------------------------------------------------")
}
