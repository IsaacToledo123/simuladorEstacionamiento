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
    mu        sync.RWMutex // Mutex para proteger el acceso concurrente
}

// String representations of vehicle states
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

// GetState returns the current state of the vehicle
func (v *Vehicle) GetState() VehicleState {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.state
}

// GetParkingDuration returns the duration the vehicle has been parked
func (v *Vehicle) GetParkingDuration() time.Duration {
    v.mu.RLock()
    defer v.mu.RUnlock()
    
    if v.state == Exiting || v.ExitTime.After(v.EntryTime) {
        return v.ExitTime.Sub(v.EntryTime)
    }
    return time.Since(v.EntryTime)
}

// IsParked returns true if the vehicle is currently parked
func (v *Vehicle) IsParked() bool {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.state == Parked
}

// GetEntryTime returns a copy of the entry time
func (v *Vehicle) GetEntryTime() time.Time {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.EntryTime
}

// GetExitTime returns a copy of the exit time
func (v *Vehicle) GetExitTime() time.Time {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return v.ExitTime
}

// GetStateString returns the string representation of the current state
func (v *Vehicle) GetStateString() string {
    v.mu.RLock()
    defer v.mu.RUnlock()
    return stateStrings[v.state]
}