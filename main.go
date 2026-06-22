package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const configFileName = "config.json"

type AppConfig struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

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

// ==========================================
// ОСНОВНОЙ НАБОР ИНТЕРФЕЙСА
// ==========================================

func main() {
	var JobID string
	var CapsResp CapabilitiesResponse
	currentConfig := loadConfig()
	myApp := app.New()
	myWindow := myApp.NewWindow("Stable Diffusion CPP GUI")
	//myWindow.Resize(fyne.NewSize(1920, 1080))
	myWindow.Resize(fyne.NewSize(1100, 1000))
	myWindow.CenterOnScreen()

	sdAPI := NewSDClient(4 * time.Second)

	// ---- ЛЕВАЯ КОЛОНКА (настройки) ----
	promptInput := widget.NewMultiLineEntry()
	//promptInput.SetText("A hyperrealistic digital oil painting of a magical encounter...")
	promptInput.SetText("big home")
	promptInput.SetPlaceHolder("Enter your prompt here...")
	promptInput.Wrapping = fyne.TextWrapWord
	promptInput.SetMinRowsVisible(10)

	negativeInput := widget.NewMultiLineEntry()
	negativeInput.SetPlaceHolder("What should be excluded?")
	negativeInput.Wrapping = fyne.TextWrapWord
	negativeInput.SetMinRowsVisible(5)

	//promptContainer := container.NewStack(promptInput)
	//promptContainer.Resize(fyne.NewSize(0, 500))

	//negativeContainer := container.NewStack(negativeInput)
	//negativeContainer.Resize(fyne.NewSize(0, 100))

	widthInput := NewNumberStepper(64, 1024, 32, 128, true)
	heightInput := NewNumberStepper(64, 1024, 32, 128, true)

	sizeRow := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Width"), widthInput.Container),
		container.NewVBox(widget.NewLabel("Height"), heightInput.Container),
	)

	batchInput := NewNumberStepper(1, 100, 1, 1, true)
	seedInput := NewNumberStepper(-1, 999999, 1, -1, true)

	batchRow := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Batch Count"), batchInput.Container),
		container.NewVBox(widget.NewLabel("Seed"), seedInput.Container),
	)
	sampleParams := createSampleContent()
	GuidanceParams := createGuidanceContent()
	loraPanel := CreateLoraBlock()
	conditioningParams := createConditioningContent()
	vaeTilingParams := createVaeTilingContent()
	accordion := widget.NewAccordion(
		widget.NewAccordionItem("SAMPLE", sampleParams.Container),
		widget.NewAccordionItem("GUIDANCE", GuidanceParams.Container),
		widget.NewAccordionItem("CONDITIONING", conditioningParams.Container),
		widget.NewAccordionItem("LORA", loraPanel.Container),
		widget.NewAccordionItem("IMAGE INPUTS", createImageInputsContent()),
		widget.NewAccordionItem("VAE TILING", vaeTilingParams.Container),
		widget.NewAccordionItem("CACHE", createCacheContent()),
	)

	leftColumn := container.NewVBox(
		widget.NewLabel("INPUT"),
		widget.NewLabel("Prompt"),
		promptInput,
		widget.NewLabel("Negative Prompt"),
		negativeInput,
		sizeRow,
		batchRow,
		widget.NewSeparator(),
		accordion,
	)
	leftScroll := container.NewVScroll(leftColumn)

	// ---- НИЖНИЙ СТАТУС-БЛОК ----
	statusLabel := widget.NewLabel("Status: Ready")

	statusBar := container.NewHBox(
		statusLabel,
	)
	statusBar.Resize(fyne.NewSize(0, 30)) // небольшая высота

	modelLabel := widget.NewLabel("Model:")

	// ---- ПРАВАЯ КОЛОНКА (вывод) ----
	// Блок параметров соединения
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("127.0.0.1")
	ipEntry.SetText("127.0.0.1")
	ipContainer := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(200, ipEntry.MinSize().Height)),
		ipEntry,
	)

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("1234")
	portEntry.SetText("1234")
	portContainer := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(80, portEntry.MinSize().Height)),
		portEntry,
	)

	testBtn := widget.NewButton("Connect", func() {
		modelLabel.SetText("Model:")
		if ipEntry.Text == "" {
			statusLabel.SetText("Статус: Ошибка (Введите IP)")
			return
		}

		// 1. Обновляем адрес
		sdAPI.SetAddress(ipEntry.Text, portEntry.Text)

		// 2. Делаем сетевой запрос (в этот момент UI замрет)
		caps, err := sdAPI.CheckConnection()

		// 3. Этот код выполнится ТОЛЬКО после того, как запрос завершится или отвалится по таймауту
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Статус: ❌ Ошибка (%v)", err.Error()))
		} else {
			JobID = ""
			CapsResp = *caps
			statusLabel.SetText("Статус:  Подключено успешно!")
			modelLabel.SetText(fmt.Sprintf("Model: %s (mode: %s)", caps.Model.Name, caps.CurrentMode))
			sampleParams.SchedulerSelect.Options = caps.Schedulers
			sampleParams.MethodSelect.Options = caps.Samplers
			widthInput.SetValue(float64(caps.Defaults.Width))
			heightInput.SetValue(float64(caps.Defaults.Height))
			//-- Loras
			loraPanel.UpdateAvailableLoras(caps.Loras)
			//-- Defaults
			conditioningParams.ControlStrengthInput.SetValue(caps.Defaults.ControlStrength)
			conditioningParams.StrengthInput.SetValue(caps.Defaults.Strength)
			GuidanceParams.DistilledInput.SetValue(caps.Defaults.SampleParams.Guidance.DistilledGuidance)
			GuidanceParams.TxtCfgInput.SetValue(caps.Defaults.SampleParams.Guidance.TxtCfg)
			GuidanceParams.SlgLayerEndInput.SetValue(caps.Defaults.SampleParams.Guidance.Slg.LayerEnd)
			GuidanceParams.SlgLayerStartInput.SetValue(caps.Defaults.SampleParams.Guidance.Slg.LayerStart)
			GuidanceParams.SlgScaleInput.SetValue(caps.Defaults.SampleParams.Guidance.Slg.Scale)
			GuidanceParams.SlgLayersInput.SetText(intSliceToString(caps.Defaults.SampleParams.Guidance.Slg.Layers))
			//-- Limits
			batchInput.Max = float64(caps.Limits.MaxBatchCount)
			widthInput.Max = float64(caps.Limits.MaxWidth)
			widthInput.Min = float64(caps.Limits.MinWidth)
			heightInput.Min = float64(caps.Limits.MinHeight)
			heightInput.Max = float64(caps.Limits.MaxHeight)
			//TODO max_queue_size
			//TODO upscalers
		}
	})

	connectionBlock := container.NewHBox(
		widget.NewLabel("IP:"),
		ipContainer,
		widget.NewLabel("Port:"),
		portContainer,
		testBtn,
	)

	//imagePlaceholder := canvas.NewRectangle(color.RGBA{R: 40, G: 40, B: 40, A: 255})
	imagePlaceholder := canvas.NewImageFromResource(nil) // Создаем пустое изображение
	imagePlaceholder.FillMode = canvas.ImageFillContain
	imagePlaceholder.SetMinSize(fyne.NewSize(500, 400))

	backgroundRect := canvas.NewRectangle(color.RGBA{R: 40, G: 40, B: 40, A: 255})
	backgroundRect.SetMinSize(fyne.NewSize(500, 400))

	imageContainer := container.NewStack(backgroundRect, imagePlaceholder)

	jobStatusLabel := widget.NewLabel("Idle")
	jobQueueLabel := widget.NewLabel("")
	jobCreatedLabel := widget.NewLabel("")
	statusCard := container.NewGridWithColumns(3,
		container.NewVBox(widget.NewLabel("Status"), jobStatusLabel),
		container.NewVBox(widget.NewLabel("Queue"), jobQueueLabel),
		container.NewVBox(widget.NewLabel("Created"), jobCreatedLabel),
	)

	elapsedLabel := widget.NewLabel("")
	infoLabel := widget.NewLabel("Image generation completed.")

	var generateBtn *widget.Button
	generateBtn = widget.NewButton("Generate Image", func() {
		if sdAPI.isConnected == true {
			generateBtn.Disable()      // Блокируем кнопку, чтобы предотвратить повторные нажатия
			defer generateBtn.Enable() // Разблокируем кнопку после завершения функции
			infoLabel.SetText("Running stable-diffusion.cpp...")
			var req ImgGenRequest
			req.Prompt = promptInput.Text
			req.NegativePrompt = negativeInput.Text
			req.Width = int(widthInput.value)
			req.Height = int(heightInput.value)
			req.BatchCount = int(batchInput.value)
			req.Seed = int64(seedInput.value)
			req.ClipSkip = int(conditioningParams.ClipInput.value)
			req.Strength = conditioningParams.StrengthInput.value
			req.ControlStrength = conditioningParams.ControlStrengthInput.value
			//--Sample Params
			req.SampleParams = CapsResp.Defaults.SampleParams
			req.SampleParams.SampleMethod = sampleParams.MethodSelect.Selected
			req.SampleParams.Scheduler = sampleParams.SchedulerSelect.Selected
			req.SampleParams.SampleSteps = int(sampleParams.StepsInput.value)
			if sampleParams.EtaDefCheck.Checked == false {
				req.SampleParams.Eta = &sampleParams.EtaInput.value
			}
			req.SampleParams.ShiftedTimestep = int(sampleParams.ShiftedTimestepInput.value)
			//req.SampleParams.customSigmas
			if sampleParams.FlowShiftDefCheck.Checked == false {
				req.SampleParams.FlowShift = &sampleParams.FlowShiftInput.value
			}
			req.SampleParams.Guidance.TxtCfg = GuidanceParams.TxtCfgInput.value
			if GuidanceParams.ImageCfgDefCheck.Checked {
				req.SampleParams.Guidance.ImgCfg = GuidanceParams.TxtCfgInput.value
			} else {
				req.SampleParams.Guidance.ImgCfg = GuidanceParams.ImageCfgInput.value
			}
			req.SampleParams.Guidance.DistilledGuidance = GuidanceParams.DistilledInput.value
			req.SampleParams.Guidance.Slg.LayerEnd = GuidanceParams.SlgLayerEndInput.value
			req.SampleParams.Guidance.Slg.LayerStart = GuidanceParams.SlgLayerStartInput.value
			req.SampleParams.Guidance.Slg.Layers = stringToIntSlice(GuidanceParams.SlgLayersInput.Text)
			req.SampleParams.Guidance.Slg.Scale = GuidanceParams.SlgScaleInput.value
			//--HiRes
			req.Hires.Enabled = false      //TODO
			req.Hires.Upscaler = "default" //TODO
			req.Hires.Scale = 2.0
			req.Hires.TargetHeight = 0
			req.Hires.TargetWidth = 0
			req.Hires.Steps = 0
			req.Hires.DenoisingStrength = 0.7
			req.Hires.CustomSigmas = []float64{}
			req.Hires.UpscaleTileSize = 128
			//--VAE TILING
			req.VaeTilingParams.Enabled = vaeTilingParams.EnabledCheck.Checked
			req.VaeTilingParams.TileSizeX = int(vaeTilingParams.TileSizeX.value)
			req.VaeTilingParams.TileSizeY = int(vaeTilingParams.TileSizeY.value)
			req.VaeTilingParams.ExtraTilingArgs = ""
			req.VaeTilingParams.RelSizeX = vaeTilingParams.RelativeSizeX.value
			req.VaeTilingParams.RelSizeY = vaeTilingParams.RelativeSizeY.value
			req.VaeTilingParams.TargetOverlap = vaeTilingParams.TargetOverlap.value
			req.VaeTilingParams.TemporalTiling = false
			//--Lora
			req.Lora = loraPanel.GetCurrentConfig()
			//-- Cache
			req.CacheMode = "disabled" //TODO
			//--
			req.AutoResizeRefImage = true //TODO ???
			req.EmbedImageMetadata = true
			req.RefImages = []string{} //TODO

			req.ScmPolicyDynamic = true //TODO
			req.OutputFormat = "png"    //TODO
			req.OutputCompression = 100 //TODO

			resp, err := sdAPI.ImgGetRequest(req)
			if err != nil {
				infoLabel.SetText(fmt.Sprintf("Error: %v", err))
			} else {
				JobID = resp.ID
				jobStatusLabel.SetText("Running...")
				jobStatusLabel.Refresh()
				jobCreatedLabel.SetText(formatTimestamp(&resp.Created))
				jobCreatedLabel.Refresh()
				beginTimestamp := time.Now().Unix()
				fmt.Printf("Response: %+v\n", resp) //infoLabel.SetText(fmt.Sprintf("Response: %v", resp))
				go func() {
					for tmo := 0; tmo < 60*5; tmo++ { //sec
						//TODO jobQueueLabel
						var job_stat *JobResponse
						job_stat, err = sdAPI.GetJobStatus(JobID)
						if err != nil {
							fyne.Do(func() {
								infoLabel.SetText(fmt.Sprintf("Error: %v", err))
							})
							JobID = ""
							break
						}
						if job_stat.Result != nil && len(job_stat.Result.Images) > 0 {
							JobID = ""
							fyne.Do(func() {
								infoLabel.SetText("Image generation completed.")
							})
							// Берем b64-строку первой картинки
							b64Data := job_stat.Result.Images[0].B64JSON

							// 2. Декодируем Base64 строку в байты
							imgBytes, err := base64.StdEncoding.DecodeString(b64Data)
							if err != nil {
								fyne.Do(func() {
									infoLabel.SetText(fmt.Sprintf("Ошибка декодирования Base64: %v\n", err))
									jobStatusLabel.SetText("Error")
								})
								return
							}
							file_name := fmt.Sprintf("images/%d.png", resp.Created)
							os.WriteFile(file_name, imgBytes, 0644)

							// 3. Создаем стандартный image.Image из байт
							img, _, err := image.Decode(bytes.NewReader(imgBytes))
							if err != nil {
								fyne.Do(func() {
									infoLabel.SetText(fmt.Sprintf("Ошибка парсинга изображения: %v\n", err))
									jobStatusLabel.SetText("Error")
								})
								return
							}

							fyne.Do(func() {
								// 4. Записываем изображение в наш Rectangle
								imagePlaceholder.Image = img // Присваиваем декодированное изображение
								imagePlaceholder.Refresh()   // Перерисовываем элемент на экране
								jobStatusLabel.SetText("Completed")
							})
							break
						}
						time.Sleep(time.Second)
						diff_Time := time.Now().Unix() - beginTimestamp
						fyne.Do(func() {
							elapsedLabel.SetText(fmt.Sprintf("Elapsed: %v:%02v", diff_Time/60, diff_Time%60))
							elapsedLabel.Refresh()
						})
					}
				}()

			}
		}
	})
	generateBtn.Importance = widget.HighImportance

	downloadBtn := widget.NewButton("Download", func() {})
	cancelBtn := widget.NewButton("Cancel", func() {
		if JobID != "" {
			sdAPI.CancelJob(JobID)
		}
		JobID = ""
	})
	actionRow := container.NewGridWithColumns(2, downloadBtn, cancelBtn)

	galleryBtn := widget.NewButton("Gallery", func() {
		openGallery(myApp)
	})

	rightColumn := container.NewVBox(
		connectionBlock, // блок параметров соединения (над OUTPUT)
		modelLabel,
		widget.NewSeparator(),
		galleryBtn, //widget.NewLabel("OUTPUT"),
		imageContainer,
		statusCard,
		elapsedLabel,
		infoLabel,
		generateBtn,
		actionRow,
	)

	// ---- РАЗДЕЛИТЕЛЬ ЛЕВОЙ И ПРАВОЙ ЧАСТЕЙ ----
	splitLayout := container.NewHSplit(leftScroll, rightColumn)
	splitLayout.Offset = 0.55

	myWindow.SetOnClosed(func() { //
		currentConfig.IP = ipEntry.Text
		currentConfig.Port = portEntry.Text

		jsonData, err := json.MarshalIndent(currentConfig, "", "    ")
		if err == nil {
			err = os.WriteFile(configFileName, jsonData, 0644)
			if err != nil {
				fmt.Println("Error save config:", err)
			}
		}
	})

	// ---- ФИНАЛЬНАЯ КОМПОНОВКА ----
	content := container.NewBorder(nil, statusBar, nil, nil, splitLayout)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
