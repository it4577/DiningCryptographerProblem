package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"sync"
)

var wg sync.WaitGroup

// Objects for use in protocol

// Cryptographer
type Cryptographer struct {
	num int  // Position
	key bool // Coin Flip (Local)
	rec bool // Coin Flip (Recieved)
	pay bool // Payer
	msg bool // Broadcast Message
}

// Create functions performed by the Cryptographers

func (c *Cryptographer) FlipCoin (channel chan bool) {
	defer wg.Done()			      // End wait group (decrements counter)
	c.key = (rand.Int() % 2 == 0) // Determine coin flip
	channel <- c.key			  // Share flip result
	close(channel)				  // End communication
	// Print the coin flip for confirmation only (not seen by other processes)
	fmt.Println("Cryptographer " + strconv.Itoa(c.num) + " > : Coin flip was: " + strconv.FormatBool(c.key))
}

func (c *Cryptographer) CompCoin (channel chan bool) {
	defer wg.Done()			      // End wait group (decrements counter)
	c.rec = <-channel             // Store the result from other cryptographer
	c.msg = xor(c.key, c.rec)     // Compare and build broadcast message
	// Negate message if paying
	if c.pay { c.msg = xor(c.msg, true) }
}

// Create functions performed by the Observers

func Observer (cryptographers []Cryptographer) {
	fmt.Println()	              // Format only
	// Print the broadcast messages the observer sees
	for _, c := range cryptographers {
		wg.Done()			      // End wait group (decrements counter)
		fmt.Println("Observer > Cryptographer " + strconv.Itoa(c.num) + "'s broadcast message was: " + strconv.FormatBool(c.msg))
	}
	fmt.Println()	              // Format only
}

func Owner (cryptographers []Cryptographer) {
	defer wg.Done()			      // End wait group (decrements counter)
	var paid bool
	// Determine if a cryptographer has paid the bill
	paid = xor(cryptographers[0].msg, cryptographers[1].msg)
	paid = xor(paid, cryptographers[2].msg)
	// Print the conclusion
	if paid {
		fmt.Println("Owner > A Cryptographer has paid the bill.\n")
	} else {
		fmt.Println("Owner > The NSA must have paid the bill.\n")
	}
}

func CryptographerZero (cryptographers []Cryptographer) {
	defer wg.Done()			      // End wait group (decrements counter)
	var payer string
	// Compute who paid based on actual coin flips and broadcast messages
	if nxor(cryptographers[0].key, cryptographers[2].key) == cryptographers[0].msg {
		payer = "Cryptographer 1 Paid"
	} else if nxor(cryptographers[1].key, cryptographers[0].key) == cryptographers[1].msg {
		payer = "Cryptographer 2 Paid"
	} else if nxor(cryptographers[2].key, cryptographers[1].key) == cryptographers[2].msg {
		payer = "Cryptographer 3 Paid"
	} else {
		payer = "The NSA Paid"
	}
	// Print the conclusion
	fmt.Println("CryptographerZero > " + payer + "\n")
}

func CryptographerZeroDeterminesPayer (cryptographers []Cryptographer) {
	var payer int
	// Select a participant at random to pay
	payer = rand.Int() % 4
	// Print if the NSA is paying
	if payer == 3 {
		fmt.Println("\nCryptographerZero > Chooses no cryptographer as payer. NSA pays.\n")
		return
	}
	// Update cryptographer and print who was chosen to pay
	cryptographers[payer].pay = true
	fmt.Println("\nCryptographerZero > Chooses cryptographer " + strconv.Itoa(payer + 1) + " as payer.\n")
}

// Create functions to support protocol

// XOR operation
func xor(a bool, b bool) bool {
	return a != b
}

// XOR negation operation
func nxor(a bool, b bool) bool {
	return xor((a != b), true)
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// Create channels to share coin flip results
	coin0 := make(chan bool, 1)
	coin1 := make(chan bool, 1)
	coin2 := make(chan bool, 1)

	// Create the Cryptographers
	Cryptographers := []Cryptographer {
		{ num: 1, pay: false, },
		{ num: 2, pay: false, },
		{ num: 3, pay: false, },
	}

	// This is only known to the payer and CryptographerZero
	CryptographerZeroDeterminesPayer(Cryptographers) // Determine who pays

	// Cryptographers perform the coin flips
	wg.Add(3)
	go Cryptographers[0].FlipCoin(coin0)
	time.Sleep(time.Millisecond * 5)          // Ensures display order
	go Cryptographers[1].FlipCoin(coin1)
	time.Sleep(time.Millisecond * 5)          // Ensures display order
	go Cryptographers[2].FlipCoin(coin2)
	wg.Wait()

	// Cryptographers perform the coin comparisons for creating broadcast messages
	wg.Add(3)
	go Cryptographers[0].CompCoin(coin2)
	go Cryptographers[1].CompCoin(coin0)
	go Cryptographers[2].CompCoin(coin1)
	wg.Wait()

	// The Observer calls out the broadcast messages
	wg.Add(3)
	go Observer(Cryptographers)
	wg.Wait()

	// Owner determines from the broadcast messages which cryptographer paid
	wg.Add(1)
	go Owner(Cryptographers)
	wg.Wait()

	// CryptographerZero computes the actual payer
	wg.Add(1)
	go CryptographerZero(Cryptographers)
	wg.Wait()
}
