package utils

import (
	"errors"
)

func TrySend[T any](value T, channel chan T) error {
	select {
	case channel <- value:
		return nil
	default:
		return errors.New("could not send to this channel, is it full?")
	}
}

func TryReceive[T any](channel chan T) (T, error) {
	var value T
	select {
	case value = <-channel:
		return value, nil
	default:
		return value, errors.New("could not receive")
	}
}
