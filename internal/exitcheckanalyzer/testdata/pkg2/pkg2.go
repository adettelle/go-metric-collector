package main

import "os"

func main() {
	// формулируем ожидания: анализатор должен находить ошибку,
	// описанную в комментарии
	os.Exit(1) // want "expression os.Exit\\(\\) in main func in main package"
}
