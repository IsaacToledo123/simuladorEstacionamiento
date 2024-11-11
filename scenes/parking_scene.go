package scenes

import (
    "fmt"
    "image/color"
    _"time"
    "strconv"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "holafyne/services"
    "fyne.io/fyne/v2/theme"
)

type ParkingScene struct {
    window         fyne.Window
    simulation     *services.Simulation
    spacesLabel    *widget.Label
    logBox         *widget.TextGrid
    startButton    *widget.Button
    stopButton     *widget.Button
    spaceIcons     []*canvas.Rectangle
    carImages      []*canvas.Image
    queueIcons     []*canvas.Rectangle
    queueBox       *fyne.Container
    statsContainer *fyne.Container
    gameContainer  *fyne.Container
    maxQueueSize   int
}

func NewParkingScene(window fyne.Window) *ParkingScene {
    scene := &ParkingScene{
        window:      window,
        spacesLabel: widget.NewLabel("Espacios disponibles: " + strconv.Itoa(services.PARKING_CAPACITY)),
        logBox:      widget.NewTextGrid(),
        carImages:   make([]*canvas.Image, services.PARKING_CAPACITY),
        maxQueueSize: services.MAX_QUEUE_SIZE,
    }
    scene.setupUI()

    // Establece el tamaño inicial y fija el tamaño de la ventana
    scene.window.Resize(fyne.NewSize(100, 168))
    scene.window.SetFixedSize(true)

    return scene
}

func (s *ParkingScene) setupUI() {
    s.window.SetTitle("Parking Game Simulator")
    s.startButton = widget.NewButtonWithIcon("Iniciar", theme.MediaPlayIcon(), s.handleStart)
    s.stopButton = widget.NewButtonWithIcon("Detener", theme.MediaStopIcon(), s.handleStop)
    s.stopButton.Disable()
    s.statsContainer = container.NewVBox(
        widget.NewLabelWithStyle("🎮", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Monospace: true}),
        widget.NewSeparator(),
    )
    s.setupParkingLot()
    s.queueBox = container.NewHBox()
    queueLabel := widget.NewLabelWithStyle("🚗 Cola de Espera", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
    queueContainer := container.NewVBox(queueLabel, s.queueBox)
    controls := container.NewHBox(
        s.startButton,
        s.stopButton,
        widget.NewButtonWithIcon("Limpiar Log", theme.DeleteIcon(), func() {
            s.logBox.SetText("")
        }),
    )
    infoPanel := container.NewVBox(
        s.createInfoHeader(),
        widget.NewSeparator(),
        s.spacesLabel,
    )
    gameArea := container.NewVBox(
        infoPanel,
        widget.NewSeparator(),
        s.gameContainer,
        widget.NewSeparator(),
        controls,
    )
    rightPanel := container.NewVBox(
        s.statsContainer,
        widget.NewSeparator(),
        queueContainer,
        widget.NewSeparator(),
        container.NewScroll(s.logBox),
    )
    mainContainer := container.NewHSplit(
        gameArea,
        rightPanel,
    )
    mainContainer.SetOffset(0.7)
    s.window.SetContent(mainContainer)
    s.simulation = services.NewSimulation(s.updateUI)
    s.simulation.SetQueueUpdateCallback(s.updateQueueVisual)
}

func (s *ParkingScene) updateQueueVisual(queueSize int) {
    s.queueBox.Objects = nil
    s.queueIcons = []*canvas.Rectangle{}
    for i := 0; i < s.maxQueueSize; i++ {
        carContainer := container.NewVBox()
        var car *canvas.Rectangle
        if i < queueSize {
            car = canvas.NewRectangle(color.RGBA{0, 100, 255, 255})
        } else {
            car = canvas.NewRectangle(color.RGBA{80, 80, 80, 255})
        }
        car.SetMinSize(fyne.NewSize(40, 60))
        carNumber := canvas.NewText(fmt.Sprintf("%d", i+1), color.White)
        carNumber.TextSize = 16
        carNumber.TextStyle = fyne.TextStyle{Bold: true}
        carStack := container.NewStack(car, carNumber)
        carContainer.Add(carStack)
        s.queueIcons = append(s.queueIcons, car)
        s.queueBox.Add(carContainer)
    }
    s.queueBox.Refresh()
}

func (s *ParkingScene) createInfoHeader() fyne.CanvasObject {
    title := canvas.NewText("🎮 Simulador de Estacionamiento", color.White)
    title.TextSize = 24
    title.TextStyle = fyne.TextStyle{Bold: true}
    return container.NewVBox(
        container.NewCenter(title),
    )
}

func (s *ParkingScene) setupParkingLot() {
    s.gameContainer = container.NewVBox()
    background := canvas.NewRectangle(color.RGBA{40, 40, 40, 255})
    background.SetMinSize(fyne.NewSize(600, 400))
    parkingContainer := container.NewGridWithColumns(5)
    s.spaceIcons = make([]*canvas.Rectangle, services.PARKING_CAPACITY)
    for i := 0; i < services.PARKING_CAPACITY; i++ {
        space := canvas.NewRectangle(color.RGBA{50, 50, 50, 255})
        space.SetMinSize(fyne.NewSize(50, 100))
        s.spaceIcons[i] = space
        spaceNum := canvas.NewText(fmt.Sprintf("P%d", i+1), color.White)
        spaceNum.TextSize = 20
        spaceNum.TextStyle = fyne.TextStyle{Bold: true}
        spaceContainer := container.NewStack(
            space,
            container.NewPadded(spaceNum),
        )
        parkingContainer.Add(spaceContainer)
    }
    road := s.createRoad()
    s.gameContainer.Add(parkingContainer)
    s.gameContainer.Add(road)
}

func (s *ParkingScene) createRoad() fyne.CanvasObject {
    road := canvas.NewRectangle(color.RGBA{80, 80, 80, 255})
    road.SetMinSize(fyne.NewSize(600, 40))
    lines := container.NewHBox()
    for i := 0; i < 10; i++ {
        line := canvas.NewRectangle(color.White)
        line.SetMinSize(fyne.NewSize(30, 5))
        lines.Add(line)
    }
    return container.NewStack(road, lines)
}


func (s *ParkingScene) handleStart() {
    s.startButton.Disable()
    s.stopButton.Enable()
    go s.simulation.Start()
}

func (s *ParkingScene) handleStop() {
    s.stopButton.Disable()
    s.startButton.Enable()
    s.simulation.Stop()
}

func (s *ParkingScene) updateUI(spaces int, message string) {
    s.spacesLabel.SetText(fmt.Sprintf("🅿️ Espacios disponibles: %d", spaces))
    s.logBox.SetText(s.logBox.Text() + "\n" + message)
    for i, space := range s.spaceIcons {
        if i < services.PARKING_CAPACITY-spaces {
            space.FillColor = color.RGBA{R: 200, G: 50, B: 50, A: 255}
        } else {
            space.FillColor = color.RGBA{R: 50, G: 150, B: 50, A: 255}
        }
        space.Refresh()
    }
    s.updateQueueBasedOnSpaces(services.PARKING_CAPACITY - spaces)
}

func (s *ParkingScene) updateQueueBasedOnSpaces(occupiedSpaces int) {
    s.queueBox.Objects = nil
    s.queueIcons = []*canvas.Rectangle{}
    queueLength := 0
    if occupiedSpaces >= services.PARKING_CAPACITY {
        queueLength = occupiedSpaces - services.PARKING_CAPACITY + 1
    }
    displayQueueLength := queueLength
    if queueLength > s.maxQueueSize {
        displayQueueLength = s.maxQueueSize
    }
    for i := 0; i < displayQueueLength; i++ {
        carContainer := container.NewVBox()
        car := canvas.NewRectangle(color.RGBA{0, 255, 0, 255})
        car.SetMinSize(fyne.NewSize(40, 60))
        carContainer.Add(car)
        s.queueBox.Add(carContainer)
    }
    s.queueBox.Refresh()
}
