package main

import (
	DB "root/backend/database"
	se "root/backend"
)

func main() {
	DB.InitDB()
	se.ServerRunner()


}