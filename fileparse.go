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
		priority, err := strconv.Atoi(todo[2])
		if err != nil {
			fmt.Println("Invalid type")
			return []todoitem{}, err
		}
		todoitems = append(todoitems, todoitem{
			id:       id,
			done:     done,
			todo:     todo[3],
			priority: priority,
		})
	}
	return todoitems, nil
}

func saveTasksToCSV(items []todoitem) error {
	file, err := os.Create("todo.csv")
	if err != nil {
		fmt.Println("Failed to create CSV file")
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, item := range items {
		record := []string{
			strconv.Itoa(item.id),
			strconv.FormatBool(item.done),
			strconv.Itoa(item.priority),
			item.todo,
		}
		if err := writer.Write(record); err != nil {
			fmt.Println("Failed to write record")
			return err
		}
	}

	return nil
}
