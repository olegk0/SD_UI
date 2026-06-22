package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Кастомный виджет для обработки кликов по картинке
type TappableImage struct {
	widget.BaseWidget
	image *canvas.Image
	OnTap func()
}

// Создание нового кликабельного изображения
func NewTappableImage(img *canvas.Image, onTap func()) *TappableImage {
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(150, 150)) // Базовый размер миниатюры

	tappable := &TappableImage{
		image: img,
		OnTap: onTap,
	}
	tappable.ExtendBaseWidget(tappable) // Теперь этот метод существует, так как мы используем widget.BaseWidget
	return tappable
}

func NewTappableImageFile(path string, onTap func()) *TappableImage {
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

func (t *TappableImage) Cursor() desktop.Cursor {
	// Возвращаем стандартный курсор руки-указателя (Pointer)
	return desktop.PointerCursor
}

func (t *TappableImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.image)
}

// Реализация интерфейса fyne.Tappable для отслеживания кликов
func (t *TappableImage) Tapped(_ *fyne.PointEvent) {
	if t.OnTap != nil {
		t.OnTap()
	}
}

// ===========================================

func loadImageExif(imgPath string) string {

	// 2. Бинарное чтение ВСЕХ чанков PNG (без лимитов по длине строк)
	exifText := "Metadata not found"
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
											builder.WriteString(line)
											builder.WriteByte('\n')
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

	return exifText
}

// ================================================
// Окно просмотра: с блоком метаданных EXIF
func showFullImage(myApp fyne.App, parentWin fyne.Window, imgPath string, setParams func(SDCPPParams)) *widget.PopUp {
	var popUp *widget.PopUp
	img := canvas.NewImageFromFile(imgPath)
	img.FillMode = canvas.ImageFillContain

	// Текст из EXIF
	exifInfo := loadImageExif(imgPath)
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
			fmt.Println("Tag SDCPP not found!")
			return
		}

		// 2. Вырезаем всё, что идет после "SDCPP: " (это и есть чистый JSON)
		jsonString := exifInfo[startIndex+len(tag):]

		// 3. Парсим JSON в нашу структуру данных
		var params SDCPPParams
		err := json.Unmarshal([]byte(jsonString), &params) // Декодируем байты в объект
		if err != nil {
			fmt.Printf("Error parse JSON: %v\n", err)
			return
		}
		setParams(params)
		/*// 4. Проверяем результат работы: выводим нужные поля на экран
		fmt.Println("=== Успешно распарсено ===")
		fmt.Printf("Промпт:       %s\n", params.Prompt.Positive)
		fmt.Printf("Разрешение:   %dx%d\n", params.Width, params.Height)
		fmt.Printf("Сид (Seed):   %d\n", params.Seed)
		fmt.Printf("Метод/Шаги:   %s (%d шагов)\n", params.Sampling.Method, params.Sampling.Steps)
		fmt.Printf("Используемая LLM: %s\n", params.Models.Llm)
		*/
	})
	copyClipBoardBtn := widget.NewButton("Params to clipboard", func() {
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

// ===========================================
type AdvancedBrowser struct {
	Window     fyne.Window
	RootDir    string
	CurrentDir string
	Grid       *fyne.Container
	TopBar     *fyne.Container
	PathLabel  *widget.Label
	BackButton *widget.Button

	setParamsFunc *func(SDCPPParams)
	popUpImg      *widget.PopUp
	myApp         *fyne.App

	// Списки для массовых операций
	CheckedFolders []string
	CheckedImages  []string

	// Буфер обмена для перемещения (вырезания) картинок
	CutImagesBuffer []string

	// Чекбоксы картинок
	imageChecks []*widget.Check
}

func (ab *AdvancedBrowser) initUI() {
	ab.BackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		if ab.CurrentDir != ab.RootDir {
			ab.CurrentDir = filepath.Dir(ab.CurrentDir)
			ab.Refresh()
		}
	})

	newDirBtn := widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Dir name")
		dialog.ShowCustomConfirm("New dir", "Make", "Cancel", entry, func(ok bool) {
			if ok && entry.Text != "" {
				_ = os.Mkdir(filepath.Join(ab.CurrentDir, entry.Text), 0755)
				ab.Refresh()
			}
		}, ab.Window)
	})

	// УЛУЧШЕНО: Кнопка-переключатель "Выбрать всё / Снять выделение"
	var selectAllBtn *widget.Button
	selectAllBtn = widget.NewButton("Select all", func() {
		// Проверяем, есть ли хотя бы один невыделенный чекбокс
		hasUnchecked := false
		for _, chk := range ab.imageChecks {
			if !chk.Checked {
				hasUnchecked = true
				break
			}
		}

		// Если есть невыделенные — выделяем все. Если всё уже было выделено — снимаем.
		targetState := hasUnchecked
		for _, chk := range ab.imageChecks {
			if chk.Checked != targetState {
				chk.SetChecked(targetState)
			}
		}

		// Меняем текст на кнопке для наглядности
		if targetState {
			selectAllBtn.SetText("Unselect")
		} else {
			selectAllBtn.SetText("Select all")
		}
	})

	renameBtn := widget.NewButtonWithIcon("Name", theme.DocumentCreateIcon(), func() {
		totalSelected := len(ab.CheckedFolders) + len(ab.CheckedImages)
		if totalSelected != 1 {
			dialog.ShowInformation("Attention", "Please select exactly ONE item to rename", ab.Window)
			return
		}

		var oldPath string
		if len(ab.CheckedFolders) == 1 {
			oldPath = ab.CheckedFolders[0]
		} else {
			oldPath = ab.CheckedImages[0]
		}

		entry := widget.NewEntry()
		entry.SetText(filepath.Base(oldPath))

		dialog.ShowCustomConfirm("Rename", "Save", "Cancel", entry, func(ok bool) {
			if ok && entry.Text != "" {
				newPath := filepath.Join(filepath.Dir(oldPath), entry.Text)
				if len(ab.CheckedImages) == 1 && !strings.HasSuffix(strings.ToLower(entry.Text), ".png") {
					newPath += ".png"
				}
				_ = os.Rename(oldPath, newPath)
				ab.Refresh()
			}
		}, ab.Window)
	})

	// УЛУЧШЕНО: Удаление с проверкой на пустые папки и выводом предупреждений
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		foldersCount := len(ab.CheckedFolders)
		imagesCount := len(ab.CheckedImages)

		if foldersCount == 0 && imagesCount == 0 {
			dialog.ShowInformation("Attention", "Nothing selected for deletion", ab.Window)
			return
		}

		msg := fmt.Sprintf("Delete selected items?Folders: %d, Images: %d", foldersCount, imagesCount)

		dialog.ShowConfirm("Confirm deletion", msg, func(confirm bool) {
			if confirm {
				var failedFolders []string

				// 1. Пытаемся удалить выбранные папки
				for _, dirPath := range ab.CheckedFolders {
					err := os.Remove(dirPath)
					if err != nil {
						// Если папка не пустая, os.Remove вернет ошибку. Запоминаем имя папки.
						failedFolders = append(failedFolders, filepath.Base(dirPath))
					}
				}

				// 2. Удаляем выбранные картинки (они удаляются всегда успешно, если есть права)
				for _, imgPath := range ab.CheckedImages {
					_ = os.Remove(imgPath)
				}

				// Обновляем интерфейс, чтобы убрать успешно удаленные элементы
				ab.Refresh()

				// 3. Если какие-то папки не удалились, показываем предупреждение об отказе
				if len(failedFolders) > 0 {
					errMessage := fmt.Sprintf("Cannot delete these folders: %s. Reason: Folders are not empty. Please empty or move files first.", strings.Join(failedFolders, ", "))
					dialog.ShowError(errors.New(errMessage), ab.Window)

				}
			}
		}, ab.Window)
	})

	cutBtn := widget.NewButtonWithIcon("Cut", theme.ContentCutIcon(), func() {
		if len(ab.CheckedImages) == 0 {
			dialog.ShowInformation("Attention", "Select PNG images to move", ab.Window)
			return
		}
		ab.CutImagesBuffer = append([]string{}, ab.CheckedImages...)
		dialog.ShowInformation("Clipboard", fmt.Sprintf("Cut %d images. Go to the destination folder and press Paste.", len(ab.CutImagesBuffer)), ab.Window)
	})

	pasteBtn := widget.NewButtonWithIcon("Paste", theme.ContentPasteIcon(), func() {
		if len(ab.CutImagesBuffer) == 0 {
			dialog.ShowInformation("Attention", "Clipboard is empty", ab.Window)
			return
		}
		for _, srcPath := range ab.CutImagesBuffer {
			destPath := filepath.Join(ab.CurrentDir, filepath.Base(srcPath))
			if srcPath != destPath {
				_ = os.Rename(srcPath, destPath)
			}
		}
		ab.CutImagesBuffer = nil
		ab.Refresh()
	})

	ab.TopBar = container.NewHBox(ab.BackButton, newDirBtn, selectAllBtn, renameBtn, cutBtn, pasteBtn, deleteBtn, ab.PathLabel)
}

func (ab *AdvancedBrowser) Refresh() {
	ab.Grid.Objects = nil
	ab.CheckedFolders = nil
	ab.CheckedImages = nil
	ab.imageChecks = nil

	if ab.CurrentDir == ab.RootDir {
		ab.BackButton.Disable()
		ab.PathLabel.SetText("Home")
	} else {
		ab.BackButton.Enable()
		rel, _ := filepath.Rel(ab.RootDir, ab.CurrentDir)
		ab.PathLabel.SetText("Home / " + rel)
	}

	entries, err := os.ReadDir(ab.CurrentDir)
	if err != nil {
		return
	}

	var folders []os.DirEntry
	var images []os.DirEntry

	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry)
		} else if strings.ToLower(filepath.Ext(entry.Name())) == ".png" {
			images = append(images, entry)
		}
	}

	// 1. Отрисовка ПАПОК (Размер кнопок увеличен за счет GridWrap)
	for _, folder := range folders {
		name := folder.Name()
		fullPath := filepath.Join(ab.CurrentDir, name)

		chk := widget.NewCheck(name, func(checked bool) {
			if checked {
				ab.CheckedFolders = append(ab.CheckedFolders, fullPath)
			} else {
				ab.removeFromSlice(&ab.CheckedFolders, fullPath)
			}
		})

		folderResource := theme.FolderIcon()
		largeIcon := canvas.NewImageFromResource(folderResource)
		largeIcon.FillMode = canvas.ImageFillContain
		//largeIcon.SetMinSize(fyne.NewSize(90, 90))

		btn := NewTappableImage(largeIcon, func() {
			ab.CurrentDir = fullPath
			ab.Refresh()
		})

		folderLayout := container.NewGridWrap(fyne.NewSize(190, 200), btn)
		card := container.NewBorder(nil, chk, nil, nil, folderLayout)
		ab.Grid.Add(container.NewPadded(card))
	}

	for _, imgFile := range images {
		name := imgFile.Name()
		fullPath := filepath.Join(ab.CurrentDir, name)

		chk := widget.NewCheck(name, func(checked bool) {
			if checked {
				ab.CheckedImages = append(ab.CheckedImages, fullPath)
			} else {
				ab.removeFromSlice(&ab.CheckedImages, fullPath)
			}
		})
		ab.imageChecks = append(ab.imageChecks, chk)

		img := NewTappableImageFile(fullPath, func() {
			//ab.openPreview(fullPath)
			ab.popUpImg = showFullImage(*ab.myApp, ab.Window, fullPath, *ab.setParamsFunc)
		})

		card := container.NewVBox(container.NewGridWrap(fyne.NewSize(190, 200), img), chk)
		ab.Grid.Add(container.NewPadded(card))
	}

	ab.Grid.Refresh()
}

/*
	func (ab *AdvancedBrowser) openPreview(path string) {
		previewWin := fyne.CurrentApp().NewWindow("Просмотр: " + filepath.Base(path))
		img := canvas.NewImageFromFile(path)
		img.FillMode = canvas.ImageFillContain
		previewWin.SetContent(img)
		previewWin.Resize(fyne.NewSize(600, 500))
		previewWin.Show()
	}
*/
func (ab *AdvancedBrowser) removeFromSlice(slice *[]string, value string) {
	for i, v := range *slice {
		if v == value {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			break
		}
	}
}

func Gallery(myApp fyne.App, setParams func(SDCPPParams)) {
	galleryWin := myApp.NewWindow("Gallery — Press ESC to close")

	rootDir := "./images"
	_ = os.MkdirAll(rootDir, 0755)
	absRoot, _ := filepath.Abs(rootDir)

	browser := &AdvancedBrowser{
		Window:        galleryWin,
		RootDir:       absRoot,
		CurrentDir:    absRoot,
		Grid:          container.NewGridWrap(fyne.NewSize(200, 250)),
		PathLabel:     widget.NewLabel(""),
		setParamsFunc: &setParams,
		popUpImg:      nil,
		myApp:         &myApp,
	}

	// Перехватываем нажатия клавиш на уровне окна
	galleryWin.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			if key.Name == fyne.KeyEscape && browser.popUpImg != nil && browser.popUpImg.Visible() {
				browser.popUpImg.Hide() // Закрываем/скрываем попап
			} else {
				galleryWin.Close() // Закрываем окно галереи
			}
		}
	})

	browser.initUI()
	browser.Refresh()

	mainLayout := container.NewBorder(
		browser.TopBar,
		nil, nil, nil,
		container.NewVScroll(browser.Grid),
	)

	galleryWin.SetContent(container.NewPadded(mainLayout))
	galleryWin.Resize(fyne.NewSize(1920, 1080))
	galleryWin.Show()
}
