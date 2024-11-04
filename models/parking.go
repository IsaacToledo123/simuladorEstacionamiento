package models

import (
    "context"
    "fmt"
    "sync"
    "golang.org/x/sync/semaphore"
)

type ParkingLot struct {
    Capacity       int64
    spaceSem       *semaphore.Weighted
    gateSem        *semaphore.Weighted
    vehicles       map[int]*Vehicle
    waitingQueue   []*Vehicle
    occupiedSpaces int64
    UpdateUI       func(spaces int, message string)
    ctx            context.Context
    mu            sync.Mutex // Mutex para proteger el acceso concurrente
}

func NewParkingLot(capacity int, updateUI func(spaces int, message string)) *ParkingLot {
    return &ParkingLot{
        Capacity:       int64(capacity),
        spaceSem:       semaphore.NewWeighted(int64(capacity)),
        gateSem:        semaphore.NewWeighted(1),
        vehicles:       make(map[int]*Vehicle),
        waitingQueue:   []*Vehicle{},
        occupiedSpaces: 0,
        UpdateUI:       updateUI,
        ctx:            context.Background(),
    }
}

func (p *ParkingLot) TryEnter(vehicle *Vehicle) bool {
    p.mu.Lock()
    defer p.mu.Unlock()

    // Verificar si hay espacio disponible
    if p.occupiedSpaces >= p.Capacity {
        // Agregar el vehículo a la cola si no hay espacio
        p.waitingQueue = append(p.waitingQueue, vehicle)
        return false
    }

    // Intentar adquirir un espacio
    if !p.spaceSem.TryAcquire(1) {
        p.waitingQueue = append(p.waitingQueue, vehicle)
        return false
    }

    // Adquirir la puerta
    err := p.gateSem.Acquire(p.ctx, 1)
    if err != nil {
        p.spaceSem.Release(1)
        return false
    }

    // Procesar la entrada del vehículo
    vehicle.SetState(Entering)
    p.vehicles[vehicle.ID] = vehicle
    p.occupiedSpaces++
    spaces := p.GetAvailableSpaces()
    message := fmt.Sprintf("%s ha entrado. Espacios disponibles: %d", vehicle, spaces)
    p.UpdateUI(int(spaces), message)

    // Liberar la puerta
    p.gateSem.Release(1)
    vehicle.SetState(Parked)

    return true
}

func (p *ParkingLot) Exit(vehicle *Vehicle) {
    p.mu.Lock()
    defer p.mu.Unlock()

    // Verificar si el vehículo está en el estacionamiento
    if _, exists := p.vehicles[vehicle.ID]; !exists {
        return
    }

    // Adquirir la puerta para salir
    err := p.gateSem.Acquire(p.ctx, 1)
    if err != nil {
        return
    }

    // Procesar la salida del vehículo
    vehicle.SetState(Exiting)
    delete(p.vehicles, vehicle.ID)
    p.occupiedSpaces--
    
    // Actualizar UI antes de procesar la cola
    availableSpaces := p.GetAvailableSpaces()
    message := fmt.Sprintf("%s ha salido. Espacios disponibles: %d", vehicle, availableSpaces)
    p.UpdateUI(int(availableSpaces), message)

    // Liberar el espacio
    p.spaceSem.Release(1)

    // Procesar la cola de espera
    if len(p.waitingQueue) > 0 {
        nextVehicle := p.waitingQueue[0]
        p.waitingQueue = p.waitingQueue[1:]
        
        // Liberar el mutex antes de intentar ingresar el siguiente vehículo
        p.mu.Unlock()
        go p.TryEnter(nextVehicle)
        p.mu.Lock()
    }

    // Liberar la puerta
    p.gateSem.Release(1)
}

func (p *ParkingLot) GetAvailableSpaces() int64 {
    return p.Capacity - p.occupiedSpaces
}

func (p *ParkingLot) GetOccupancy() int {
    p.mu.Lock()
    defer p.mu.Unlock()
    return int(p.occupiedSpaces)
}

func (p *ParkingLot) GetWaitingVehicles() []*Vehicle {
    p.mu.Lock()
    defer p.mu.Unlock()
    // Retornar una copia de la cola para evitar problemas de concurrencia
    queueCopy := make([]*Vehicle, len(p.waitingQueue))
    copy(queueCopy, p.waitingQueue)
    return queueCopy
}