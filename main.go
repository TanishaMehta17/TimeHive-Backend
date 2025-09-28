package main

import (
	"fmt"

	"github.com/TanishaMehta17/TimeHive-Backend/config"
)

func main() {
	db := config.ConnectDB()

	if db != nil {
		fmt.Println(" Connected to PostgreSQL successfully!")
	} else {
		fmt.Println(" Failed to connect to PostgreSQL.")
	}

}
