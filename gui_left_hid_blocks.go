package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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

func createImageInputsContent() fyne.CanvasObject {
	createUploadCard := func(title, description string) fyne.CanvasObject {
		titleLabel := widget.NewLabel(title)
		titleLabel.Alignment = fyne.TextAlignCenter

		descLabel := widget.NewLabel(description)
		descLabel.Alignment = fyne.TextAlignCenter
		descLabel.Importance = widget.LowImportance

		statusLabel := widget.NewLabel("No file selected")
		statusLabel.Alignment = fyne.TextAlignCenter
		statusLabel.Importance = widget.LowImportance

		selectBtn := widget.NewButton("Select", func() {})

		centerText := container.NewVBox(titleLabel, descLabel, statusLabel)

		bg := canvas.NewRectangle(color.RGBA{R: 245, G: 245, B: 245, A: 255})
		bg.StrokeColor = color.RGBA{R: 220, G: 220, B: 220, A: 255}
		bg.StrokeWidth = 1
		bg.CornerRadius = 8

		cardContent := container.NewStack(bg, container.NewPadded(centerText))

		return container.NewVBox(cardContent, container.NewHBox(selectBtn))
	}

	initImageCard := createUploadCard("Init Image", "Drop, paste, or browse an image to seed\ngeneration.")
	maskImageCard := createUploadCard("Mask Image", "One-channel mask image.")

	topRow := container.NewGridWithColumns(2, initImageCard, maskImageCard)

	controlImageCard := createUploadCard("Control Image", "ControlNet-style guidance image.")

	refImageCardRaw := createUploadCard("Reference Images", "Multiple reference images supported.")
	refStatusExtra := widget.NewLabel("No files selected.")
	refStatusExtra.Importance = widget.LowImportance

	refImageCard := container.NewVBox(
		widget.NewLabel("Reference Images"),
		refImageCardRaw,
		refStatusExtra,
	)

	return container.NewVBox(
		topRow,
		widget.NewSeparator(),
		controlImageCard,
		widget.NewSeparator(),
		refImageCard,
	)
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

func createCacheContent() fyne.CanvasObject {
	modeSelect := widget.NewSelect([]string{"disabled", "enabled", "aggressive"}, func(value string) {})
	modeSelect.SetSelected("disabled")

	cacheOption := widget.NewEntry()
	cacheOption.SetText("threshold=0.25,start=0.15,end=0.95")

	scmMask := widget.NewEntry()
	scmMask.SetPlaceHolder("")

	dynamicCheck := widget.NewCheck("Dynamic SCM Policy", func(checked bool) {})
	dynamicCheck.SetChecked(true)

	return container.NewVBox(
		container.NewVBox(widget.NewLabel("Mode"), modeSelect),
		container.NewVBox(widget.NewLabel("Cache Option"), cacheOption),
		container.NewVBox(widget.NewLabel("SCM Mask"), scmMask),
		dynamicCheck,
	)
}
