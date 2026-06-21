package main

import (
	"fmt"

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
	MainBox    *fyne.Container // Контейнер этой конкретной строки
	SelectCtrl *widget.Select  // Выпадающий список
	Multiplier *NumberStepper  // Поле ввода веса (Multiplier)
	Enabled    *widget.Check   // Чекбокс Enabled
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
	tableHeader := container.NewGridWithColumns(4,
		widget.NewLabel("LORA"),
		widget.NewLabel("MULTIPLIER"),
		widget.NewLabel("Enabled"),
		widget.NewLabel(""),
	)

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
		options = append(options, fmt.Sprintf("%s (%s)", lora.Name, lora.Path))
	}

	// 2. Создаем контроллы для новой строки
	loraSelect := widget.NewSelect(options, func(value string) {
		fmt.Println("Выбран LoRA:", value)
	})
	loraSelect.SetSelected(options[0]) // По умолчанию выбираем первый

	multiplierInput := NewNumberStepper(1, 100, 0.1, 1, false) // Дефолтный множитель 1

	enableCheck := widget.NewCheck("On", func(checked bool) {
		// Логика чекбокса
	})
	enableCheck.SetChecked(true) // По умолчанию включен

	// Объявляем переменную строки заранее, чтобы кнопка Remove могла сослаться на неё
	var row *LoraRow

	removeBtn := widget.NewButton("Remove", func() {
		b.RemoveRow(row)
	})

	// Собираем строку в Grid или HBox, чтобы выровнять по ширине шапки
	/*	rowBox := container.NewGridWithColumns(4,
			loraSelect,
			multiplierInput.Container,
			enableCheck,
			removeBtn,
		)
	*/
	rowBox := container.NewHBox(
		loraSelect,
		multiplierInput.Container,
		enableCheck,
		removeBtn,
	)

	row = &LoraRow{
		MainBox:    rowBox,
		SelectCtrl: loraSelect,
		Multiplier: multiplierInput,
		Enabled:    enableCheck,
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

func (b *LoraBlock) GetCurrentConfig() []map[string]any {
	var result []map[string]any

	for _, row := range b.rows {
		// row.SelectCtrl.Selected вернет строку вида "Name (Path)"
		// При необходимости вы можете вытащить оттуда чистый Name или индекс

		result = append(result, map[string]any{
			"name":       row.SelectCtrl.Selected,
			"multiplier": row.Multiplier.Value(),
			"enabled":    row.Enabled.Checked,
		})
	}

	return result
}
