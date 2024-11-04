package services

import (
    "math/rand"
    "sync"
    "time"
    "context"
    "holafyne/models"
    "holafyne/utils"
)

const (
    PARKING_CAPACITY = 20
    MAX_VEHICLES     = 100
    MIN_PARK_TIME    = 10
    MAX_PARK_TIME    = 20
    MAX_QUEUE_SIZE   = 80  
)

type SimulationConfig struct {
    ParkingCapacity int
    MaxVehicles     int
    MinParkTime     float64
    MaxParkTime     float64
    ArrivalRate     float64
}

type Simulation struct {
    config       SimulationConfig
    parking      *models.ParkingLot
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
    poissonGen   *utils.PoissonGenerator
    stats        *SimulationStats
    statsMutex   sync.RWMutex
    queue        []*models.Vehicle
    queueMutex   sync.RWMutex
    onQueueUpdate func(queueSize int)  // Nuevo: callback para actualizar la UI
}
func (s *Simulation) SetQueueUpdateCallback(callback func(queueSize int)) {
    s.onQueueUpdate = callback
}
type SimulationStats struct {
    ProcessedVehicles  int
    TotalWaitingTime   time.Duration
    MaxWaitingTime     time.Duration
    TotalParkingTime   time.Duration
    MaxQueueLength     int
    CurrentQueueLength int
}

func DefaultConfig() SimulationConfig {
    return SimulationConfig{
        ParkingCapacity: PARKING_CAPACITY,
        MaxVehicles:     MAX_VEHICLES,
        MinParkTime:     MIN_PARK_TIME,
        MaxParkTime:     MAX_PARK_TIME,
        ArrivalRate:     2.0,
    }
}

func NewSimulation(updateUI func(spaces int, message string)) *Simulation {
    return NewSimulationWithConfig(DefaultConfig(), updateUI)
}
func NewSimulationWithConfig(config SimulationConfig, updateUI func(spaces int, message string)) *Simulation {
    ctx, cancel := context.WithCancel(context.Background())
    
    poissonConfig := utils.DefaultPoissonConfig()
    poissonConfig.Lambda = config.ArrivalRate
    
    return &Simulation{
        config:     config,
        parking:    models.NewParkingLot(config.ParkingCapacity, updateUI),
        ctx:        ctx,
        cancel:     cancel,
        poissonGen: utils.NewPoissonGenerator(poissonConfig),
        stats:      &SimulationStats{},
        queue:      make([]*models.Vehicle, 0, MAX_QUEUE_SIZE),
    }
}

func (s *Simulation) Start() {
    s.wg.Add(1)
    go s.runSimulation()
    go s.processQueue() 
}

func (s *Simulation) Stop() {
    s.cancel()
    s.wg.Wait()
}

func (s *Simulation) processQueue() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.tryProcessNextInQueue()
        }
    }
}

func (s *Simulation) tryProcessNextInQueue() {
    s.queueMutex.Lock()
    if len(s.queue) > 0 && s.parking.GetAvailableSpaces() > 0 {
        vehicle := s.queue[0]
        s.queue = s.queue[1:]
        s.queueMutex.Unlock()
        
        s.wg.Add(1)
        go s.processVehicle(vehicle)
    } else {
        s.queueMutex.Unlock()
    }
}

func (s *Simulation) runSimulation() {
    defer s.wg.Done()
    
    vehicleCount := 0
    for vehicleCount < s.config.MaxVehicles {
        select {
        case <-s.ctx.Done():
            return
        default:
            vehicleCount++
            vehicle := models.NewVehicle(vehicleCount)
            
            if s.parking.GetAvailableSpaces() > 0 {
                s.wg.Add(1)
                go s.processVehicle(vehicle)
            } else {
                s.addToQueue(vehicle)
            }
            
            interval := s.poissonGen.NextInterval()
            select {
            case <-s.ctx.Done():
                return
            case <-time.After(interval):
                continue
            }
        }
    }
}

func (s *Simulation) addToQueue(vehicle *models.Vehicle) bool {
    s.queueMutex.Lock()
    defer s.queueMutex.Unlock()

    if len(s.queue) >= MAX_QUEUE_SIZE {
        return false
    }

    s.queue = append(s.queue, vehicle)
    queueLength := len(s.queue)

    s.statsMutex.Lock()
    if queueLength > s.stats.MaxQueueLength {
        s.stats.MaxQueueLength = queueLength
    }
    s.stats.CurrentQueueLength = queueLength
    s.statsMutex.Unlock()

    // Notificar cambio en la cola
    if s.onQueueUpdate != nil {
        s.onQueueUpdate(queueLength)
    }

    return true
}

func (s *Simulation) processVehicle(vehicle *models.Vehicle) {
    defer s.wg.Done()
    
    startWait := time.Now()
    entered := s.parking.TryEnter(vehicle)
    
    if !entered {
        if !s.addToQueue(vehicle) {
            // Si la cola está llena, el vehículo se va
            return
        }
        return
    }
    waitingTime := time.Since(startWait)
    s.updateWaitingStats(waitingTime)
    
    parkTime := s.generateParkingTime()
    timer := time.NewTimer(parkTime)
    
    select {
    case <-s.ctx.Done():
        timer.Stop()
        s.parking.Exit(vehicle)
        return
    case <-timer.C:
        s.parking.Exit(vehicle)
        s.updateParkingStats(parkTime)
    }
}

func (s *Simulation) GetQueueLength() int {
    s.queueMutex.RLock()
    defer s.queueMutex.RUnlock()
    return len(s.queue)
}

func (s *Simulation) generateParkingTime() time.Duration {
    parkTime := s.config.MinParkTime + rand.Float64()*(s.config.MaxParkTime-s.config.MinParkTime)
    return time.Duration(parkTime * float64(time.Second))
}

func (s *Simulation) updateWaitingStats(waitTime time.Duration) {
    s.statsMutex.Lock()
    defer s.statsMutex.Unlock()
    
    s.stats.ProcessedVehicles++
    s.stats.TotalWaitingTime += waitTime
    if waitTime > s.stats.MaxWaitingTime {
        s.stats.MaxWaitingTime = waitTime
    }
}

func (s *Simulation) updateParkingStats(parkTime time.Duration) {
    s.statsMutex.Lock()
    defer s.statsMutex.Unlock()
    
    s.stats.TotalParkingTime += parkTime
}

func (s *Simulation) GetStats() SimulationStats {
    s.statsMutex.RLock()
    defer s.statsMutex.RUnlock()
    
    return *s.stats
}

func (s *Simulation) GetAverageWaitTime() time.Duration {
    s.statsMutex.RLock()
    defer s.statsMutex.RUnlock()
    
    if s.stats.ProcessedVehicles == 0 {
        return 0
    }
    return s.stats.TotalWaitingTime / time.Duration(s.stats.ProcessedVehicles)
}

func (s *Simulation) GetAverageParkTime() time.Duration {
    s.statsMutex.RLock()
    defer s.statsMutex.RUnlock()
    
    if s.stats.ProcessedVehicles == 0 {
        return 0
    }
    return s.stats.TotalParkingTime / time.Duration(s.stats.ProcessedVehicles)
}