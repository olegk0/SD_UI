package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const maxHistory = 20

var drawColor = color.RGBA{0, 255, 0, 255} //color.RGBA{255, 0, 255, 255}.

type PaintCanvas struct {
	widget.BaseWidget
	backgroundImage image.Image
	currentMask     *image.RGBA // Указатель на внешнюю маску с прозрачностью
	history         []*image.RGBA
	historyIndex    int
	brushRadius     int
	brushCursor     *canvas.Circle
	//window          fyne.Window
}

func NewPaintCanvas(bg image.Image, mask *image.RGBA) *PaintCanvas {
	cursor := canvas.NewCircle(color.Transparent)
	cursor.StrokeColor = color.RGBA{255, 0, 0, 180}
	cursor.StrokeWidth = 1.5
	cursor.Hide()

	pc := &PaintCanvas{
		backgroundImage: bg,
		currentMask:     mask,
		historyIndex:    0,
		brushRadius:     10,
		brushCursor:     cursor,
		//window:          w,
	}
	// Первый кадр истории — текущее состояние переданной маски
	pc.history = append(pc.history, pc.cloneRGBA(mask))
	pc.ExtendBaseWidget(pc)
	return pc
}

func (pc *PaintCanvas) cloneRGBA(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	copy(dst.Pix, src.Pix)
	return dst
}

// Обновляет не только внутреннее состояние, но и синхронизирует пиксели с оригиналом
func (pc *PaintCanvas) applyHistoryState(src *image.RGBA) {
	copy(pc.currentMask.Pix, src.Pix)
	pc.Refresh()
}

func (pc *PaintCanvas) commitToHistory() {
	pc.history = pc.history[:pc.historyIndex+1]
	pc.history = append(pc.history, pc.cloneRGBA(pc.currentMask))
	pc.historyIndex++

	if len(pc.history) > maxHistory {
		pc.history = pc.history[1:]
		pc.historyIndex--
	}
}

func (pc *PaintCanvas) Undo() {
	if pc.historyIndex > 0 {
		pc.historyIndex--
		pc.applyHistoryState(pc.history[pc.historyIndex])
	}
}

func (pc *PaintCanvas) Redo() {
	if pc.historyIndex < len(pc.history)-1 {
		pc.historyIndex++
		pc.applyHistoryState(pc.history[pc.historyIndex])
	}
}

// Функция полной очистки холста
func (pc *PaintCanvas) Clear() {
	bounds := pc.currentMask.Bounds()
	// Заливаем текущую маску прозрачным цветом
	draw.Draw(pc.currentMask, bounds, image.NewUniform(color.Transparent), image.Point{}, draw.Src)
	pc.commitToHistory() // Записываем очистку в историю, чтобы можно было сделать Undo
	pc.Refresh()
}
func (pc *PaintCanvas) Inverse() {

	for i := 0; i < len(pc.currentMask.Pix); i += 4 {
		// Если пиксель был видимым (Alpha == 255), делаем его полностью прозрачным
		if pc.currentMask.Pix[i+3] == 255 {
			pc.currentMask.Pix[i] = 0
			pc.currentMask.Pix[i+1] = 0
			pc.currentMask.Pix[i+2] = 0
			pc.currentMask.Pix[i+3] = 0
		} else {
			// Если пиксель был прозрачным (Alpha == 0), записываем базовый цвет
			pc.currentMask.Pix[i] = drawColor.R
			pc.currentMask.Pix[i+1] = drawColor.G
			pc.currentMask.Pix[i+2] = drawColor.B
			pc.currentMask.Pix[i+3] = 255
		}
	}
	pc.Refresh()
}

func (pc *PaintCanvas) updateCursor(pos fyne.Position) {
	r := float32(pc.brushRadius)
	pc.brushCursor.Move(fyne.NewPos(pos.X-r, pos.Y-r))
	pc.brushCursor.Resize(fyne.NewSize(r*2, r*2))
	pc.brushCursor.Show()
	pc.brushCursor.Refresh()
}

func (pc *PaintCanvas) MouseIn(ev *desktop.MouseEvent)    { pc.updateCursor(ev.Position) }
func (pc *PaintCanvas) MouseMoved(ev *desktop.MouseEvent) { pc.updateCursor(ev.Position) }
func (pc *PaintCanvas) MouseOut() {
	pc.brushCursor.Hide()
	pc.brushCursor.Refresh()
}

func (pc *PaintCanvas) CreateRenderer() fyne.WidgetRenderer {
	b := pc.backgroundImage.Bounds()
	combined := image.NewRGBA(b)

	raster := canvas.NewRaster(func(w, h int) image.Image {
		draw.Draw(combined, b, pc.backgroundImage, image.Point{}, draw.Src)
		draw.Draw(combined, b, pc.currentMask, image.Point{}, draw.Over)
		return combined
	})

	return &paintCanvasRenderer{pc: pc, raster: raster}
}

type paintCanvasRenderer struct {
	pc     *PaintCanvas
	raster *canvas.Raster
}

func (r *paintCanvasRenderer) Destroy()              {}
func (r *paintCanvasRenderer) Layout(size fyne.Size) { r.raster.Resize(size) }
func (r *paintCanvasRenderer) MinSize() fyne.Size {
	b := r.pc.backgroundImage.Bounds()
	return fyne.NewSize(float32(b.Dx()), float32(b.Dy()))
}
func (r *paintCanvasRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster, r.pc.brushCursor}
}
func (r *paintCanvasRenderer) Refresh() {
	canvas.Refresh(r.raster)
	r.pc.brushCursor.Refresh()
}

func (pc *PaintCanvas) drawAt(pos fyne.Position) {
	x := int(pos.X)
	y := int(pos.Y)
	r := int(float32(pc.brushRadius))
	bounds := pc.currentMask.Bounds()

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < bounds.Dx() && ny >= 0 && ny < bounds.Dy() {
					// ИСПОЛЬЗУЕМ ЯРКИЙ ЗЕЛЕНЫЙ ЦВЕТ (вместо черного)
					pc.currentMask.SetRGBA(nx, ny, drawColor)
				}
			}
		}
	}
	pc.Refresh()
}

func (pc *PaintCanvas) Tapped(ev *fyne.PointEvent) {
	pc.drawAt(ev.Position)
	pc.commitToHistory()
}

func (pc *PaintCanvas) Dragged(ev *fyne.DragEvent) {
	pc.drawAt(ev.Position)
	pc.updateCursor(ev.Position)
}

func (pc *PaintCanvas) DragEnd() { pc.commitToHistory() }

//==============================================

type MaskEditor struct {
	parentWindow fyne.Window
	popUpWin     *widget.PopUp

	canvas        *PaintCanvas
	undoShortcut  *desktop.CustomShortcut
	redoShortcut  *desktop.CustomShortcut
	updateMaskFun func([]byte)
}

func convertGrayToCustomColorRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	rgbaImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := src.At(x, y)

			// Получаем яркость (от 0 до 255)
			grayColor := color.GrayModel.Convert(oldColor).(color.Gray)
			intensity := grayColor.Y

			// Если черный (0) -> полная прозрачность {0, 0, 0, 0}
			// Если белый (255) -> целевой цвет с его оригинальной альфой
			// Промежуточные значения интерполируются
			alphaFactor := float64(intensity) / 255.0

			rgbaImg.Set(x, y, color.RGBA{
				R: uint8(float64(drawColor.R) * alphaFactor),
				// Умножаем каналы на фактор, так как в RGBA для графических движков
				// цвет часто должен соответствовать уровню прозрачности (Premultiplied Alpha)
				G: uint8(float64(drawColor.G) * alphaFactor),
				B: uint8(float64(drawColor.B) * alphaFactor),
				A: uint8(float64(drawColor.A) * alphaFactor),
			})
		}
	}
	return rgbaImg
}

func NewMaskEditor(parentWin fyne.Window, bg image.Image, mask_img image.Image, update_mask func([]byte)) *MaskEditor {
	var mask *image.RGBA
	if mask_img == nil {
		//fmt.Println("mask_img is nil, creating empty mask")
		bounds := bg.Bounds()
		mask = image.NewRGBA(bounds)
		draw.Draw(mask, bounds, image.NewUniform(color.Transparent), image.Point{}, draw.Src)
	} else {
		//bounds := mask_img.Bounds()
		//mask = image.NewRGBA(bounds)
		//draw.Draw(mask, bounds, mask_img, bounds.Min, draw.Src)
		mask = convertGrayToCustomColorRGBA(mask_img)
	}

	painter := NewPaintCanvas(bg, mask)

	me := &MaskEditor{
		parentWindow:  parentWin,
		canvas:        painter,
		updateMaskFun: update_mask,
	}

	me.setupUI()
	me.popUpWin.Show()
	//me.setupShortcuts()

	return me
}

func (me *MaskEditor) setupUI() {
	undoBtn := widget.NewButton("Undo", func() { me.canvas.Undo() })
	redoBtn := widget.NewButton("Redo", func() { me.canvas.Redo() })

	// Новая кнопка очистки маски
	clearBtn := widget.NewButton("Clear", func() { me.canvas.Clear() })

	invBtn := widget.NewButton("Inverse", func() { me.canvas.Inverse() })

	saveBtn := widget.NewButton("Save Mask", func() {
		bounds := me.canvas.currentMask.Bounds()
		monoImg := image.NewGray(bounds)

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := me.canvas.currentMask.RGBAAt(x, y)
				if c.A == 0 { //tranparent
					monoImg.SetGray(x, y, color.Gray{Y: 0})
				} else {
					monoImg.SetGray(x, y, color.Gray{Y: 255})
				}
			}
		}

		var buf bytes.Buffer

		err := png.Encode(&buf, monoImg)
		if err != nil {
			fmt.Print("Error encode:", err.Error())
			return
		}
		me.updateMaskFun(buf.Bytes())
		me.Destroy()
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		me.Destroy()
	})

	brushLabel := widget.NewLabel("Brush: 10px")
	slider := widget.NewSlider(1, 50)
	slider.SetValue(10)
	slider.OnChanged = func(val float64) {
		size := int(val)
		me.canvas.brushRadius = size
		brushLabel.SetText("Brush: " + strconv.Itoa(size) + "px")
	}

	controls := container.NewHBox(
		undoBtn,
		redoBtn,
		widget.NewSeparator(),
		clearBtn,
		invBtn,
		widget.NewSeparator(),
		brushLabel,
		container.NewGridWrap(fyne.NewSize(150, 36), slider),
		widget.NewSeparator(),
		widget.NewSeparator(),
		saveBtn,
		cancelBtn,
	)

	centeredCanvas := container.NewCenter(me.canvas)
	scrollContainer := container.NewScroll(centeredCanvas)
	content := container.NewBorder(controls, nil, nil, nil, scrollContainer)

	me.popUpWin = widget.NewModalPopUp(content, me.parentWindow.Canvas())

	b := me.canvas.backgroundImage.Bounds()
	winW := float32(b.Dx()) + 20
	winH := float32(b.Dy()) + 60
	me.popUpWin.Resize(fyne.NewSize(winW, winH))
}

func (me *MaskEditor) Destroy() {
	//me.removeShortcuts()
	me.popUpWin.Hide()
}

func (me *MaskEditor) removeShortcuts() {
	me.parentWindow.Canvas().RemoveShortcut(me.undoShortcut)
	me.parentWindow.Canvas().RemoveShortcut(me.redoShortcut)
}

func (me *MaskEditor) setupShortcuts() {
	me.undoShortcut = &desktop.CustomShortcut{KeyName: fyne.KeyZ}
	me.parentWindow.Canvas().AddShortcut(me.undoShortcut, func(shortcut fyne.Shortcut) { me.canvas.Undo() })

	me.redoShortcut = &desktop.CustomShortcut{KeyName: fyne.KeyY}
	me.parentWindow.Canvas().AddShortcut(me.redoShortcut, func(shortcut fyne.Shortcut) { me.canvas.Redo() })

}
