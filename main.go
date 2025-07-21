package main

import (
	parser "benparse/parser"
	"fmt"
	"os"
	"reflect"
)

func main() {
	filePath := "/home/kcx/Downloads/ubuntu-25.04-desktop-amd64.iso.torrent"

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Parse the content
	val, err := parser.Parse(string(content))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	printMapKeys(val)

}

func printMapKeys(input any) {
	val := reflect.ValueOf(input)
	typ := val.Type()

	if val.Kind() != reflect.Map {
		fmt.Println("Not a map")
		return
	}

	fmt.Printf("Map of type: %s\n", typ)
	fmt.Println("Keys:")

	for _, key := range val.MapKeys() {
		fmt.Println(key.Interface())
	}
}
