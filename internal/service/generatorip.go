package service

import (
	"fmt"
	"math/rand"
)

// GeneratePublicIPv4 случайныйх IP адрес
func GeneratePublicIPv4() string {
	for {
		octets := make([]int, 4)
		for i := range octets {
			octets[i] = rand.Intn(256)
		}

		if !isPrivate(octets) {
			return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
		}
	}
}

func isPrivate(octets []int) bool {
	// Проверка на 10.0.0.0/8
	if octets[0] == 10 {
		return true
	}

	// Проверка на 127.0.0.0/8 (loopback)
	if octets[0] == 127 {
		return true
	}

	// Проверка на 169.254.0.0/16 (link-local)
	if octets[0] == 169 && octets[1] == 254 {
		return true
	}

	// Проверка на 172.16.0.0/12
	if octets[0] == 172 && octets[1] >= 16 && octets[1] <= 31 {
		return true
	}

	// Проверка на 192.168.0.0/16
	if octets[0] == 192 && octets[1] == 168 {
		return true
	}

	return false
}
