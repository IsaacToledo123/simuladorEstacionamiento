package models

import (
    "fmt"
    "sync"
    "time"
)

type VehicleState int

const (
    Waiting VehicleState = iota
    Entering
    Parked
    Exiting
)

type Vehicle struct {
    ID        int
    state     VehicleState
    EntryTime time.Time
    ExitTime  time.Time
    mu        sync.RWMutex 
}

var stateStrings = map[VehicleState]string{
    Waiting:  "esperando",
    Entering: "entrando",
    Parked:   "estacionado",
    Exiting:  "saliendo",
}

func NewVehicle(id int) *Vehicle {
    return &Vehicle{
        ID:        id,
        state:     Waiting,
        EntryTime: time.Now(),
    }
}

func (v *Vehicle) String() string {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return fmt.Sprintf("Vehículo %d [%s]", v.ID, stateStrings[v.state])
}

func (v *Vehicle) SetState(state VehicleState) {
    v.mu.Lock()
    defer v.mu.Unlock()
    
    v.state = state
    if state == Entering && v.EntryTime.IsZero() {
        v.EntryTime = time.Now()
    } else if state == Exiting {
        v.ExitTime = time.Now()
    }
}

func (v *Vehicle) GetState() VehicleState {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.state
}

func (v *Vehicle) GetParkingDuration() time.Duration {
    v.mu.RLock()
    defer v.mu.RUnlock()
    
    if v.state == Exiting || v.ExitTime.After(v.EntryTime) {
        return v.ExitTime.Sub(v.EntryTime)
    }
    return time.Since(v.EntryTime)
}

func (v *Vehicle) IsParked() bool {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.state == Parked
}

func (v *Vehicle) GetEntryTime() time.Time {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.EntryTime
}

func (v *Vehicle) GetExitTime() time.Time {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.ExitTime
}

func (v *Vehicle) GetStateString() string {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return stateStrings[v.state]
}