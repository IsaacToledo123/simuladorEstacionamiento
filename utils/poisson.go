package utils

import (
    "math"
    "math/rand"
    "sync"
    "time"
)

// PoissonGenerator genera intervalos y eventos siguiendo una distribución de Poisson
type PoissonGenerator struct {
    lambda     float64    // tasa media de ocurrencia
    minTime    float64    // tiempo mínimo entre eventos (en segundos)
    maxTime    float64    // tiempo máximo entre eventos (en segundos)
    rng        *rand.Rand // generador de números aleatorios
    mu         sync.Mutex // protege el acceso concurrente
}

// PoissonConfig contiene la configuración para el generador de Poisson
type PoissonConfig struct {
    Lambda     float64 // tasa media de ocurrencia
    MinTime    float64 // tiempo mínimo entre eventos (en segundos)
    MaxTime    float64 // tiempo máximo entre eventos (en segundos)
    RandomSeed int64   // semilla para el generador de números aleatorios
}

// DefaultPoissonConfig retorna una configuración por defecto
func DefaultPoissonConfig() PoissonConfig {
    return PoissonConfig{
        Lambda:     2.0,   // 2 eventos por segundo en promedio
        MinTime:    0.1,   // mínimo 0.1 segundos entre eventos
        MaxTime:    10.0,  // máximo 10 segundos entre eventos
        RandomSeed: time.Now().UnixNano(),
    }
}

// NewPoissonGenerator crea un nuevo generador con la configuración proporcionada
func NewPoissonGenerator(config PoissonConfig) *PoissonGenerator {
    return &PoissonGenerator{
        lambda:     config.Lambda,
        minTime:    config.MinTime,
        maxTime:    config.MaxTime,
        rng:        rand.New(rand.NewSource(config.RandomSeed)),
    }
}

// NewPoissonGeneratorWithLambda crea un nuevo generador solo con lambda (para compatibilidad)
func NewPoissonGeneratorWithLambda(lambda float64) *PoissonGenerator {
    config := DefaultPoissonConfig()
    config.Lambda = lambda
    return NewPoissonGenerator(config)
}

// NextInterval genera el siguiente intervalo de tiempo siguiendo la distribución de Poisson
func (pg *PoissonGenerator) NextInterval() time.Duration {
    pg.mu.Lock()
    defer pg.mu.Unlock()

    // Genera un intervalo usando la distribución exponencial inversa
    u := pg.rng.Float64()
    x := -math.Log(1.0-u) / pg.lambda

    // Limita el intervalo al rango especificado
    x = math.Max(pg.minTime, math.Min(pg.maxTime, x))

    return time.Duration(x * float64(time.Second))
}

// NextEvents genera el número de eventos que ocurrirían en una duración dada
func (pg *PoissonGenerator) NextEvents(duration time.Duration) int {
    pg.mu.Lock()
    defer pg.mu.Unlock()

    // Calcula lambda ajustado para la duración
    adjustedLambda := pg.lambda * duration.Seconds()
    
    // Usa el algoritmo de Knuth para generar variables aleatorias de Poisson
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

// GenerateEventTimes genera una lista de tiempos de eventos en una duración dada
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

// SetLambda actualiza la tasa media de ocurrencia
func (pg *PoissonGenerator) SetLambda(lambda float64) {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    pg.lambda = lambda
}

// SetTimeConstraints actualiza los límites de tiempo
func (pg *PoissonGenerator) SetTimeConstraints(minTime, maxTime float64) {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    pg.minTime = minTime
    pg.maxTime = maxTime
}

// GetLambda retorna la tasa media actual
func (pg *PoissonGenerator) GetLambda() float64 {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    return pg.lambda
}

// GetTimeConstraints retorna los límites de tiempo actuales
func (pg *PoissonGenerator) GetTimeConstraints() (float64, float64) {
    pg.mu.Lock()
    defer pg.mu.Unlock()
    return pg.minTime, pg.maxTime
}