# Simulador de Estacionamiento con Fyne

Un simulador interactivo que modela el comportamiento de un estacionamiento utilizando el framework gráfico **Fyne** y técnicas de concurrencia en **Go**. Este proyecto es ideal para aprender cómo gestionar procesos concurrentes mientras se desarrollan aplicaciones gráficas.

## Comenzando 🚀

Estas instrucciones te permitirán obtener una copia del proyecto en funcionamiento en tu máquina local para propósitos de desarrollo y pruebas.

Mira **Despliegue** para conocer cómo desplegar el proyecto.

### Pre-requisitos 📋

Necesitarás tener instalado **Go** y el framework **Fyne**:

```bash
# Instalar Go
sudo apt install golang-go

# Instalar Fyne
go get fyne.io/fyne/v2
```

### Instalación 🔧

1. Clona este repositorio:
   ```bash
   git clone https://github.com/tuusuario/simuladorEstacionamiento.git
   cd simuladorEstacionamiento
   ```

2. Ejecuta el proyecto:
   ```bash
   go run main.go
   ```

3. Interactúa con la interfaz gráfica para simular la entrada y salida de vehículos.



### **1. Goroutines: Tareas Concurrentes**

En Go, las **goroutines** permiten ejecutar funciones de forma concurrente. En este simulador, se usaron para manejar las operaciones de ingreso, permanencia y salida de vehículos.

Ejemplo:  
Cada vehículo se procesa en una goroutine para simular su entrada al estacionamiento sin bloquear otras operaciones.

```go
func simulateVehicleEntry(vehicleID int) {
    go func() {
        fmt.Printf("Vehículo %d está entrando...\n", vehicleID)
        time.Sleep(2 * time.Second) // Simula el tiempo de entrada.
        fmt.Printf("Vehículo %d ha ingresado.\n", vehicleID)
    }()
}
```

* **Conceptos Clave**  
  - Usar **goroutines** para ejecutar funciones concurrentemente.
  - Manejar múltiples tareas de manera eficiente sin bloquear el programa principal.

---

### **2. `sync.WaitGroup`: Coordinación de Tareas Concurrentes**

El paquete `sync.WaitGroup` se utilizó para esperar a que todas las goroutines terminen antes de proceder. Esto es útil para tareas que deben completarse antes de continuar, como el cierre del estacionamiento.

Ejemplo:
```go
var wg sync.WaitGroup

func simulateVehicle(vehicleID int) {
    defer wg.Done() // Marca esta goroutine como completada al final.
    fmt.Printf("Vehículo %d está estacionando...\n", vehicleID)
    time.Sleep(3 * time.Second) // Simula tiempo de estacionamiento.
    fmt.Printf("Vehículo %d ha salido.\n", vehicleID)
}

func main() {
    for i := 1; i <= 5; i++ {
        wg.Add(1) // Añade una tarea al contador.
        go simulateVehicle(i)
    }

    wg.Wait() // Espera a que todas las tareas terminen.
    fmt.Println("Todos los vehículos han salido.")
}
```

* **Conceptos Clave**  
  - Usar `WaitGroup` para sincronizar goroutines.
  - Evitar que el programa termine antes de completar tareas concurrentes.

---

### **3. `semaphore.Weighted`: Control de Acceso Concurrente**

El paquete `semaphore.Weighted` de `golang.org/x/sync` se usó para limitar el número de vehículos que pueden ingresar al estacionamiento al mismo tiempo, respetando la capacidad máxima.

Ejemplo:
```go
import "golang.org/x/sync/semaphore"

var parkingSemaphore = semaphore.NewWeighted(20) // Capacidad máxima: 20 vehículos.

func enterParking(vehicleID int) {
    if err := parkingSemaphore.Acquire(context.Background(), 1); err != nil {
        fmt.Printf("Error: Vehículo %d no pudo entrar.\n", vehicleID)
        return
    }

    fmt.Printf("Vehículo %d ha entrado al estacionamiento.\n", vehicleID)
    time.Sleep(2 * time.Second) // Simula tiempo dentro del estacionamiento.

    parkingSemaphore.Release(1) // Libera un espacio.
    fmt.Printf("Vehículo %d ha salido del estacionamiento.\n", vehicleID)
}
```

* **Conceptos Clave**  
  - Usar semáforos para limitar el acceso concurrente a recursos compartidos.
  - Manejar el ingreso y salida de recursos de forma controlada.

---

### **4. Canalización para Comunicación entre Goroutines**

Se utilizaron canales (`chan`) para enviar mensajes entre goroutines, como actualizar la interfaz gráfica cuando cambia el estado del estacionamiento.

Ejemplo:
```go
statusChan := make(chan string)

func monitorStatus() {
    for status := range statusChan {
        fmt.Printf("Estado del estacionamiento: %s\n", status)
    }
}

func updateStatus(newStatus string) {
    statusChan <- newStatus // Envía el nuevo estado al canal.
}

func main() {
    go monitorStatus()

    updateStatus("Estacionamiento lleno")
    time.Sleep(1 * time.Second)
    updateStatus("Espacios disponibles")
    close(statusChan) // Cierra el canal al finalizar.
}
```

* **Conceptos Clave**  
  - Usar canales para la comunicación entre goroutines.
  - Implementar patrones de productor-consumidor en Go.

---

### **5. Concurrencia y Actualización de Interfaz**

En el simulador, los datos del estacionamiento (como el número de vehículos estacionados) se actualizan concurrentemente desde goroutines y se reflejan en la interfaz gráfica mediante bindings de **Fyne**.

Ejemplo:
```go
var vehicleCount = binding.NewInt()

func updateVehicleCount(delta int) {
    vehicleCount.Add(delta) // Actualiza el contador de vehículos de forma concurrente.
}

func createCounterLabel() *fyne.Container {
    label := widget.NewLabelWithData(binding.IntToString(vehicleCount))
    return container.NewVBox(label)
}

func simulateEntry(vehicleID int) {
    go func() {
        updateVehicleCount(1)
        time.Sleep(3 * time.Second)
        updateVehicleCount(-1)
    }()
}
```

* **Conceptos Clave**  
  - Integrar bindings de Fyne con datos actualizados concurrentemente.
  - Sincronizar datos de concurrencia con la interfaz gráfica.


---

### **2.Funciones Clave de Fyne Utilizadas en el Proyecto**

**1. `binding.String` para Datos Dinámicos**  
La función `binding.String` permite conectar datos a widgets gráficos para que la interfaz se actualice automáticamente cuando los datos cambien. Esto es útil para mostrar información en tiempo real, como el estado del estacionamiento.

Ejemplo de uso:
```go
import "fyne.io/fyne/v2/data/binding"

var parkingStatus = binding.NewString()

func updateStatus(status string) {
    parkingStatus.Set(status) // Cambia el valor y actualiza automáticamente los widgets vinculados.
}

func createStatusLabel() *fyne.Container {
    label := widget.NewLabelWithData(parkingStatus) // Crea un widget conectado al binding.
    return container.NewVBox(label)
}
```

* **Conceptos Clave**  
  - Usar `binding.String` para sincronizar datos con la interfaz gráfica.
  - Diseñar interfaces que reflejen cambios dinámicos sin necesidad de recargar manualmente.

---

**2. `widget.NewLabelWithData` para Widgets Vinculados**  
Esta función crea un `Label` que muestra el contenido de un binding. Es ideal para mostrar información que cambia frecuentemente, como estadísticas o estados.

Ejemplo:
```go
label := widget.NewLabelWithData(parkingStatus)
```

* **Conceptos Clave**  
  - Vincular widgets a datos dinámicos para interfaces más interactivas.
  - Integrar bindings con otros widgets para mejorar la experiencia de usuario.

---

**3. `container.NewVBox` para Diseño en Columnas**  
`container.NewVBox` organiza los widgets en una columna vertical. Es una herramienta básica y flexible para diseñar interfaces ordenadas.

Ejemplo:
```go
container := container.NewVBox(widget.NewLabel("Estado del Estacionamiento"), label)
```

* **Conceptos Clave**  
  - Diseñar interfaces ordenadas con contenedores predefinidos.
  - Usar diferentes contenedores (`HBox`, `VBox`, `Grid`) según la disposición deseada.

---

**4. Animación y Movimiento de Widgets**  
Aunque Fyne no tiene soporte nativo para animaciones avanzadas, es posible crear efectos visuales básicos modificando la posición de los widgets en bucles.

Ejemplo de movimiento de un vehículo:
```go
func animateVehicle(entry *fyne.Container, start fyne.Position, end fyne.Position) {
    steps := 100
    for i := 0; i <= steps; i++ {
        time.Sleep(10 * time.Millisecond) // Control de velocidad
        x := start.X + (end.X-start.X)*float32(i)/float32(steps)
        y := start.Y + (end.Y-start.Y)*float32(i)/float32(steps)
        entry.Move(fyne.NewPos(x, y))
    }
}
```

* **Conceptos Clave**  
  - Crear animaciones básicas manipulando posiciones de widgets.
  - Usar bucles y temporizadores para simular movimiento fluido.

---

**5. `canvas.NewRectangle` para Representación Gráfica**  
El componente `canvas.NewRectangle` permite dibujar formas simples como rectángulos para representar visualmente elementos del sistema, como los vehículos.

Ejemplo:
```go
rect := canvas.NewRectangle(color.NRGBA{R: 100, G: 200, B: 150, A: 255})
rect.Resize(fyne.NewSize(50, 30)) // Tamaño del rectángulo.
rect.Move(fyne.NewPos(100, 200)) // Posición inicial.
```

* **Conceptos Clave**  
  - Crear gráficos básicos para representar datos o elementos.
  - Integrar formas en contenedores para interfaces personalizadas.

---

**6. Ventana Principal y Contenedores con `App.New`**  
`App.New()` inicia la aplicación principal y crea una ventana que contiene toda la interfaz.

Ejemplo:
```go
func main() {
    app := app.New()
    window := app.NewWindow("Simulador de Estacionamiento")
    
    window.SetContent(createStatusLabel()) // Añade los componentes a la ventana.
    window.Resize(fyne.NewSize(800, 600))  // Tamaño de la ventana.
    window.ShowAndRun() // Inicia la interfaz gráfica.
}
```

* **Conceptos Clave**  
  - Configurar aplicaciones gráficas con **Fyne**.
  - Diseñar interfaces iniciales y gestionar la ventana principal.

---
## Ejecutando las pruebas ⚙️

### Pruebas de concurrencia

```go
func TestVehicleEntry(t *testing.T) {
    ctx := context.Background()
    sem := semaphore.NewWeighted(2) // Prueba con 2 espacios.
    go handleVehicleEntry(1)
    go handleVehicleEntry(2)
    go handleVehicleEntry(3) // Este debería esperar hasta que haya espacio disponible.
    time.Sleep(5 * time.Second)
    if sem.TryAcquire(1) {
        t.Errorf("El semáforo no limitó correctamente la concurrencia")
    }
}
```
---
### Resumen de Aprendizajes

Este proyecto demuestra cómo:

1. Ejecutar tareas concurrentemente con **goroutines**.
2. Sincronizar tareas usando `sync.WaitGroup`.
3. Limitar el acceso a recursos compartidos con `semaphore.Weighted`.
4. Comunicar goroutines con canales (`chan`).
5. Actualizar interfaces gráficas en tiempo real mediante bindings y datos concurrentes.

- La combinación de estas herramientas y conceptos de concurrencia hace que Go sea poderoso para construir simulaciones como esta.
---

## Construido con 🛠️

* [Fyne](https://fyne.io/) - Framework gráfico.
* [Go](https://go.dev/) - Lenguaje de programación.

## Autores ✒️

* **Isaac Toledo** - *Desarrollo del proyecto* - [IsaacToledo123](https://github.com/IsaacToledo123)

## Licencia 📄

Este proyecto está bajo la Licencia MIT - mira el archivo [LICENSE.md](LICENSE.md) para detalles.

---
## Expresiones de Gratitud 🎁

- A la comunidad de Fyne por su framework poderoso.  
- A los entusiastas de la programación concurrente.  
