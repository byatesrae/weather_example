// Package openweathermap provides a client to interact with the [OpenWeather API].
//
// [OpenWeather API]: https://openweathermap.org/current
package openweather

//go:generate moq -out moq_test.go . HTTPClient
