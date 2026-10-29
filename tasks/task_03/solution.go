package main

import (
	"errors"
	"strconv"
)

var ErrZero = errors.New("zero is not allowed")

func fizzBuzz(n int) (string, error) {
	if n == 0 {
		return "", ErrZero
	}

	fizz := n%3 == 0
	buzz := n%5 == 0

	switch {
	case fizz && buzz:
		return "FizzBuzz", nil
	case fizz:
		return "Fizz", nil
	case buzz:
		return "Buzz", nil
	}
	return strconv.Itoa(n), nil
}
