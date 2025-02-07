package tools

import (
	"os"
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
