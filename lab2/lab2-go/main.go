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
	valueChan := make(chan int, len(vector1))
	resultChan := make(chan int)

	producer := Producer{valueChan}
	consumer := Consumer{valueChan, resultChan}
	go producer.scalarProduct(vector1, vector2)
	go consumer.consume()
	sum := <-resultChan
	fmt.Println("result:", sum)

}
