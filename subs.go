package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func loadConfig() AppConfig {

	jsonData, err := os.ReadFile(configFileName)
	if err != nil {
		fmt.Println("Error read config:", err)
	} else {
		var config AppConfig
		err = json.Unmarshal(jsonData, &config)
		if err != nil {
			fmt.Println("Error parse config:", err)
		} else {
			return config
		}
	}
	return AppConfig{
		IP:   "127.0.0.1",
		Port: "1234",
	}
}

func formatTimestamp(ts *int64) string {
	if ts == nil {
		return "—" // Или "Не запущено" / "В очереди"
	}

	// 1. Превращаем int64 в объект time.Time
	t := time.Unix(*ts, 0)

	// 2. Форматируем по эталонному шаблону Go: "02.01.2006 15:04:05"
	// (В Go вместо символов YYYY-MM-DD используется строго это эталонное время)
	return t.Format("02.01.2006 15:04:05")
}

func stringToIntSlice(input string) []int {
	// 1. Удаляем случайные пробелы по краям строки
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		return []int{}
	}

	// 2. Разбиваем строку по запятой
	strParts := strings.Split(cleaned, ",")
	intSlice := make([]int, 0, len(strParts))

	// 3. Конвертируем каждую часть в число
	for _, part := range strParts {
		// Удаляем пробелы вокруг конкретного числа (например, если было "1, 2 , 3")
		trimmedPart := strings.TrimSpace(part)

		num, err := strconv.Atoi(trimmedPart)
		if err != nil {
			fmt.Printf("Error conversion '%s' to int: %w", trimmedPart, err)
			return []int{}
		}
		intSlice = append(intSlice, num)
	}

	return intSlice
}

func intSliceToString(numbers []int) string {
	if len(numbers) == 0 {
		return ""
	}

	// 1. Создаем массив строк нужной длины
	strParts := make([]string, len(numbers))

	// 2. Превращаем каждое число в строку
	for i, num := range numbers {
		strParts[i] = strconv.Itoa(num)
	}

	// 3. Объединяем их через запятую
	return strings.Join(strParts, ",")
}

func parseGuidance(guidance GuidanceConfig, guidanceParams *GuidanceParamsPanel) {
	guidanceParams.DistilledInput.SetValue(guidance.DistilledGuidance)
	guidanceParams.TxtCfgInput.SetValue(guidance.TxtCfg)
	guidanceParams.ImageCfgInput.SetValue(guidance.ImgCfg)
	// Slg
	guidanceParams.SlgLayerEndInput.SetValue(guidance.Slg.LayerEnd)
	guidanceParams.SlgLayerStartInput.SetValue(guidance.Slg.LayerStart)
	guidanceParams.SlgScaleInput.SetValue(guidance.Slg.Scale)
	guidanceParams.SlgLayersInput.SetText(intSliceToString(guidance.Slg.Layers))
}

func FileToBase64(filePath string) *string {
	if filePath == "" {
		//fmt.Println("Error, empty path:", err)
		return nil // Пустая строка — возвращаем nil
	}
	// 1. Читаем байты файла
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error read file:", err)
		return nil // Ошибка чтения — возвращаем nil
	}

	// 2. Кодируем в Base64
	b64Str := base64.StdEncoding.EncodeToString(fileBytes)

	// 3. Возвращаем указатель на созданную строку
	return &b64Str
}
