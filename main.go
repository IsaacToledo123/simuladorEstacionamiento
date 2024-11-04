package main

import (
    "holafyne/scenes"
    "fyne.io/fyne/v2/app"
)

func main() {
    myApp := app.New()
    window := myApp.NewWindow("Simulador de Estacionamiento")
    
    scenes.NewParkingScene(window)
    
    window.ShowAndRun()
}