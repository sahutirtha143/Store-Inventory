package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

type InventoryItem struct {
	Date         string
	Item         string
	Quantity     float64
	PricePerUnit float64
	// TotalPrice        float64
	TotalPriceBuying  float64
	TotalPriceSelling float64
	TotalPriceAllItem float64
}

var (
	inventory        []InventoryItem
	sellingInventory []InventoryItem // To store selling inventory
	mu               sync.Mutex
)

func main() {
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/add-item", addItemHandler)
	http.HandleFunc("/add-item-selling", addItemHandlerSelling)

	fmt.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.ParseFiles("index.html"))

	// Calculate the TotalPriceAllItem
	totalPriceAllItem := calculateTotalPriceAllItems()
	totalPriceAllSellingItems := calculateTotalPriceAllSellingItems()
	// Calculate the ProfitLoss:
	totalProfitLoss := calculateProfitLoss()

	// Pass the total price along with inventory data
	data := struct {
		Inventory            []InventoryItem
		SellingInventory     []InventoryItem
		TotalPriceAllItems   float64
		TotalPriceAllSelling float64
		TotalProfitLoss      float64
	}{
		Inventory:            inventory,
		SellingInventory:     sellingInventory,
		TotalPriceAllItems:   totalPriceAllItem,
		TotalPriceAllSelling: totalPriceAllSellingItems,
		TotalProfitLoss:      totalProfitLoss,
	}

	t.Execute(w, data)
}

func addItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		// Create a new inventory item from form data
		item := InventoryItem{
			Date:         r.FormValue("date"),
			Item:         r.FormValue("item"),
			Quantity:     parsePrice(r.FormValue("quantity")),
			PricePerUnit: parsePrice(r.FormValue("price")),
		}

		// Calculate total price for the item
		item.TotalPriceBuying = item.Quantity * item.PricePerUnit

		mu.Lock()
		// Add the item to the inventory
		inventory = append(inventory, item)
		mu.Unlock()

		// Redirect back to the main page to show the updated inventory
		http.Redirect(w, r, "/#buy", http.StatusSeeOther)
		// fmt.Println("In", inventory)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func addItemHandlerSelling(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		// Create a new inventory item from form data
		item := InventoryItem{
			Date:         r.FormValue("date"),
			Item:         r.FormValue("item"),
			Quantity:     parsePrice(r.FormValue("quantity")),
			PricePerUnit: parsePrice(r.FormValue("price")),
		}
		// Calculate total price for the item
		item.TotalPriceSelling = item.Quantity * item.PricePerUnit

		mu.Lock()
		// Add the item to the inventory
		sellingInventory = append(sellingInventory, item)
		mu.Unlock()

		// Redirect back to the main page to show the updated inventory
		http.Redirect(w, r, "/#sell", http.StatusSeeOther)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Function to parse the price
func parsePrice(value string) float64 {
	var price float64
	fmt.Sscanf(value, "%f", &price)
	return price
}

// Function to calculate total price for all items in the buying inventory
func calculateTotalPriceAllItems() float64 {
	total := 0.0
	for _, item := range inventory {
		total += item.TotalPriceBuying
	}
	return total
}

// Function to calculate total price for all items in the selling inventory
func calculateTotalPriceAllSellingItems() float64 {
	total := 0.0
	for _, item := range sellingInventory {
		total += item.TotalPriceSelling
	}
	return total
}

// Function to calculate total Profit and Loss:

func calculateProfitLoss() float64 {

	// Calculate the total price for all bought items
	totalBuying := calculateTotalPriceAllItems()

	// Calculate the total price for all sold items
	totalSelling := calculateTotalPriceAllSellingItems()

	// Profit or Loss is the difference between selling and buying
	profitLoss := totalSelling - totalBuying

	return profitLoss
}
