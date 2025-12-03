package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func getIntialItems() ([]todoitem, error) {
	todofile, err := os.Open("todo.csv")
	if err != nil {
		fmt.Println("Failed to read file")
		return []todoitem{}, err
	}
	defer todofile.Close()

	reader := csv.NewReader(todofile)

	todoitemsraw, err := reader.ReadAll()

	todoitems := []todoitem{}

	for _, todo := range todoitemsraw {
		id, err := strconv.Atoi(todo[0])
		if err != nil {
			fmt.Println("Invalid type")
			return []todoitem{}, err
		}
		done, err := strconv.ParseBool(todo[1])
		if err != nil {
			fmt.Println("Invalid type")
			return []todoitem{}, err
		}
		priority, err := strconv.Atoi(todo[3])
		if err != nil {
			fmt.Println("Invalid type")
			return []todoitem{}, err
		}
		todoitems = append(todoitems, todoitem{
			id:       id,
			done:     done,
			todo:     todo[2],
			priority: priority,
			notes:    todo[4],
		})
	}
	return todoitems, nil
}
