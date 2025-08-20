package main

import (
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	insecure()
	authenticated()
	oauth2clientauthenticated()
	reportResource()
	deleteResource()
}
