package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	initialAmounts := map[dish]int{
		"dish(chopitos)":            10,
		"dish(chorizo)":             5,
		"dish(croquetas)":           5,
		"dish(patatas bravas)":      7,
		"dish(pimientos de padron)": 5,
	}

	consumers := []string{"Alice", "Bob", "Charlie", "Dave"}
	bar := &bar{
		incomingFromKitchen:    make(chan dish, 100),
		incomingFromConsumers:  make(chan dish, 100),
		quantityRemainingTapas: initialAmounts,
	}
	chef := &chef{pace: 1, bar: bar}
	go chefWorker(chef)
	go barWorker(bar)
	// Create a consumer for each consumer in consumers
	for _, name := range consumers {
		c := &consumer{
			name: name,
		}
		go consumeDish(c, bar)
	}

	// Anonymous function that logs a kill signal then exits
	func() {
		killSignal := make(chan os.Signal, 1)
		signal.Notify(killSignal, syscall.SIGINT, syscall.SIGTERM)
		signal := <-killSignal
		fmt.Println("Received kill signal:", signal)
		os.Exit(0)
	}()

}

// func consumer(c *consumer) {
// 	ticker := time.NewTicker(time.Second * c.minPace)
// 	defer ticker.Stop()
// }

func barWorker(bar *bar) {
	for {
		select {
		case dish := <-bar.incomingFromKitchen:
			bar.quantityRemainingTapas[dish]++
			fmt.Println("Dish (", dish, ") has been produced by a chef:", bar.quantityRemainingTapas)
		case dish := <-bar.incomingFromConsumers:
			if bar.quantityRemainingTapas[dish] > 0 {
				bar.quantityRemainingTapas[dish]--
				fmt.Println("Dish (", dish, ") has been consumed by a consumer")
			}
			if bar.quantityRemainingTapas[dish] <= 0 {
				noRemainingTapas := true
				for _, quantity := range bar.quantityRemainingTapas {
					if quantity > 0 {
						noRemainingTapas = false
						break
					}
				}
				if noRemainingTapas {
					fmt.Println("No more tapas in bar:", bar.quantityRemainingTapas)
					os.Exit(0)
				}
			}
		}
	}
}

func chefWorker(chef *chef) {
	duration := time.Duration(chef.pace) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Create a random tapas
			tapas := getRandomTapas(chef.bar, false)
			produceDish(chef.bar, tapas)
		}
	}
}

func getRandomTapas(b *bar, isConsumer bool) dish { // Generate a random tapas from the initialAmounts map
	tapasList := []dish{}
	for d := range b.quantityRemainingTapas {
		if b.quantityRemainingTapas[d] == 0 && isConsumer {
			continue
		}
		tapasList = append(tapasList, d)
	}

	randIndex := rand.Intn(len(tapasList))
	return dish(tapasList[randIndex])
}

// function that operates on a bar and inserts a dish into the bartop chan
// as well as increment the quantityRemainingTapas atomically
func produceDish(b *bar, d dish) {
	b.quantityRemainingTapas[d]++
	b.incomingFromKitchen <- d
}

func consumeDish(c *consumer, b *bar) {
	for {
		// Create a random dish
		randomDish := getRandomTapas(b, true)
		b.incomingFromConsumers <- randomDish
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		duration := time.Duration(r.Intn(1000)) * time.Millisecond
		time.Sleep(duration)
	}
}

// Data types
type bar struct {
	incomingFromKitchen    chan dish
	incomingFromConsumers  chan dish
	quantityRemainingTapas map[dish]int
}

type dish string

type chef struct {
	pace int
	bar  *bar
}

type consumer struct {
	name string
}
