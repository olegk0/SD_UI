package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LoraBlock struct {
	Container      *fyne.Container // Главный контейнер всего блока
	rowsContainer  *fyne.Container // Контейнер только для строк с LoRA
	availableLoras []LoraInfo      // Хранилище распарсенных данных с сервера
	rows           []*LoraRow      // Список активных строк на экране
	addBtn         *widget.Button  // Кнопка добавления, нужна для Enable/Disable
}

type LoraRow struct {
	MainBox     *fyne.Container // Контейнер этой конкретной строки
	SelectCtrl  *widget.Select  // Выпадающий список
	Multiplier  *NumberStepper  // Поле ввода веса (Multiplier)
	IsHighNoise *widget.Check   // Чекбокс IsHighNoise
}

func CreateLoraBlock() *LoraBlock {
	rowsContainer := container.NewVBox()
	block := &LoraBlock{
		rowsContainer:  rowsContainer,
		availableLoras: []LoraInfo{},
		rows:           []*LoraRow{},
	}

	// Кнопка добавления новой строки
	addBtn := widget.NewButton("Add LoRA", func() {
		block.AddRow()
	})
	addBtn.Disable() // Изначально выключена, пока нет данных с сервера
	block.addBtn = addBtn

	// Собираем весь блок (Заголовок, список строк, кнопка добавления)
	//header := widget.NewLabelWithStyle("LORA", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Шапка таблицы для визуального соответствия вашему макету
	/*
		tableHeader := container.NewGridWithColumns(4,
			widget.NewLabel("LORA"),
			widget.NewLabel("MULTIPLIER"),
			widget.NewLabel("IsHighNoise"),
			widget.NewLabel(""),
		)
	*/
	leftHeader := container.NewGridWithColumns(2,
		widget.NewLabel("LORA"),
		widget.NewLabel("MULTIPLIER"),
	)

	rightHeader := container.NewHBox(
		widget.NewLabel("IsHighNoise"),
		widget.NewLabel(""),
	)

	tableHeader := container.NewBorder(nil, nil, nil, rightHeader, leftHeader)

	block.Container = container.NewVBox(
		//header,
		tableHeader,
		rowsContainer,
		container.NewHBox(addBtn),
	)

	return block
}

// Метод обновления данных, вызывается при успешном подключении к серверу
func (b *LoraBlock) UpdateAvailableLoras(loras []LoraInfo) {
	b.availableLoras = loras

	// Очищаем старые строки, если они были
	b.rowsContainer.Objects = nil
	b.rows = nil
	b.rowsContainer.Refresh()

	// Активируем кнопку "Add LoRA", если сервер прислал доступные варианты
	if len(loras) > 0 {
		b.addBtn.Enable()
	}
}

// Метод динамического добавления строки
func (b *LoraBlock) AddRow() {
	if len(b.availableLoras) == 0 {
		return
	}

	// 1. Формируем список строк для widget.NewSelect
	var options []string
	for _, lora := range b.availableLoras {
		//options = append(options, fmt.Sprintf("%s (%s)", lora.Name, lora.Path))
		options = append(options, lora.Path)
	}

	// 2. Создаем контроллы для новой строки
	loraSelect := widget.NewSelect(options, func(value string) {
		//fmt.Println("Выбран LoRA:", value)
	})
	loraSelect.SetSelected(options[0]) // По умолчанию выбираем первый

	multiplierInput := NewNumberStepper(1, 100, 0.1, 1, false) // Дефолтный множитель 1

	isHighNoiseCheck := widget.NewCheck("", func(checked bool) {
		// Логика чекбокса
	})
	//isHighNoiseCheck.SetChecked(false) // По умолчанию включен

	// Объявляем переменную строки заранее, чтобы кнопка Remove могла сослаться на неё
	var row *LoraRow

	removeBtn := widget.NewButton("Remove", func() {
		b.RemoveRow(row)
	})

	// Собираем строку в Grid или HBox, чтобы выровнять по ширине шапки
	/*	rowBox := container.NewGridWithColumns(4,
			loraSelect,
			multiplierInput.Container,
			isHighNoiseCheck,
			removeBtn,
		)
	*/

	leftBox := container.NewGridWithColumns(2,
		loraSelect,
		multiplierInput.Container,
	)

	rightBox := container.NewHBox(
		container.NewVBox(),
		isHighNoiseCheck,
		removeBtn,
		container.NewVBox(),
	)

	rowBox := container.NewBorder(nil, nil, nil, rightBox, leftBox)

	row = &LoraRow{
		MainBox:     rowBox,
		SelectCtrl:  loraSelect,
		Multiplier:  multiplierInput,
		IsHighNoise: isHighNoiseCheck,
	}

	// 3. Регистрируем строку в менеджере и добавляем на экран
	b.rows = append(b.rows, row)
	b.rowsContainer.Add(rowBox)
	b.rowsContainer.Refresh()
}

// Метод удаления конкретной строки
func (b *LoraBlock) RemoveRow(rowToRemove *LoraRow) {
	// Удаляем визуальный объект из контейнера Fyne
	b.rowsContainer.Remove(rowToRemove.MainBox)

	// Удаляем из нашего внутреннего слайса rows трекинга данных
	for i, r := range b.rows {
		if r == rowToRemove {
			b.rows = append(b.rows[:i], b.rows[i+1:]...)
			break
		}
	}
	b.rowsContainer.Refresh()
}

func (b *LoraBlock) GetCurrentConfig() []LoraParams {
	result := []LoraParams{}

	for _, row := range b.rows {
		result = append(result, LoraParams{
			Path:        row.SelectCtrl.Selected,
			Multiplier:  row.Multiplier.Value(),
			IsHighNoise: row.IsHighNoise.Checked,
		})
	}

	return result
}
