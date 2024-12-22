package main

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"os"
	"time"
)

func main() {
	const numRecords = 100000
	var users []map[string]interface{}
	startDate := time.Date(1924, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2006, 12, 31, 0, 0, 0, 0, time.UTC)
	for i := 0; i < numRecords; i++ {
		user := map[string]interface{}{
			"username":   gofakeit.Username(),
			"firstname":  gofakeit.Name(),
			"age":        gofakeit.Number(18, 99),
			"password":   gofakeit.Password(true, false, false, false, false, 32),
			"birth_date": gofakeit.DateRange(startDate, endDate).Format("2006-01-02"),
			"gender":     gofakeit.Gender(),
		}
		users = append(users, user)
	}

	file, err := os.Create("data.json")
	if err != nil {
		fmt.Println("Ошибка при создании файла:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Устанавливаем отступ для читаемости
	if err := encoder.Encode(users); err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
		return
	}

	fmt.Println("Данные успешно сгенерированы и записаны в data.json!")
}
