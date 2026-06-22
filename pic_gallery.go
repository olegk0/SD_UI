package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type ImageData struct {
	Description string `json:"description"`
}

type ImageItem struct {
	ImgPath  string
	ExifInfo string
	Checked  bool
}

// Кастомный виджет для обработки кликов по картинке
type TappableImage struct {
	widget.BaseWidget
	image *canvas.Image
	OnTap func()
}

// Создание нового кликабельного изображения
func NewTappableImage(path string, onTap func()) *TappableImage {
	img := &canvas.Image{File: path}
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(150, 150)) // Базовый размер миниатюры

	tappable := &TappableImage{
		image: img,
		OnTap: onTap,
	}
	tappable.ExtendBaseWidget(tappable) // Теперь этот метод существует, так как мы используем widget.BaseWidget
	return tappable
}

// Обязательный метод для отображения содержимого кастомного виджета
func (t *TappableImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.image)
}

// Реализация интерфейса fyne.Tappable для отслеживания кликов
func (t *TappableImage) Tapped(_ *fyne.PointEvent) {
	if t.OnTap != nil {
		t.OnTap()
	}
}

func openGallery(myApp fyne.App) {
	galleryWin := myApp.NewWindow("Gallery")
	galleryWin.Resize(fyne.NewSize(1920, 1080))

	var popUpImg *widget.PopUp

	// Перехватываем нажатия клавиш на уровне окна
	galleryWin.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			if key.Name == fyne.KeyEscape && popUpImg != nil && popUpImg.Visible() {
				popUpImg.Hide() // Закрываем/скрываем попап
			} else {
				galleryWin.Close() // Закрываем окно галереи
			}
		}
	})

	mainContainer := container.NewStack()

	var refreshGallery func()
	refreshGallery = func() {
		items := loadImagesFromFolder("images")

		if len(items) == 0 {
			mainContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Images not found in 'images' folder.")}
			mainContainer.Refresh()
			return
		}

		cellSize := fyne.NewSize(200, 200)
		gridLayout := layout.NewGridWrapLayout(cellSize)

		var grid_items []fyne.CanvasObject
		for i := range items {
			item := &items[i]

			// Нажатие на саму картинку теперь открывает окно просмотра
			imgWidget := NewTappableImage(item.ImgPath, func() {
				popUpImg = showFullImage(myApp, galleryWin, item.ImgPath, item.ExifInfo) // Передаем родительское окно для модальности
			})

			chk := widget.NewCheck("", func(b bool) {
				item.Checked = b
			})

			// Компактная карточка: Клик по картинке увеличивает, чекбокс снизу управляет выбором
			itemBox := container.NewVBox(
				imgWidget,
				container.NewHBox(chk, widget.NewLabel(filepath.Base(item.ImgPath))),
			)
			grid_items = append(grid_items, itemBox)
		}

		gridContainer := container.New(gridLayout, grid_items...)

		deleteSelectedBtn := widget.NewButton("Delete Selected", func() {
			for _, item := range items {
				if item.Checked {
					_ = os.Remove(item.ImgPath)
				}
			}
			refreshGallery()
		})

		scroll := container.NewVScroll(gridContainer)
		content := container.NewBorder(nil, deleteSelectedBtn, nil, nil, scroll)
		mainContainer.Objects = []fyne.CanvasObject{content}
		mainContainer.Refresh()
	}

	refreshGallery()
	galleryWin.SetContent(mainContainer)
	galleryWin.Show()
}

// Окно просмотра: с блоком метаданных EXIF
func showFullImage(myApp fyne.App, parentWin fyne.Window, imgPath, exifInfo string) *widget.PopUp {
	var popUp *widget.PopUp
	img := canvas.NewImageFromFile(imgPath)
	img.FillMode = canvas.ImageFillContain

	// Текст из EXIF
	exifLabel := widget.NewLabel(exifInfo)
	exifLabel.Wrapping = fyne.TextWrapWord

	// Оформляем блок метаданных красивыми разделителями
	infoContainer := container.NewVBox(
		widget.NewSeparator(),
		widget.NewSeparator(),
		//widget.NewLabelWithStyle("Info:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		exifLabel,
	)

	// Упаковываем нижнюю панель в Scroll на случай, если EXIF-данных будет слишком много
	bottomScroll := container.NewVScroll(infoContainer)
	bottomScroll.SetMinSize(fyne.NewSize(0, 180))

	closeBtn := widget.NewButton("X", func() {
		popUp.Hide()
	})
	useThisBtn := widget.NewButton("Use this parameters", func() {
		// 1. Находим маркер "SDCPP: " в тексте
		tag := "SDCPP: "
		startIndex := strings.Index(exifInfo, tag)
		if startIndex == -1 {
			fmt.Println("Тег SDCPP не найден в логе!")
			return
		}

		// 2. Вырезаем всё, что идет после "SDCPP: " (это и есть чистый JSON)
		jsonString := exifInfo[startIndex+len(tag):]

		// 3. Парсим JSON в нашу структуру данных
		var params SDCPPParams
		err := json.Unmarshal([]byte(jsonString), &params) // Декодируем байты в объект
		if err != nil {
			fmt.Printf("Ошибка парсинга JSON: %v\n", err)
			return
		}

		// 4. Проверяем результат работы: выводим нужные поля на экран
		fmt.Println("=== Успешно распарсено ===")
		fmt.Printf("Промпт:       %s\n", params.Prompt.Positive)
		fmt.Printf("Разрешение:   %dx%d\n", params.Width, params.Height)
		fmt.Printf("Сид (Seed):   %d\n", params.Seed)
		fmt.Printf("Метод/Шаги:   %s (%d шагов)\n", params.Sampling.Method, params.Sampling.Steps)
		fmt.Printf("Используемая LLM: %s\n", params.Models.Llm)
	})
	copyClipBoardBtn := widget.NewButton("Copy to Clipboard", func() {
		myApp.Clipboard().SetContent(exifInfo)
	})

	bottomButtonBar := container.NewHBox(
		layout.NewSpacer(),
		useThisBtn,
		layout.NewSpacer(),
		copyClipBoardBtn,
		layout.NewSpacer(),
	)

	bottomPanel := container.NewVBox(
		bottomScroll,
		bottomButtonBar,
	)

	topButtonBar := container.NewHBox(
		layout.NewSpacer(),
		closeBtn,
	)

	//topPanel := container.NewVBox(
	//topButtonBar,
	//widget.NewSeparator(),
	//img,
	//)

	content := container.NewBorder(topButtonBar, bottomPanel, nil, nil, img)

	popUp = widget.NewModalPopUp(content, parentWin.Canvas())
	popUp.Resize(fyne.NewSize(1024, 768))
	popUp.Show()
	return popUp
}

// Сканирование папки, чтение JSON и бинарный разбор ВСЕХ скрытых метаданных PNG любой длины
func loadImagesFromFolder(dir string) []ImageItem {
	var items []ImageItem

	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Ошибка чтения папки:", err)
		return items
	}

	for _, file := range files {
		if !file.IsDir() && strings.ToLower(filepath.Ext(file.Name())) == ".png" {
			imgPath := filepath.Join(dir, file.Name())

			// 2. Бинарное чтение ВСЕХ чанков PNG (без лимитов по длине строк)
			exifText := "Метаданные не найдены"
			imgFile, err := os.Open(imgPath)
			if err == nil {
				var builder strings.Builder

				// Проверяем сигнатуру PNG
				header := make([]byte, 8)
				if _, err := io.ReadFull(imgFile, header); err == nil && bytes.Equal(header, []byte("\x89PNG\r\n\x1a\n")) {

					for {
						// Читаем длину текущего блока (4 байта)
						var length uint32
						if err := binary.Read(imgFile, binary.BigEndian, &length); err != nil {
							break
						}

						// Читаем имя чанка (4 байта, например tEXt, iTXt, zTXt)
						typeBytes := make([]byte, 4)
						if _, err := io.ReadFull(imgFile, typeBytes); err != nil {
							break
						}
						chunkType := string(typeBytes)

						// Проверяем, содержит ли чанк текстовые метаданные
						if chunkType == "tEXt" || chunkType == "iTXt" || chunkType == "zTXt" {
							chunkData := make([]byte, length)
							if _, err := io.ReadFull(imgFile, chunkData); err == nil {

								// Разделяем Ключ и Данные по нулевому байту 0x00
								parts := bytes.SplitN(chunkData, []byte{0x00}, 2)
								if len(parts) >= 2 {
									key := string(parts[0])
									valBytes := parts[1]

									// Корректируем смещение для iTXt (пропускаем системные флаги языка/сжатия)
									if chunkType == "iTXt" && len(valBytes) > 2 {
										subParts := bytes.Split(valBytes, []byte{0x00})
										if len(subParts) > 0 {
											valBytes = subParts[len(subParts)-1]
										}
									}

									//fmt.Printf("% X\n", valBytes) // Выведет: 48 65 6c 6c 6f
									// Фильтруем бинарные данные, вытаскивая чистый UTF-8 текст любой длины
									var cleanVal strings.Builder
									for _, r := range string(valBytes) {
										// r == utf8.RuneError означает, что байты изначально были битыми
										if r == utf8.RuneError {
											continue
										}

										// Проверяем ASCII управляющие символы и печатные символы
										if (r >= 32 && r <= 126) || r == '\n' || r == '\r' || r == '\t' || r > 127 {
											cleanVal.WriteRune(r) // Записываем целую руну, а не один байт
										}
									}
									//fmt.Printf("\n% X\n", cleanVal.String()) // Выведет: 48 65 6c 6c 6f
									valueStr := strings.TrimSpace(cleanVal.String())
									if valueStr != "" {
										//fmt.Printf("Найден чанк %s с ключом '%s' и данными: %s\n", chunkType, key, valueStr)

										// Если внутри текстового блока лежит XMP (XML) структура (характерно для Photoshop/AI)
										if strings.Contains(valueStr, "<?xpacket") || strings.Contains(valueStr, "<x:xmpmeta") {
											// Построчно парсим XML и забираем все содержательные параметры
											lines := strings.Split(valueStr, "\n")
											for _, line := range lines {
												line = strings.TrimSpace(line)
												// Ищем строки метаданных exif, tiff или параметров генерации
												if strings.Contains(line, "exif:") || strings.Contains(line, "tiff:") || strings.Contains(line, "xmp:") {
													// Очищаем XML теги для красивого вывода: <exif:Model>Camera</exif:Model> -> exif:Model: Camera
													line = strings.NewReplacer("<", "", ">", "", "/", "").Replace(line)
													builder.WriteString(line + "\n")
												}
											}
										} else {
											// Обычный текстовый блок (Ключ: Значение)
											builder.WriteString(fmt.Sprintf("%s: %s\n", key, valueStr))
										}
									}
								}
							}
						} else {
							// Пропускаем нетекстовые чанки (например, IDAT с пикселями) без загрузки в память
							if _, err := imgFile.Seek(int64(length), io.SeekCurrent); err != nil {
								break
							}
						}

						// Пропускаем CRC сумму (4 байта в конце каждого чанка)
						if _, err := imgFile.Seek(4, io.SeekCurrent); err != nil {
							break
						}

						if chunkType == "IEND" {
							break // Конец PNG структуры
						}
					}
				}
				imgFile.Close()

				if builder.Len() > 0 {
					exifText = builder.String()
				}
			}

			items = append(items, ImageItem{
				ImgPath:  imgPath,
				ExifInfo: exifText,
				Checked:  false,
			})
		}
	}

	return items
}
