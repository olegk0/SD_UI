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
	MethodSelect         *widget.Select
	StepsInput           *NumberStepper
	EtaInput             *NumberStepper
	ShiftedTimestepInput *NumberStepper
}

func createSampleContent() *SampleParamsPanel {
	schedulerSelect := widget.NewSelect([]string{"default"}, func(value string) {})
	schedulerSelect.SetSelected("default")

	flowShiftInput := NewNumberStepper(-10, 10, 0.01, 0, false)

	methodSelect := widget.NewSelect([]string{"default"}, func(value string) {})
	methodSelect.SetSelected("default")

	stepsInput := NewNumberStepper(1, 100, 1, 2, true) // min=1, max=100, step=1, initial=1

	etaInput := NewNumberStepper(-10, 10, 0.01, 1, false)
	shiftedTimestepInput := NewNumberStepper(0, 100, 1, 0, true)

	grid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Scheduler"), schedulerSelect),
		container.NewVBox(widget.NewLabel("Flow Shift"), flowShiftInput.Container),
		container.NewVBox(widget.NewLabel("Method"), methodSelect),
		container.NewVBox(widget.NewLabel("Steps"), stepsInput.Container),
		container.NewVBox(),
		container.NewVBox(widget.NewLabel("Extras")),
		container.NewVBox(widget.NewLabel("Eta"), etaInput.Container),
		container.NewVBox(widget.NewLabel("Shifted Timestep"), shiftedTimestepInput.Container),
	)

	return &SampleParamsPanel{
		Container:            grid,
		SchedulerSelect:      schedulerSelect,
		FlowShiftInput:       flowShiftInput,
		MethodSelect:         methodSelect,
		StepsInput:           stepsInput,
		EtaInput:             etaInput,
		ShiftedTimestepInput: shiftedTimestepInput,
	}
}

type GuidanceParamsPanel struct {
	Container      *fyne.Container
	CfgInput       *NumberStepper
	DistilledInput *NumberStepper
}

func createGuidanceContent() *GuidanceParamsPanel {
	cfgInput := NewNumberStepper(-10, 10, 0.1, 1, false)

	distilledInput := NewNumberStepper(-10, 10, 0.1, 3.5, false)

	grid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("CFG Scale"), cfgInput.Container),
		container.NewVBox(widget.NewLabel("Distilled Guidance"), distilledInput.Container),
	)

	showExtras := widget.NewHyperlink("Show extras", nil)

	return &GuidanceParamsPanel{
		Container:      container.NewVBox(grid, container.NewHBox(showExtras)),
		CfgInput:       cfgInput,
		DistilledInput: distilledInput,
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

func createVaeTilingContent() fyne.CanvasObject {
	enabledCheck := widget.NewCheck("Enabled", func(checked bool) {})

	tileSizeX := widget.NewEntry()
	tileSizeX.SetText("0")
	tileSizeY := widget.NewEntry()
	tileSizeY.SetText("0")

	targetOverlap := widget.NewEntry()
	targetOverlap.SetText("0.5")

	relativeSizeX := widget.NewEntry()
	relativeSizeX.SetText("0")
	relativeSizeY := widget.NewEntry()
	relativeSizeY.SetText("0")

	sizeGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Tile Size X"), tileSizeX),
		container.NewVBox(widget.NewLabel("Tile Size Y"), tileSizeY),
	)

	relativeGrid := container.NewGridWithColumns(2,
		container.NewVBox(widget.NewLabel("Relative Size X"), relativeSizeX),
		container.NewVBox(widget.NewLabel("Relative Size Y"), relativeSizeY),
	)

	return container.NewVBox(
		enabledCheck,
		sizeGrid,
		container.NewVBox(widget.NewLabel("Target Overlap"), targetOverlap),
		relativeGrid,
	)
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
