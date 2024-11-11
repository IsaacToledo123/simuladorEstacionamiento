package utils

import (
    "math"
    "math/rand"
    "sync"
    "time"
)

type PoissonGenerator struct {
    lambda     float64    
    minTime    float64   
    maxTime    float64    
    rng        *rand.Rand 
    mu         sync.Mutex 
}

type PoissonConfig struct {
    Lambda     float64 
    MinTime    float64 
    MaxTime    float64 
    RandomSeed int64   
}

func DefaultPoissonConfig() PoissonConfig {
    return PoissonConfig{
        Lambda:     2.0,   
        MinTime:    0.1,   
        MaxTime:    10.0,  
        RandomSeed: time.Now().UnixNano(),
    }
}

func NewPoissonGenerator(config PoissonConfig) *PoissonGenerator {
    return &PoissonGenerator{
        lambda:     config.Lambda,
        minTime:    config.MinTime,
        maxTime:    config.MaxTime,
        rng:        rand.New(rand.NewSource(config.RandomSeed)),
    }
}

func NewPoissonGeneratorWithLambda(lambda float64) *PoissonGenerator {
    config := DefaultPoissonConfig()
    config.Lambda = lambda
    return NewPoissonGenerator(config)
}

func (pg *PoissonGenerator) NextInterval() time.Duration {
    pg.mu.Lock()
    defer pg.mu.Unlock()

    u := pg.rng.Float64()
    x := -math.Log(1.0-u) / pg.lambda

    x = math.Max(pg.minTime, math.Min(pg.maxTime, x))

    return time.Duration(x * float64(time.Second))
}

func (pg *PoissonGenerator) NextEvents(duration time.Duration) int {
    pg.mu.Lock()
    defer pg.mu.Unlock()

    adjustedLambda := pg.lambda * duration.Seconds()
    
    L := math.Exp(-adjustedLambda)
    k := 0
    p := 1.0

    for {
        k++
        p *= pg.rng.Float64()
        if p < L {
            return k - 1
        }
    }
}

func (pg *PoissonGenerator) GenerateEventTimes(duration time.Duration) []time.Duration {
    pg.mu.Lock()
    defer pg.mu.Unlock()

    var times []time.Duration
    currentTime := time.Duration(0)

    for currentTime < duration {
        interval := pg.NextInterval()
        currentTime += interval
        if currentTime < duration {
            times = append(times, currentTime)
        }
    }

    return times
}

func (pg *PoissonGenerator) SetLambda(lambda float64) {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    pg.lambda = lambda
}

func (pg *PoissonGenerator) SetTimeConstraints(minTime, maxTime float64) {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    pg.minTime = minTime
    pg.maxTime = maxTime
}

func (pg *PoissonGenerator) GetLambda() float64 {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    return pg.lambda
}

func (pg *PoissonGenerator) GetTimeConstraints() (float64, float64) {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    return pg.minTime, pg.maxTime
}