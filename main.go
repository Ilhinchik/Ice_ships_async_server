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
	Result       bool   `json:"result"`
	Token        string `json:"token"`
}

type RequestBody struct {
	IcebreakerID int `json:"icebreaker_id"`
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

	// Генерация случайного результата
	result := rand.Float64() < 0.5 // 50% шанс на true или false

	// Успешный ответ в формате JSON
	successMessage := map[string]interface{}{
		"message": "Successful",
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
		// Задержка 5 секунд
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