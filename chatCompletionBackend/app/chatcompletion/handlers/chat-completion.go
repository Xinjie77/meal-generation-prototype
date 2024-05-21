package app

import (
	"bytes"
	"chatcompletion/app/chatcompletion/model"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

const (
	apiURL = "https://api.openai.com/v1/chat/completions"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Meal struct {
	Day         string `json:"day"`
	MealType    string `json:"mealType"`
	Ingredients string `json:"ingredients"`
	Instructions string `json:"instructions"`
	Nutrition   string `json:"nutrition"`
}

type UserOnboardingData struct {
	Purpose      string    `json:"primaryPurpose"`
	Height               float64         `json:"height"` // in cm
	Weight               float64         `json:"weight"` // in kg
	Gender               string          `json:"gender"`
	DietPreference      map[string]bool `json:"dietPreference"`
	FitnessLevel	 string `json:"fitnessLevel"`
	Allergies           string `json:"allergies"`
	MealsPerDay         int    `json:"mealsPerDay"`
	SnacksPerDay         int    `json:"snacksPerDay"`
}

type ReturnData struct {
	Meals []Meal `json:"meal"`
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/get-meal-data", getMealDataHandler)
	mux.HandleFunc("/swap-meal", swapMealHandler)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"POST", "GET"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})

	handler := c.Handler(mux)
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func getMealDataHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key not set in environment variables", http.StatusInternalServerError)
		return
	}

	userOnboardingData, err := getUserOnboardingData(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting user onboarding data: %v", err), http.StatusInternalServerError)
		return
	}
	messageContent := formatPrompt(*userOnboardingData)
	fmt.Println("Message Content: ", messageContent)
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: messageContent,
		},
	}

	chatReq := chatRequest{
		Model:    "gpt-4o",
		Messages: messages,
	}

	maxRetries := 5

	var resp *http.Response
	var openAIResp model.OpenAIResponse
	var respData string
	var parsedMealData interface{}
	attempts := 0

	// Retry the request up to maxRetries times
	for ; attempts < maxRetries; attempts++ {
		client := &http.Client{}
		requestBody, err := json.Marshal(chatReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error marshalling request: %v", err), http.StatusInternalServerError)
			return
		}
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("Error making request, retrying (%d/%d): %v\n", attempts+1, maxRetries, err)
			continue
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&openAIResp)
		if err != nil {
			fmt.Printf("Error decoding response body, retrying (%d/%d): %v\n", attempts+1, maxRetries, err)
			continue
		}

		if len(openAIResp.Choices) == 0 || len(openAIResp.Choices[0].Message.Content) == 0 {
			fmt.Printf("No content found in response, retrying (%d/%d)\n", attempts+1, maxRetries)
			continue
		}

		respData = openAIResp.Choices[0].Message.Content
		fmt.Println("Response Data: ", respData)

		parsedMealData, err = handleMealData(respData)
		if err != nil {
			fmt.Printf("Error handling meal data, retrying (%d/%d): %v\n", attempts+1, maxRetries, err)
			chatReq.Messages = append(chatReq.Messages, Message{
				Role:    "assistant",
				Content: respData,
			})
			chatReq.Messages = append(chatReq.Messages, Message{
				Role:    "user",
				Content: "There is an error with your data: " + err.Error() + ". Please fix your data by formatting the JSON correctly and resend it.",
			})
			fmt.Println(chatReq.Messages)
			fmt.Println(attempts)
			continue
		}
		break
	}
	if attempts == maxRetries {
		http.Error(w, fmt.Sprintf("Error after %d attempts: %v", maxRetries, err), http.StatusInternalServerError)
		return
	}

	returnData := ReturnData{
		Meals: parsedMealData.([]Meal),
	}
	
	// Marshal the meals to JSON
	jsonData, err := json.Marshal(returnData)
	if err != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}
	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func getUserOnboardingData(r *http.Request) (*UserOnboardingData, error) {
	var userData UserOnboardingData
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		return nil, err
	}

	return &userData, nil
}

func formatPrompt(userOnboardingData UserOnboardingData) string {
	Preferences := []string{}
	for Preference, selected := range userOnboardingData.DietPreference {
		if selected {
			Preferences = append(Preferences, Preference)
		}
	}

	return fmt.Sprintf(`
Generate a meal plan with the user's details as follows:
Purpose: %s
Height: %g cm
Weight: %g kg
Gender: %s
Diet Preference: %s
Allergies: %s
Meals Per Day: %d
Snacks Per Day: %d
Approximate Calories intake per Day: %f

Provide daily meal plans from Monday to Sunday. Include details like ingredients, instructions, and nutrition information for each meal.

Respond in CSV format, with each row representing a meal entry. The columns should be as follows:
day of week,meal type,ingredients,instructions,nutrition
Do not include a header row. The first line should be the first meal entry.
Ensure all meals align with the user's diet preference and do not contain any allergens specified.
Ensure the response is valid CSV without any extraneous characters or formatting. Include no additional text in your response.
Split each column using semicolons (;)`,
		userOnboardingData.Purpose,
		userOnboardingData.Height,
		userOnboardingData.Weight,
		userOnboardingData.Gender,
		strings.Join(Preferences, ", "),
		userOnboardingData.Allergies,
		userOnboardingData.MealsPerDay,
		userOnboardingData.SnacksPerDay,
		CaloricIntake(userOnboardingData.Height, userOnboardingData.Weight, userOnboardingData.Gender, userOnboardingData.Purpose, userOnboardingData.FitnessLevel),
		)
}

func CaloricIntake(height, weight float64, gender string, purpose string, fitnessLevel string) float64 {
	var bmr float64

	// Trim and lowercase
	gender = strings.TrimSpace(strings.ToLower(gender))
	fitnessLevel = strings.TrimSpace(strings.ToLower(fitnessLevel))

	// Calculate Basal Metabolic Rate (BMR) using the Harris-Benedict equation
	if gender == "male" {
		bmr = 88.362 + (13.397 * weight) + (4.799 * height) - (5.677 * float64(30))
	} else if gender == "female" {
		bmr = 447.593 + (9.247 * weight) + (3.098 * height) - (4.330 * float64(30))
	}

	// Adjust BMR based on activity level
	var activityMultiplier float64
	switch fitnessLevel {
	case "Beginner":
		activityMultiplier = 1.2
	case "I do Sport from time to time":
		activityMultiplier = 1.43
	case "I do sport regularly":
		activityMultiplier = 1.67
	case "advanced":
		activityMultiplier = 1.9
	default:
		activityMultiplier = 1.2
	}
	// Calculate Total Daily Energy Expenditure (TDEE)
	tdee := bmr * activityMultiplier

	// Adjust TDEE based on the goals
	if purpose == "Gain Muscle" {
		tdee += 500
	} else if purpose == "Lose Weight" {
		tdee -= 500
	}
	return tdee
}

func handleMealData(csvString string) ([]Meal, error) {
	var meals []Meal
	// Validate and correct the CSV string line by line
	lines := strings.Split(csvString, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		commaCount := strings.Count(line, ",")
		if commaCount != 4 {
			line = fixCommaCount(line, commaCount)
		}
		lines[i] = line
	}

	correctedCSVString := strings.Join(lines, "\n")

	r := csv.NewReader(strings.NewReader(correctedCSVString))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	for _, record := range records {
		meal := Meal{
			Day:         record[0],
			MealType:    record[1],
			Ingredients: record[2],
			Instructions: record[3],
			Nutrition:   record[4],
		}
		meals = append(meals, meal)
	}

	return meals, nil
}

func fixCommaCount(line string, commaCount int) string {
	fmt.Println("Fixing comma count for line: ", line)
	if commaCount < 4 {
		missingCommas := 4 - commaCount
		for missingCommas > 0 {
			index := strings.Index(line, ",,")
			if index != -1 {
				line = line[:index+1] + "," + line[index+1:]
			} else {
				line += ","
			}
			missingCommas--
		}
	} else if commaCount > 4 {
		extraCommas := commaCount - 4
		for extraCommas > 0 {
			index := strings.Index(line, ",,")
			if index != -1 {
				line = line[:index] + line[index+1:]
			} else {
				index = strings.LastIndex(line, ",")
				line = line[:index] + line[index+1:]
			}
			extraCommas--
		}
	}
	return line
}

func swapMealHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key not set in environment variables", http.StatusInternalServerError)
		return
	}

	var requestData struct {
		Meal           Meal `json:"meal"`
		OtherMeals     []string `json:"otherMeals"`
		DietPreference      map[string]bool `json:"dietPreference"`
		Allergies           string `json:"allergies"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusInternalServerError)
		return
	}
	otherMealsStr := strings.Join(requestData.OtherMeals, ", ")
	Preferences := []string{}
	for Preference, selected := range requestData.DietPreference {
		if selected {
			Preferences = append(Preferences, Preference)
		}
	}

	messageContent := fmt.Sprintf(`
I need an alternative meal option for the following meal:
{
	Day: %s
	Meal Type: %s
	Ingredients: %s
	Instructions: %s
	Nutrition: %s
}
Diet Preference: %s
Allergies: %s
The new meal should not be any of the following: %s.
The new meal should be suitable for the user's diet preference and allergies
.
Provide an alternative meal with the same format.
Respond in CSV format with the columns as follows:
day of week,meal type,ingredients,instructions,nutrition
Do not include a header row. The first line should be the new meal entry.`,
		requestData.Meal.Day, requestData.Meal.MealType, requestData.Meal.Ingredients, requestData.Meal.Instructions, requestData.Meal.Nutrition,
		strings.Join(Preferences, ", "),requestData.Allergies,otherMealsStr)

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: messageContent,
		},
	}

	chatReq := chatRequest{
		Model:    "gpt-4o",
		Messages: messages,
	}

	client := &http.Client{}
	requestBody, err := json.Marshal(chatReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling request: %v", err), http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var openAIResp model.OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding response body: %v", err), http.StatusInternalServerError)
		return
	}

	if len(openAIResp.Choices) == 0 || len(openAIResp.Choices[0].Message.Content) == 0 {
		http.Error(w, "No content found in response", http.StatusInternalServerError)
		return
	}

	respData := openAIResp.Choices[0].Message.Content
	fmt.Println("Response Data: ", respData) // Log the response data for debugging

	// Validate and correct the CSV string line by line (note that GPT should only return one line for a swap meal request)
	lines := strings.Split(respData, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		commaCount := strings.Count(line, ",")
		if commaCount != 5 {
			// Find two adjacent commas and add/remove from there
			line = fixCommaCount(line, commaCount)
		}
		lines[i] = line
	}

	correctedCSVString := strings.Join(lines, "\n")

	// Read the corrected CSV string
	reader := csv.NewReader(strings.NewReader(correctedCSVString))
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading CSV: %v", err), http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		http.Error(w, "No records found in CSV response", http.StatusInternalServerError)
		return
	}

	record := records[0]
	newMeal:= Meal{
		Day:         record[0],
		MealType:    record[1],
		Ingredients: record[2],
		Instructions: record[3],
		Nutrition:   record[4],
	}

	response := map[string]interface{}{
		"oldMealIngredients": requestData.Meal.Ingredients,
		"newMeal":     newMeal,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling response: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Sending response: ", string(jsonResponse)) // Log the JSON response

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
