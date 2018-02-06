package clfft

import (
	"fmt"
)

type NoPlatformsError struct {}

func (_ NoPlatformsError) Error() string {
    return "No OpenCL platforms found"
}

type NoDevicesError struct {
    PlatformName string
}

func (err NoDevicesError) Error() string {
    return fmt.Sprintf("No OpenCL devices found on platform '%s'", err.PlatformName)
}

type NotInitialisedError struct {}

func (_ NotInitialisedError) Error() string {
	return fmt.Sprintf("CLFourier is not initialised")
}

type AlreadyInitialisedError struct {}

func (_ AlreadyInitialisedError) Error() string {
	return fmt.Sprintf("CLFourier is already initialised")
}

type InvalidInputSizeError struct {
	Expected int
	Got int
}

func (err InvalidInputSizeError) Error() string {
	return fmt.Sprintf("Expected input (sample data / frequencies) of size %d, but got %d", err.Expected, err.Got)
}