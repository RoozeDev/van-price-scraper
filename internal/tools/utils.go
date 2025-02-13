package tools

import (
	"os"
	"time"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
func ReadFileAsString(filePath string) string {

	contents, err := os.ReadFile(filePath)
	Check(err)

	return string(contents)
}

func ReadSecretAsString(secret string) string {

	var filePath = "/run/secrets/" + secret

	return ReadFileAsString(filePath)
}

func GetDatabaseName() string {
	var database = ReadSecretAsString("db_name")
	return database
}

func GetDatabaseConnectionURL() string {
	var username = ReadSecretAsString("db_username")
	var password = ReadSecretAsString("db_password")
	var database = GetDatabaseName()

	var uri = "mongodb://" + username + ":" + password + "@mongo:27017/" + database

	return uri

}

func GetAllMondaysInYear(year int) []time.Time {

	// Start with January 1st of the given year
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// Find the weekday of January 1st
	weekday := startDate.Weekday()

	// Calculate the offset to the first Monday
	offset := (8 - int(weekday)) % 7

	// Add the offset days to get the first Monday
	firstMonday := startDate.AddDate(0, 0, offset)

	loopDate := firstMonday

	var dates []time.Time

	for loopDate.Year() == year {
		dates = append(dates, loopDate)
		loopDate = loopDate.AddDate(0, 0, 7)
	}

	return dates

}
