package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomPerc1(t *testing.T) {
	trues := 0
	falses := 0
	for i := 0; i < 10000; i++ {
		result := randomPercRange(0, 10, fmt.Sprintf("t%d", i), 1)
		if result {
			trues = trues + 1
		} else {
			falses = falses + 1
		}
	}
	assert.Equal(t, 971, trues)
	assert.Equal(t, 9029, falses)
}

func TestRandomPerc1Invert(t *testing.T) {
	trues := 0
	falses := 0
	for i := 0; i < 10000; i++ {
		result := randomPercRange(10, 100, fmt.Sprintf("t%d", i), 1)
		if result {
			trues = trues + 1
		} else {
			falses = falses + 1
		}
	}
	assert.Equal(t, 971, falses)
	assert.Equal(t, 9029, trues)
}

func TestRandomPerc2(t *testing.T) {
	trues := ""
	falses := ""
	for i := 0; i < 10; i++ {
		item := fmt.Sprintf("t%d", i)
		result := randomPercRange(0, 50, item, 1)
		if result {
			trues = trues + item + ","
		} else {
			falses = falses + item + ","
		}
	}
	assert.Equal(t, "t0,t1,t4,t6,t7,t8,", trues)
	assert.Equal(t, "t2,t3,t5,t9,", falses)
}

func TestRandomPerc2NewSeed(t *testing.T) {
	trues := ""
	falses := ""
	for i := 0; i < 10; i++ {
		item := fmt.Sprintf("t%d", i)
		result := randomPercRange(0, 50, item, 2)
		if result {
			trues = trues + item + ","
		} else {
			falses = falses + item + ","
		}
	}
	assert.Equal(t, "t0,t1,t2,t3,t6,t8,t9,", trues)
	assert.Equal(t, "t4,t5,t7,", falses)
}

func TestRandomPerc3(t *testing.T) {
	trues := 0
	falses := 0
	for i := 0; i < 100000; i++ {
		result := randomPercRange(0, 25, fmt.Sprintf("t%d", i), 1111)
		if result {
			trues = trues + 1
		} else {
			falses = falses + 1
		}
	}
	assert.Equal(t, 24801, trues)
	assert.Equal(t, 75199, falses)
}
