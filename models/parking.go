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
    mu             sync.Mutex                 
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

    if p.occupiedSpaces >= p.Capacity {
        p.waitingQueue = append(p.waitingQueue, vehicle)
        return false
    }

    if !p.spaceSem.TryAcquire(1) {
        p.waitingQueue = append(p.waitingQueue, vehicle)
        return false
    }

    err := p.gateSem.Acquire(p.ctx, 1)
    if err != nil {
        p.spaceSem.Release(1) 
        return false
    }

    vehicle.SetState(Entering) 
    p.vehicles[vehicle.ID] = vehicle 
    p.occupiedSpaces++ 
    
    spaces := p.GetAvailableSpaces()
    message := fmt.Sprintf("%s ha entrado. Espacios disponibles: %d", vehicle, spaces)
    p.UpdateUI(int(spaces), message) 
    p.gateSem.Release(1)
    vehicle.SetState(Parked) 

    return true
}

func (p *ParkingLot) Exit(vehicle *Vehicle) {
    p.mu.Lock()        
    defer p.mu.Unlock() 

    if _, exists := p.vehicles[vehicle.ID]; !exists {
        return 
    }

    err := p.gateSem.Acquire(p.ctx, 1)
    if err != nil {
        return 
    }

    vehicle.SetState(Exiting) 
    delete(p.vehicles, vehicle.ID) 
    p.occupiedSpaces-- 
    
    availableSpaces := p.GetAvailableSpaces()
    message := fmt.Sprintf("%s ha salido. Espacios disponibles: %d", vehicle, availableSpaces)
    p.UpdateUI(int(availableSpaces), message)

    p.spaceSem.Release(1)

    if len(p.waitingQueue) > 0 {
        nextVehicle := p.waitingQueue[0]
        p.waitingQueue = p.waitingQueue[1:] 
        
        p.mu.Unlock()
        go p.TryEnter(nextVehicle)
        p.mu.Lock() 
    }

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
    
    queueCopy := make([]*Vehicle, len(p.waitingQueue))
    copy(queueCopy, p.waitingQueue)
    return queueCopy
}
