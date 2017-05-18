package mirc

import "time"

// GetTime returns the current time as a string
func GetTime() string {
	return time.Now().String()
}

// Retry wrapped function specified times
func Retry(retries int, function func() error) error {
	var err error
	for i := 0; i < retries; i++ {
		err = function()
		if err == nil {
			return err
		}
	}
	return err
}
