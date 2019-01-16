package main

func main() {
	a := App{}
	// You need to set your Username and Password here
	a.Initialize("root", "2627", "Contacts")

	a.Run(":8080")
}
