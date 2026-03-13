package main

import (
    "goquickactu/routes"
)

func main() {
    r := routes.SetupRouter()
    r.Run(":8080")
}