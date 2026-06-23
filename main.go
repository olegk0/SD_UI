package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const configFileName = "config.json"

type AppConfig struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// ==========================================
// ОСНОВНОЙ НАБОР ИНТЕРФЕЙСА
// ==========================================

func main() {
	var JobID string
	var CapsResp CapabilitiesResponse

	var batch_imgs []image.Image
	currentBatchPage := 0
	//totalPages := 10

	myApp := app.New()
	myWindow := myApp.NewWindow("Stable Diffusion CPP GUI")
	//myWindow.Resize(fyne.NewSize(1920, 1080))
	myWindow.Resize(fyne.NewSize(1100, 1000))
	myWindow.CenterOnScreen()

	sdAPI := NewSDClient(4 * time.Second)

	// ---- ЛЕВАЯ КОЛОНКА (настройки) ----
	promptInput := widget.NewMultiLineEntry()
	//promptInput.SetText("A hyperrealistic digital oil painting of a magical encounter...")
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

	widthInput := NewNumberStepper(64, 1024, 16, 128, true)
	heightInput := NewNumberStepper(64, 1024, 16, 128, true)

	sizeRow := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Width"), widthInput.Container),
		container.NewVBox(widget.NewLabel("Height"), heightInput.Container),
	)

	batchInput := NewNumberStepper(1, 100, 1, 1, true)
	seedInput := NewNumberStepper(-1, 999999999999, 1, -1, true)

	batchRow := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Batch Count"), batchInput.Container),
		container.NewVBox(widget.NewLabel("Seed"), seedInput.Container),
	)
	sampleParams := createSampleContent()
	GuidanceParams := createGuidanceContent()
	loraPanel := CreateLoraBlock()
	conditioningParams := createConditioningContent()
	vaeTilingParams := createVaeTilingContent()
	cacheParams := createCacheContent()
	hiResParams := createHiResContent()
	imageInputsParams := createImageInputsContent(myWindow)

	accordion := widget.NewAccordion(
		widget.NewAccordionItem("SAMPLE", sampleParams.Container),
		widget.NewAccordionItem("GUIDANCE", GuidanceParams.Container),
		widget.NewAccordionItem("CONDITIONING", conditioningParams.Container),
		widget.NewAccordionItem("LORA", loraPanel.Container),
		widget.NewAccordionItem("IMAGE INPUTS", imageInputsParams.Container),
		widget.NewAccordionItem("VAE TILING", vaeTilingParams.Container),
		widget.NewAccordionItem("CACHE", cacheParams.Container),
		widget.NewAccordionItem("HIRES", hiResParams.Container),
	)
	accordion.MultiOpen = true

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

	cancelBtn := widget.NewButton("Cancel", func() {
		if JobID != "" {
			sdAPI.CancelJob(JobID)
		}
		JobID = ""
	})

	var generateBtn *widget.Button

	testBtn := widget.NewButton("Connect", func() {
		modelLabel.SetText("Model:")
		if ipEntry.Text == "" {
			statusLabel.SetText("Status: Error (Enter IP)")
			return
		}

		// 1. Обновляем адрес
		sdAPI.SetAddress(ipEntry.Text, portEntry.Text)

		// 2. Делаем сетевой запрос (в этот момент UI замрет)
		caps, err := sdAPI.CheckConnection()

		// 3. Этот код выполнится ТОЛЬКО после того, как запрос завершится или отвалится по таймауту
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Status: ❌ Error (%v)", err.Error()))
		} else {
			JobID = ""
			CapsResp = *caps
			generateBtn.Enable()
			statusLabel.SetText("Status: Connected!")
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
			sampleParams.ShiftedTimestepInput.SetValue(float64(caps.Defaults.SampleParams.ShiftedTimestep))
			parseGuidance(caps.Defaults.SampleParams.Guidance, GuidanceParams)
			//-- Features
			if caps.Features.CancelGenerating {
				cancelBtn.Enable()
			} else {
				cancelBtn.Disable()
			}
			//-- Limits
			batchInput.Max = float64(caps.Limits.MaxBatchCount)
			widthInput.Max = float64(caps.Limits.MaxWidth)
			widthInput.Min = float64(caps.Limits.MinWidth)
			heightInput.Min = float64(caps.Limits.MinHeight)
			heightInput.Max = float64(caps.Limits.MaxHeight)
			//-- Upscalers
			ups_names := make([]string, len(caps.Upscalers)+1)
			ups_names[0] = "disabled"
			for i, upscaler := range caps.Upscalers {
				ups_names[i+1] = upscaler.Name
			}
			hiResParams.UpscalerSelect.Options = ups_names
			hiResParams.UpscalerSelect.SetSelected("disabled")

			//--
			//TODO max_queue_size

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

	jobStatusLabel := widget.NewLabelWithStyle("Idle", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	elapsedLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	jobCreatedLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})
	statusCard := container.NewGridWithColumns(3,
		container.NewVBox(widget.NewLabelWithStyle("Status", fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}), jobStatusLabel),
		container.NewVBox(widget.NewLabelWithStyle("Created", fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}), jobCreatedLabel),
		container.NewVBox(widget.NewLabelWithStyle("Elapsed:", fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true}), elapsedLabel),
	)

	infoLabel := widget.NewLabel("Idle.")

	batchCntLabel := widget.NewLabelWithStyle(
		"",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	leftBatchBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		if currentBatchPage > 0 {
			currentBatchPage--
			batchCntLabel.SetText(fmt.Sprintf("%d of %d", currentBatchPage+1, len(batch_imgs)))
			imagePlaceholder.Image = batch_imgs[currentBatchPage]
			imagePlaceholder.Refresh()
		}
	})
	rightBatchBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		if currentBatchPage < len(batch_imgs)-1 {
			currentBatchPage++
			batchCntLabel.SetText(fmt.Sprintf("%d of %d", currentBatchPage+1, len(batch_imgs)))
			imagePlaceholder.Image = batch_imgs[currentBatchPage]
			imagePlaceholder.Refresh()
		}
	})
	batchNavContainer := container.NewBorder(
		nil,
		nil,
		leftBatchBtn,  // Кнопка строго слева
		rightBatchBtn, // Кнопка строго справа
		batchCntLabel, // Текст растягивается по центру
	)
	batchNavContainer.Hide()

	generateBtn = widget.NewButton("Generate Image", func() {
		if sdAPI.isConnected == true {
			generateBtn.Disable() // Блокируем кнопку, чтобы предотвратить повторные нажатия
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
			if hiResParams.UpscalerSelect.Selected == "disabled" || hiResParams.UpscalerSelect.Selected == "None" {
				req.Hires.Enabled = false
			} else {
				req.Hires.Enabled = true
				req.Hires.Upscaler = hiResParams.UpscalerSelect.Selected
				req.Hires.Scale = hiResParams.ScaleInput.value
				req.Hires.TargetHeight = int(hiResParams.TargetHeightInput.value)
				req.Hires.TargetWidth = int(hiResParams.TargetWidthInput.value)
				req.Hires.Steps = int(hiResParams.StepsInput.value)
				req.Hires.DenoisingStrength = hiResParams.DenoisingStrengthInput.value
				req.Hires.CustomSigmas = []float64{}
				req.Hires.UpscaleTileSize = int(hiResParams.UpscaleTileSizeInput.value)
			}

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
			req.CacheMode = cacheParams.ModeSelect.Selected
			req.CacheOption = cacheParams.CacheOption.Text
			req.ScmMask = cacheParams.ScmMask.Text
			req.ScmPolicyDynamic = cacheParams.DynamicScmCheck.Checked
			//--img2img
			req.AutoResizeRefImage = true //TODO ???
			req.RefImages = []string{}
			if imageInputsParams.EnabledCheck.Checked {
				req.InitImage = FileToBase64(imageInputsParams.InitImagePath[0])
				req.MaskImage = FileToBase64(imageInputsParams.MaskImagePath[0])
				req.ControlImage = FileToBase64(imageInputsParams.ControlImagePath[0])
				if imageInputsParams.RefImagePathList != nil {
					for _, filePath := range imageInputsParams.RefImagePathList {
						lstr := FileToBase64(filePath)
						if lstr != nil && len(*lstr) > 0 {
							req.RefImages = append(req.RefImages, *lstr)
						}
					}
				}
			} else {
				req.InitImage = nil
				req.MaskImage = nil
				req.ControlImage = nil
				//req.RefImages = []string{}
			}
			//--
			req.EmbedImageMetadata = true
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
					defer fyne.Do(func() { generateBtn.Enable() })
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

							batch_imgs = nil
							currentBatchPage = 0

							if len(job_stat.Result.Images) > 1 {
								batchNavContainer.Show()
							} else {
								batchNavContainer.Hide()
							}

							for i, b64Data := range job_stat.Result.Images {
								// Берем b64-строку  картинки

								// 2. Декодируем Base64 строку в байты
								imgBytes, err := base64.StdEncoding.DecodeString(b64Data.B64JSON)
								if err != nil {
									fyne.Do(func() {
										infoLabel.SetText(fmt.Sprintf("Error decode Base64: %v\n", err))
										jobStatusLabel.SetText("Error")
									})
									continue
								}
								file_name := fmt.Sprintf("images/%d_%d.png", resp.Created, i)
								os.WriteFile(file_name, imgBytes, 0644)

								img, _, err := image.Decode(bytes.NewReader(imgBytes))
								if err != nil {
									fyne.Do(func() {
										infoLabel.SetText(fmt.Sprintf("Error img parse: %v\n", err))
										jobStatusLabel.SetText("Error")
									})
									continue
								}
								batch_imgs = append(batch_imgs, img)
							}
							if len(batch_imgs) > 0 {
								fyne.Do(func() {
									imagePlaceholder.Image = batch_imgs[0]
									batchCntLabel.SetText(fmt.Sprintf("%d of %d", currentBatchPage+1, len(batch_imgs)))
								})
							}
							fyne.Do(func() {
								imagePlaceholder.Refresh()
								jobStatusLabel.SetText("Completed")
							})
							break
						}
						time.Sleep(time.Second)
						diff_Time := time.Now().Unix() - beginTimestamp
						fyne.Do(func() {
							elapsedLabel.SetText(fmt.Sprintf("%v:%02v", diff_Time/60, diff_Time%60))
							elapsedLabel.Refresh()
						})
					}
				}()

			}
		}
	})
	generateBtn.Importance = widget.HighImportance
	generateBtn.Disable()

	actionRow := container.NewGridWithColumns(2, generateBtn, cancelBtn)

	var galleryBtn *widget.Button

	galleryBtn = widget.NewButton("Gallery", func() {
		galleryBtn.Disable()
		galleryWin := Gallery(myApp, func(params SDCPPParams) {
			sampleParams.SchedulerSelect.SetSelected(params.Sampling.Scheduler)
			sampleParams.MethodSelect.SetSelected(params.Sampling.Method)
			//
			widthInput.SetValue(float64(params.Width))
			heightInput.SetValue(float64(params.Height))
			//-- Loras
			//TODO loraPanel
			//--
			conditioningParams.ControlStrengthInput.SetValue(params.ControlStrength)
			conditioningParams.StrengthInput.SetValue(params.Strength)
			//-- Guidance
			parseGuidance(params.Sampling.Guidance, GuidanceParams)
			//--
			seedInput.SetValue(float64(params.Seed))
			promptInput.Text = params.Prompt.Positive
			negativeInput.Text = params.Prompt.Negative
			//--
			//TODO eta
			//TODO extra_sample_args
			//TODO flow_shift
			sampleParams.ShiftedTimestepInput.SetValue(float64(params.Sampling.ShiftedTimestep))
			sampleParams.StepsInput.SetValue(float64(params.Sampling.Steps))
			//--
			//-- Upscalers
			// TODO hiResParams.UpscalerSelect.SetSelected("disabled")
			leftScroll.Refresh()
		})
		galleryWin.SetOnClosed(func() {
			galleryBtn.Enable()
		})
	})

	rightColumn := container.NewVBox(
		connectionBlock, // блок параметров соединения (над OUTPUT)
		modelLabel,
		widget.NewSeparator(),
		galleryBtn, //widget.NewLabel("OUTPUT"),
		imageContainer,
		batchNavContainer,
		statusCard,
		infoLabel,
		actionRow,
	)

	// ---- РАЗДЕЛИТЕЛЬ ЛЕВОЙ И ПРАВОЙ ЧАСТЕЙ ----
	splitLayout := container.NewHSplit(leftScroll, rightColumn)
	splitLayout.Offset = 0.55

	currentConfig := loadConfig()
	ipEntry.SetText(currentConfig.IP)
	portEntry.SetText(currentConfig.Port)

	myWindow.SetOnClosed(func() { //
		fmt.Println("SetOnClosed")
		currentConfig.IP = ipEntry.Text
		currentConfig.Port = portEntry.Text

		jsonData, err := json.MarshalIndent(currentConfig, "", "    ")
		if err == nil {
			err = os.WriteFile(configFileName, jsonData, 0644)
			if err != nil {
				fmt.Println("Error save config:", err)
			}
		} else {
			fmt.Println("Error make config:", err)
		}

		myApp.Quit()
	})

	// ---- ФИНАЛЬНАЯ КОМПОНОВКА ----
	content := container.NewBorder(nil, statusBar, nil, nil, splitLayout)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
