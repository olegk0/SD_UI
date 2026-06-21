package main

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 1. Создаем структуру для нашего кастомного макета
type stepperLayout struct {
	entryWidth float32
}

// Метод Layout определяет точное положение и размеры элементов внутри рамки
func (l *stepperLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	border := objects[0]
	entryContainer := objects[1] // Теперь это контейнер, а не голый entry
	btns := objects[2]

	// Рамка занимает всю выделенную площадь
	border.Resize(size)
	border.Move(fyne.NewPos(0, 0))

	// Кнопки прижимаются к правому краю, ширина 20
	btnsWidth := float32(20)
	btns.Resize(fyne.NewSize(btnsWidth, size.Height))
	btns.Move(fyne.NewPos(size.Width-btnsWidth-2, 0))

	// Контейнер текстового поля занимает всё оставшееся место слева
	entryContainer.Resize(fyne.NewSize(size.Width-btnsWidth-6, size.Height))
	entryContainer.Move(fyne.NewPos(2, 0))
}

// Метод MinSize сообщает Fyne (и чекбоксу "On") честные минимальные габариты контрола
func (l *stepperLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	totalWidth := l.entryWidth + 20 + 6
	return fyne.NewSize(totalWidth, 38)
}

type NumberStepper struct {
	Container fyne.CanvasObject
	Min, Max  float64
	Step      float64
	IsInteger bool
	OnChanged func(float64)

	entry *widget.Entry
	value float64
}

func (s *NumberStepper) Value() float64 {
	return s.value
}

func (s *NumberStepper) formatValue(val float64) string {
	if s.IsInteger {
		return fmt.Sprintf("%.0f", val)
	}
	return strconv.FormatFloat(val, 'g', -1, 64)
}

func (s *NumberStepper) roundToStep(val float64) float64 {
	if s.IsInteger && s.Step != 1 && s.Step > 0 {
		return math.Round(val/s.Step) * s.Step
	}
	if s.IsInteger {
		return math.Round(val)
	}
	return val
}

func (s *NumberStepper) SetValue(val float64) {
	val = s.roundToStep(val)
	if val < s.Min {
		val = s.Min
	}
	if val > s.Max {
		val = s.Max
	}

	if s.value != val {
		s.value = val
		s.entry.SetText(s.formatValue(val))
		if s.OnChanged != nil {
			s.OnChanged(val)
		}
	}
}

func NewNumberStepper(min, max, step, initial float64, isInteger bool) *NumberStepper {
	stepper := &NumberStepper{
		Min:       min,
		Max:       max,
		Step:      step,
		IsInteger: isInteger,
		entry:     widget.NewEntry(),
	}

	stepper.value = stepper.roundToStep(initial)
	stepper.entry.SetText(stepper.formatValue(stepper.value))

	stepper.entry.OnChanged = func(str string) {
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return
		}
		val = stepper.roundToStep(val)
		if val >= stepper.Min && val <= stepper.Max {
			if stepper.value != val {
				stepper.value = val
				if stepper.OnChanged != nil {
					stepper.OnChanged(val)
				}
			}
		}
	}

	stepper.entry.Validator = func(str string) error {
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("must be a number")
		}
		if val < stepper.Min || val > stepper.Max {
			return fmt.Errorf("must be between %g and %g", stepper.Min, stepper.Max)
		}
		if stepper.IsInteger && stepper.Step != 1 && stepper.Step > 0 {
			if math.Mod(val, stepper.Step) != 0 {
				return fmt.Errorf("must be a multiple of %g", stepper.Step)
			}
		}
		return nil
	}

	incBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		stepper.SetValue(stepper.value + stepper.Step)
	})
	incContainer := container.NewGridWrap(fyne.NewSize(20, 18), incBtn)

	decBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		stepper.SetValue(stepper.value - stepper.Step)
	})
	decContainer := container.NewGridWrap(fyne.NewSize(20, 18), decBtn)

	buttonsContainer := container.New(
		layout.NewVBoxLayout(),
		incContainer,
		decContainer,
	)

	entryWidth := float32(200)
	//if !isInteger {
	//	entryWidth = 80
	//}

	// ВАЖНОЕ ИСПРАВЛЕНИЕ: Оборачиваем stepper.entry в GridWrap.
	// Это принудительно урежет минимальные требования самого текстового поля
	// до нужных нам 65 или 80 пикселей, и оно перестанет раздувать весь макет.
	entryContainer := container.NewGridWrap(fyne.NewSize(entryWidth, 38), stepper.entry)

	// Бордюр вокруг всего контрола
	borderColor := theme.ForegroundColor()
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = borderColor
	border.StrokeWidth = 1
	border.CornerRadius = 3

	// Инициализируем наш кастомный макет
	myLayout := &stepperLayout{entryWidth: entryWidth}

	// ВНИМАНИЕ: Заменяем stepper.entry на подготовленный entryContainer
	// Порядок для макета: 0 - рамка, 1 - контейнер с полем ввода, 2 - кнопки
	content := container.New(myLayout, border, entryContainer, buttonsContainer)

	// Передаем точные размеры наружу
	totalSize := myLayout.MinSize(content.Objects)
	stepper.Container = container.NewGridWrap(totalSize, content)

	return stepper
}
