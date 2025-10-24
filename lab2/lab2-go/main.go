package main

import (
	"fmt"
)

type Producer struct {
	value chan int
}

type Consumer struct {
	value  chan int
	result chan int
}

func (p *Producer) scalarProduct(vector1, vector2 []int) {
	if len(vector1) != len(vector2) {
		panic("vector length mismatch")
	}
	for i := 0; i < len(vector1); i++ {
		p.value <- vector1[i] * vector2[i]
	}
	close(p.value)
}

func (c *Consumer) consume() {
	sum := 0
	for value := range c.value {
		sum += value
	}
	c.result <- sum
}

func main() {
	vector1 := []int{1, 3, -2}
	vector2 := []int{4, -1, 5}
	valueChan := make(chan int)
	resultChan1 := make(chan int)
	resultChan2 := make(chan int)

	producer := Producer{valueChan}
	consumer1 := Consumer{valueChan, resultChan1}
	consumer2 := Consumer{valueChan, resultChan2}
	go producer.scalarProduct(vector1, vector2)
	go consumer1.consume()
	go consumer2.consume()
	sum1 := <-resultChan1
	sum2 := <-resultChan2
	fmt.Println("result: 1", sum1)
	fmt.Println("result: 2", sum2)
	fmt.Println("result: 3", sum1+sum2)

}
