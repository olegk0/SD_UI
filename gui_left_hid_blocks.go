package main

import (
	"fmt"
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ==========================================
// НАПОЛНЕНИЕ РАСКРЫВАЮЩИХСЯ БЛОКОВ
// ==========================================
type SampleParamsPanel struct {
	Container            *fyne.Container
	SchedulerSelect      *widget.Select
	FlowShiftInput       *NumberStepper
	FlowShiftDefCheck    *widget.Check
	MethodSelect         *widget.Select
	StepsInput           *NumberStepper
	EtaInput             *NumberStepper
	EtaDefCheck          *widget.Check
	ShiftedTimestepInput *NumberStepper
}

func createSampleContent() *SampleParamsPanel {
	schedulerSelect := widget.NewSelect([]string{"default"}, func(value string) {})
	schedulerSelect.SetSelected("default")

	flowShiftInput := NewNumberStepper(-10, 10, 0.01, 0, false)
	flowShiftDefCheck := widget.NewCheck("Default", func(checked bool) {
		if checked {
			flowShiftInput.Container.Hide()
		} else {
			flowShiftInput.Container.Show()
		}
	})
	flowShiftDefCheck.SetChecked(true)

	methodSelect := widget.NewSelect([]string{"default"}, func(value string) {})
	methodSelect.SetSelected("default")

	stepsInput := NewNumberStepper(1, 100, 1, 2, true) // min=1, max=100, step=1, initial=1

	etaInput := NewNumberStepper(-10, 10, 0.01, 1, false)
	etaDefCheck := widget.NewCheck("Default", func(checked bool) {
		if checked {
			etaInput.Container.Hide()
		} else {
			etaInput.Container.Show()
		}
	})
	etaDefCheck.SetChecked(true)

	shiftedTimestepInput := NewNumberStepper(0, 100, 1, 0, true)

	grid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Scheduler"), schedulerSelect),
		container.NewVBox(widget.NewLabel("Steps"), stepsInput.Container),
		container.NewVBox(widget.NewLabel("Method"), methodSelect),
		container.NewVBox(container.NewHBox(widget.NewLabel("Flow Shift"), flowShiftDefCheck), flowShiftInput.Container),
		//	container.NewVBox(),
		//	container.NewVBox(widget.NewLabel("Extras")),
		container.NewVBox(container.NewHBox(widget.NewLabel("Eta"), etaDefCheck), etaInput.Container),
		container.NewVBox(widget.NewLabel("Shifted Timestep"), shiftedTimestepInput.Container),
	)

	return &SampleParamsPanel{
		Container:            grid,
		SchedulerSelect:      schedulerSelect,
		FlowShiftInput:       flowShiftInput,
		FlowShiftDefCheck:    flowShiftDefCheck,
		MethodSelect:         methodSelect,
		StepsInput:           stepsInput,
		EtaInput:             etaInput,
		EtaDefCheck:          etaDefCheck,
		ShiftedTimestepInput: shiftedTimestepInput,
	}
}

type GuidanceParamsPanel struct {
	Container          *fyne.Container
	TxtCfgInput        *NumberStepper
	DistilledInput     *NumberStepper
	ImageCfgInput      *NumberStepper
	ImageCfgDefCheck   *widget.Check
	SlgLayersInput     *widget.Entry
	SlgLayerStartInput *NumberStepper
	SlgLayerEndInput   *NumberStepper
	SlgScaleInput      *NumberStepper
}

func createGuidanceContent() *GuidanceParamsPanel {
	txtCfgInput := NewNumberStepper(-10, 10, 0.1, 1, false)
	distilledInput := NewNumberStepper(-10, 10, 0.1, 3.5, false)

	imageCfgInput := NewNumberStepper(-10, 10, 0.1, 0, false)
	imageCfgDefCheck := widget.NewCheck("Default", func(checked bool) {
		if checked {
			imageCfgInput.Container.Hide()
		} else {
			imageCfgInput.Container.Show()
		}
	})
	imageCfgDefCheck.SetChecked(true)
	slgLayersInput := widget.NewEntry()
	slgLayersInput.SetText("7,8,9")
	slgLayersContainer := container.NewGridWrap(fyne.NewSize(200, slgLayersInput.MinSize().Height), slgLayersInput)

	slgLayerStartInput := NewNumberStepper(-10, 10, 0.01, 0.01, false)
	slgLayerEndInput := NewNumberStepper(-10, 10, 0.01, 0.2, false)

	slgScaleInput := NewNumberStepper(-10, 10, 0.01, 0, false)

	grid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("CFG Scale"), txtCfgInput.Container),
		container.NewVBox(widget.NewLabel("Distilled Guidance"), distilledInput.Container),

		container.NewVBox(container.NewHBox(widget.NewLabel("Image CFG"), imageCfgDefCheck), imageCfgInput.Container),
		container.NewVBox(widget.NewLabel("SLG Layers"), slgLayersContainer),

		container.NewVBox(widget.NewLabel("SLG Layer Start"), slgLayerStartInput.Container),
		container.NewVBox(widget.NewLabel("SLG Layer End"), slgLayerEndInput.Container),

		container.NewVBox(widget.NewLabel("SLG Scale"), slgScaleInput.Container),
		container.NewVBox(),
	)

	//showExtras := widget.NewHyperlink("Show extras", nil)

	return &GuidanceParamsPanel{
		//	Container:      container.NewVBox(grid, container.NewHBox(showExtras)),
		Container:          grid,
		TxtCfgInput:        txtCfgInput,
		DistilledInput:     distilledInput,
		ImageCfgInput:      imageCfgInput,
		ImageCfgDefCheck:   imageCfgDefCheck,
		SlgLayersInput:     slgLayersInput,
		SlgLayerStartInput: slgLayerStartInput,
		SlgLayerEndInput:   slgLayerEndInput,
		SlgScaleInput:      slgScaleInput,
	}
}

type ConditioningParamsPanel struct {
	Container            *fyne.Container
	ClipInput            *NumberStepper
	StrengthInput        *NumberStepper
	ControlStrengthInput *NumberStepper
}

func createConditioningContent() *ConditioningParamsPanel {
	clipInput := NewNumberStepper(-1, 10, 0.1, -1, true)

	strengthInput := NewNumberStepper(-10, 10, 0.01, 0.75, false)

	controlStrengthInput := NewNumberStepper(-10, 10, 0.01, 0.9, false) //0.8999999761581421

	grid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("CLIP Skip"), clipInput.Container),
		container.NewVBox(widget.NewLabel("Strength"), strengthInput.Container),
		container.NewVBox(widget.NewLabel("Control Strength"), controlStrengthInput.Container),
		container.NewVBox(),
	)

	return &ConditioningParamsPanel{
		Container:            container.NewVBox(grid),
		ClipInput:            clipInput,
		StrengthInput:        strengthInput,
		ControlStrengthInput: controlStrengthInput,
	}
}

func createLoraContent() fyne.CanvasObject {
	label := widget.NewLabel("No LoRA overrides configured.")
	addLoraBtn := widget.NewButton("Add LoRA", func() {})
	return container.NewVBox(label, container.NewHBox(addLoraBtn))
}

type ImageInputsParamsPanel struct {
	Container        *fyne.Container
	EnabledCheck     *widget.Check
	InitImagePath    []string
	MaskImagePath    []string
	ControlImagePath []string
	RefImagePathList []string
}

func createImageInputsContent(parentWin fyne.Window) *ImageInputsParamsPanel {
	imageInputsParams := ImageInputsParamsPanel{
		InitImagePath:    []string{""},
		MaskImagePath:    []string{""},
		ControlImagePath: []string{""},
		RefImagePathList: []string{""},
	}

	enabledCheck := widget.NewCheck("Enabled", func(checked bool) {})
	createUploadCard := func(title, description string, filesList *[]string, maxItems int, parentWin fyne.Window) fyne.CanvasObject {
		// Контейнер для вертикального списка загруженных картинок с кнопками удаления
		imagesListContainer := container.NewVBox()

		mainImg := canvas.NewImageFromResource(nil)
		mainImg.FillMode = canvas.ImageFillContain
		if maxItems > 1 {
			mainImg.SetMinSize(fyne.NewSize(150, 100))
		} else {
			mainImg.SetMinSize(fyne.NewSize(150, 150))
		}

		bg := canvas.NewRectangle(color.RGBA{R: 245, G: 245, B: 245, A: 255})
		bg.StrokeColor = color.RGBA{R: 220, G: 220, B: 220, A: 255}
		bg.StrokeWidth = 1
		bg.CornerRadius = 8
		imageContainer := container.NewStack(bg, mainImg)

		// Текстовые элементы карточки
		titleLabel := canvas.NewText(title, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		titleLabel.Alignment = fyne.TextAlignCenter

		descLabel := widget.NewLabel(description)
		descLabel.Alignment = fyne.TextAlignCenter
		descLabel.Importance = widget.LowImportance

		statusLabel := canvas.NewText("No file selected", color.RGBA{R: 255, G: 255, B: 0, A: 255})
		statusLabel.Alignment = fyne.TextAlignCenter

		var refreshImagesList func()

		refreshImagesList = func() {
			imagesListContainer.Objects = nil // Очищаем контейнер перед перерисовкой

			for idx, path := range *filesList {
				if path == "" {
					continue
				}

				// 1. Создаем миниатюру
				tImg := canvas.NewImageFromFile(path)
				tImg.FillMode = canvas.ImageFillContain
				tImg.SetMinSize(fyne.NewSize(40, 40))

				tBg := canvas.NewRectangle(color.RGBA{R: 235, G: 235, B: 235, A: 255})
				tBg.CornerRadius = 4
				thumbStack := container.NewStack(tBg, tImg)

				fileNameLabel := widget.NewLabel(filepath.Base(path))
				fileNameLabel.Truncation = fyne.TextTruncateEllipsis // Обрезаем длинные имена

				// Захватываем текущий индекс для замыкания кнопки удаления
				currentIdx := idx

				deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					*filesList = append((*filesList)[:currentIdx], (*filesList)[currentIdx+1:]...)

					if len(*filesList) == 0 {
						statusLabel.Text = "No file selected"
					} else {
						statusLabel.Text = fmt.Sprintf("Selected %d/%d files", len(*filesList), maxItems)
					}

					refreshImagesList()
				})
				deleteBtn.Importance = widget.DangerImportance

				row := container.NewBorder(nil, nil, thumbStack, deleteBtn, fileNameLabel)
				imagesListContainer.Add(row)
			}
			imagesListContainer.Refresh()
		}

		selectBtn := widget.NewButton("Select", func() {
			if maxItems > 1 && len(*filesList) >= maxItems {
				dialog.ShowInformation("Limit reached", fmt.Sprintf("Maximum %d items allowed", maxItems), parentWin)
				return
			}
			fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					dialog.ShowError(err, parentWin)
					return
				}
				if reader == nil {
					return
				}
				defer reader.Close()

				filePath := reader.URI().Path()

				if maxItems > 1 {
					*filesList = append(*filesList, filePath)
					statusLabel.Text = fmt.Sprintf("Selected %d/%d files", len(*filesList), maxItems)
					refreshImagesList()
				} else {
					if len(*filesList) == 0 {
						*filesList = append(*filesList, filePath)
					} else {
						(*filesList)[0] = filePath
					}
					statusLabel.Text = fmt.Sprintf("Selected: %s", filepath.Base(filePath))
					mainImg.File = filePath
					mainImg.Refresh()
					mainImg.Show() // На всякий случай показываем, если было скрыто
				}

			}, parentWin)

			absPath, err := filepath.Abs(ImageRootDir)
			if err != nil {
				absPath = ImageRootDir
			}
			dirURI := storage.NewFileURI(absPath)
			listableURI, err := storage.ListerForURI(dirURI)
			if err != nil {
				//fmt.Println("Ошибка создания ListableURI:", err)
			} else {
				fileDialog.SetLocation(listableURI)
			}
			fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
			fileDialog.Show()
		})

		cleanBtn := widget.NewButton("Clean", func() {
			*filesList = []string{}
			statusLabel.Text = "No file selected"

			if maxItems > 1 {
				refreshImagesList()
			} else {
				mainImg.File = ""
				mainImg.Refresh()
				mainImg.Hide()
			}
		})

		// Компоновка интерфейса карточки
		centerText := container.NewVBox(titleLabel, descLabel, statusLabel)
		buttonsBlock := container.NewHBox(selectBtn, cleanBtn)

		cardContent := container.NewStack(imageContainer, container.NewPadded(centerText))

		// Оборачиваем вертикальный список картинок в Scroll, чтобы окно не раздувалось до бесконечности
		scrollableImageList := container.NewVScroll(imagesListContainer)
		scrollableImageList.SetMinSize(fyne.NewSize(250, 120)) // Ограничиваем высоту зоны списка

		return container.NewVBox(
			cardContent,
			buttonsBlock,
			scrollableImageList, // Вертикальная лента с кнопками удаления появится тут
		)
	}

	initImageCard := createUploadCard("Init Image", "Drop, paste, or browse an image to seed\ngeneration.", &imageInputsParams.InitImagePath, 1, parentWin)
	maskImageCard := createUploadCard("Mask Image", "One-channel mask image.", &imageInputsParams.MaskImagePath, 1, parentWin)
	topRow := container.NewGridWithColumns(2, initImageCard, maskImageCard)

	controlImageCard := createUploadCard("Control Image", "ControlNet-style guidance image.", &imageInputsParams.ControlImagePath, 1, parentWin)
	refImageCardRaw := createUploadCard("Reference Images", "Multiple reference images supported.", &imageInputsParams.RefImagePathList, 10, parentWin)
	refStatusExtra := widget.NewLabel("No files selected.")
	refStatusExtra.Importance = widget.LowImportance

	refImageCard := container.NewVBox(
		widget.NewLabel("Reference Images"),
		refImageCardRaw,
		refStatusExtra,
	)

	imageInputsParams.Container = container.NewVBox(
		enabledCheck,
		topRow,
		widget.NewSeparator(),
		controlImageCard,
		widget.NewSeparator(),
		refImageCard,
	)

	imageInputsParams.EnabledCheck = enabledCheck
	return &imageInputsParams
}

type VaeTilingParamsPanel struct {
	Container     *fyne.Container
	EnabledCheck  *widget.Check
	TileSizeX     *NumberStepper
	TileSizeY     *NumberStepper
	TargetOverlap *NumberStepper
	RelativeSizeX *NumberStepper
	RelativeSizeY *NumberStepper
}

func createVaeTilingContent() *VaeTilingParamsPanel {
	enabledCheck := widget.NewCheck("Enabled", func(checked bool) {})

	tileSizeX := NewNumberStepper(0, 1024, 1, 0, true)
	tileSizeY := NewNumberStepper(0, 1024, 1, 0, true)

	targetOverlap := NewNumberStepper(0, 1024, 0.01, 0.5, false)

	relativeSizeX := NewNumberStepper(0, 1024, 0.01, 0, false)
	relativeSizeY := NewNumberStepper(0, 1024, 0.01, 0, false)

	sizeGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Tile Size X"), tileSizeX.Container),
		container.NewVBox(widget.NewLabel("Tile Size Y"), tileSizeY.Container),
	)

	relativeGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Relative Size X"), relativeSizeX.Container),
		container.NewVBox(widget.NewLabel("Relative Size Y"), relativeSizeY.Container),
	)

	result := container.NewGridWithColumns(1,
		enabledCheck,
		sizeGrid,
		container.NewVBox(widget.NewLabel("Target Overlap"), targetOverlap.Container),
		relativeGrid,
	)

	return &VaeTilingParamsPanel{
		Container:     result,
		EnabledCheck:  enabledCheck,
		TileSizeX:     tileSizeX,
		TileSizeY:     tileSizeY,
		TargetOverlap: targetOverlap,
		RelativeSizeX: relativeSizeX,
		RelativeSizeY: relativeSizeY,
	}
}

type cacheParamsPanel struct {
	Container       *fyne.Container
	ModeSelect      *widget.Select
	CacheOption     *widget.Entry
	ScmMask         *widget.Entry
	DynamicScmCheck *widget.Check
}

func createCacheContent() *cacheParamsPanel {
	options := []string{
		"disabled",
		"easycache",
		"ucache",
		"dbcache",
		"taylorseer",
		"cache-dit",
		"spectrum",
	}
	modeSelect := widget.NewSelect(options, func(value string) {})
	modeSelect.SetSelected("disabled")

	cacheOption := widget.NewEntry()
	cacheOption.SetText("threshold=0.25,start=0.15,end=0.95")

	scmMask := widget.NewEntry()
	scmMask.SetPlaceHolder("")

	dynamicScmCheck := widget.NewCheck("Dynamic SCM Policy", func(checked bool) {})
	dynamicScmCheck.SetChecked(true)

	result := container.NewVBox(
		container.NewVBox(widget.NewLabel("Mode"), modeSelect),
		container.NewVBox(widget.NewLabel("Cache Option"), cacheOption),
		container.NewVBox(widget.NewLabel("SCM Mask"), scmMask),
		dynamicScmCheck,
	)
	return &cacheParamsPanel{
		Container:       result,
		ModeSelect:      modeSelect,
		CacheOption:     cacheOption,
		ScmMask:         scmMask,
		DynamicScmCheck: dynamicScmCheck,
	}
}

type hiresParamsPanel struct {
	Container              *fyne.Container
	UpscalerSelect         *widget.Select
	ScaleInput             *NumberStepper
	TargetWidthInput       *NumberStepper
	TargetHeightInput      *NumberStepper
	StepsInput             *NumberStepper
	DenoisingStrengthInput *NumberStepper
	UpscaleTileSizeInput   *NumberStepper
}

func createHiResContent() *hiresParamsPanel {
	upscalerSelect := widget.NewSelect([]string{"disabled"}, func(value string) {})
	upscalerSelect.SetSelected("disabled")
	/*
	   "scale": 2.0,		number
	   "target_width": 0,		integer
	   "target_height": 0,		integer
	   "steps": 0,				integer
	   "denoising_strength": 0.7,	number
	   "custom_sigmas": [],
	   "upscale_tile_size": 128	integer
	*/
	scaleInput := NewNumberStepper(-100, 100, 0.1, 2.0, false)
	targetWidthInput := NewNumberStepper(0, 10000, 1, 0, true)
	targetHeightInput := NewNumberStepper(0, 10000, 1, 0, true)
	stepsInput := NewNumberStepper(0, 10000, 1, 0, true)
	denoisingStrengthInput := NewNumberStepper(0, 10, 0.01, 0.7, false)
	upscaleTileSizeInput := NewNumberStepper(0, 10000, 1, 128, true)

	grid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Upscaler"), upscalerSelect),
		container.NewVBox(widget.NewLabel("Scale"), scaleInput.Container),
		container.NewVBox(widget.NewLabel("Target Width"), targetWidthInput.Container),
		container.NewVBox(widget.NewLabel("Target Height"), targetHeightInput.Container),
		container.NewVBox(widget.NewLabel("Steps"), stepsInput.Container),
		container.NewVBox(widget.NewLabel("Denoising Strength"), denoisingStrengthInput.Container),
		container.NewVBox(widget.NewLabel("Upscale Tile Size"), upscaleTileSizeInput.Container),
	)
	return &hiresParamsPanel{
		Container:              grid,
		UpscalerSelect:         upscalerSelect,
		ScaleInput:             scaleInput,
		TargetWidthInput:       targetWidthInput,
		TargetHeightInput:      targetHeightInput,
		StepsInput:             stepsInput,
		DenoisingStrengthInput: denoisingStrengthInput,
		UpscaleTileSizeInput:   upscaleTileSizeInput,
	}
}
