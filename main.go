package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	IcebreakerToken = "icebreaker_secret_token"
	resultURL       = "http://127.0.0.1:8000/api/async_result/"
)

type IcebreakerResult struct {
	IcebreakerID int    `json:"icebreaker_id"`
	Result       int   `json:"result"`
	Token        string `json:"token"`
}

type RequestBody struct {
	IcebreakerID int `json:"icebreaker_id"`
}

// Функция для преобразования числа в соответствующее слово
func getResultWord(result int) string {
	switch result {
	case 0:
		return "Ошибка"
	case 1:
		return "В работе"
	case 2:
		return "Успех"
	case 3:
		return "Потеря"
	default:
		return "Неизвестный статус"
	}
}

func main() {
	http.HandleFunc("/api/async_calc/", handleProcess)
	fmt.Println("Server running at port :8100")
	http.ListenAndServe(":8100", nil)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "The method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody RequestBody
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		fmt.Println("Error decoding JSON:", err)
		return
	}

	icebreakerID := requestBody.IcebreakerID
	fmt.Println("Icebreaker ID:", icebreakerID)

	randomValue := rand.Float64()
	var result int
	if randomValue < 0.5 {
		result  = 2
	} else {
		result = 3
	}

	// // Генерация случайного результата
	// result := rand.Float64() < 0.5 // 50% шанс на true или false

	// Успешный ответ в формате JSON
	successMessage := map[string]interface{}{
		"message": "Successful",
		"result": getResultWord(result),
		"data": IcebreakerResult{
			IcebreakerID: icebreakerID,
			Result:       result,
			Token:        IcebreakerToken,
		},
	}

	jsonResponse, err := json.Marshal(successMessage)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("Error encoding JSON:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	go func() {
		// Задержка 10 секунд
		delay := 10
		time.Sleep(time.Duration(delay) * time.Second)

		// Отправка результата на Django-сервер
		result := IcebreakerResult{
			IcebreakerID: icebreakerID,
			Result:       result,
			Token:        IcebreakerToken,
		}

		fmt.Println("Sending result:", result)
		jsonValue, err := json.Marshal(result)
		if err != nil {
			fmt.Println("Error during JSON marshalization:", err)
			return
		}

		req, err := http.NewRequest(http.MethodPut, resultURL, bytes.NewBuffer(jsonValue))
		if err != nil {
			fmt.Println("Error when creating an update request:", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error when sending an update request:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Response from the update server:", resp.Status)
	}()
}