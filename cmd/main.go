package main

import (
	DB "root/internal/database"
	se "root/internal"
)

func main() {
	DB.InitDB()
	se.ServerRunner()
}