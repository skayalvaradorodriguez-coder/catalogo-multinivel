package controllers

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"

	"github.com/gofiber/fiber/v2"
)

// metodosPago son los mismos que ofrece Checkout.jsx.
var metodosPago = map[string]bool{
	"Tarjeta de crédito":      true,
	"Transferencia bancaria":  true,
	"Efectivo contra entrega": true,
}

// CheckoutController procesa las compras. Recibe el repositorio y la caché por inyección de dependencias.
type CheckoutController struct {
	repo  repository.DBRepository
	cache cache.CacheRepository
	// pendientes cuenta las goroutines de comisiones en curso, para no cortarlas al apagar el servidor.
	pendientes sync.WaitGroup
}

// NuevoCheckoutController crea el controlador con sus dependencias.
func NuevoCheckoutController(repo repository.DBRepository, cache cache.CacheRepository) *CheckoutController {
	return &CheckoutController{repo: repo, cache: cache}
}

// Checkout recibe el carrito, lo guarda en una transacción ACID y responde de inmediato.
// Las comisiones de la red se calculan después, en una goroutine.
func (ctl *CheckoutController) Checkout(c *fiber.Ctx) error {
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "Cuerpo de petición inválido", nil))
	}

	// 1. Validaciones antes de abrir la transacción.
	if detalles := validarCheckout(&req); len(detalles) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "Los datos del pedido no son válidos", detalles))
	}

	ctx := c.Context()
	usuario, err := ctl.repo.BuscarUsuarioPorEmail(ctx, req.UsuarioEmail)
	if err != nil {
		return errorInterno(c, err)
	}
	if usuario == nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			models.NewAPIError(fiber.StatusBadRequest, "El usuario no está registrado", req.UsuarioEmail))
	}

	// 2. Transacción ACID: BEGIN → pedidos → pedido_items + stock → COMMIT (o ROLLBACK).
	pedido, err := ctl.repo.CrearPedido(ctx, models.NuevoPedido{
		UsuarioID:  usuario.ID,
		Cliente:    req.Cliente,
		MetodoPago: req.MetodoPago,
		Items:      req.Items,
	})
	if err != nil {
		var errValidacion *repository.ErrValidacion
		if errors.As(err, &errValidacion) {
			return c.Status(fiber.StatusBadRequest).JSON(
				models.NewAPIError(fiber.StatusBadRequest, errValidacion.Mensaje, nil))
		}
		log.Printf("Checkout fallido, se aplicó ROLLBACK: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			models.NewAPIError(fiber.StatusInternalServerError,
				"No se pudo procesar el pedido. No se realizó ningún cobro; intenta de nuevo.", nil))
	}

	// El stock cambió: el catálogo en caché ya no es válido.
	ctl.cache.Eliminar(ctx, cache.ClaveProductos)

	// 3. COMMIT exitoso: las comisiones se calculan en segundo plano.
	ctl.pendientes.Add(1)
	go ctl.calcularComisiones(pedido.ID, usuario.ID, pedido.Total)

	// 4. Respondemos sin esperar a la goroutine.
	return c.Status(fiber.StatusCreated).JSON(pedido)
}

// calcularComisiones corre en una goroutine. Usa su propio contexto porque el de la
// petición HTTP deja de ser válido apenas se envía la respuesta al cliente.
func (ctl *CheckoutController) calcularComisiones(pedidoID, usuarioID int, total float64) {
	defer ctl.pendientes.Done()
	inicio := time.Now()
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	afectados, err := ctl.repo.CalcularComisiones(ctx, pedidoID, usuarioID, total)
	if err != nil {
		log.Printf("Goroutine de comisiones: error en el pedido %d: %v", pedidoID, err)
		return
	}

	// Paso 5: invalidamos el árbol en caché de cada usuario afectado para que su Dashboard se actualice.
	claves := make([]string, 0, len(afectados))
	for _, email := range afectados {
		claves = append(claves, cache.ClaveRed(strings.ToLower(email)))
	}
	ctl.cache.Eliminar(ctx, claves...)

	log.Printf("Goroutine de comisiones: pedido %d procesado en %s; caché invalidada: %v",
		pedidoID, time.Since(inicio).Round(time.Millisecond), claves)
}

// EsperarComisiones bloquea hasta que terminen las goroutines de comisiones en curso.
// main.go la llama al apagar el servidor, antes de cerrar la base de datos.
func (ctl *CheckoutController) EsperarComisiones() {
	ctl.pendientes.Wait()
}

// validarCheckout revisa los datos y agrupa cantidades si un producto viene repetido.
func validarCheckout(req *models.CheckoutRequest) map[string]string {
	detalles := map[string]string{}

	req.UsuarioEmail = strings.ToLower(strings.TrimSpace(req.UsuarioEmail))
	req.Cliente.Nombre = strings.TrimSpace(req.Cliente.Nombre)
	req.Cliente.Email = strings.TrimSpace(req.Cliente.Email)
	req.Cliente.Direccion = strings.TrimSpace(req.Cliente.Direccion)
	req.Cliente.Ciudad = strings.TrimSpace(req.Cliente.Ciudad)
	req.Cliente.Telefono = strings.TrimSpace(req.Cliente.Telefono)

	if req.UsuarioEmail == "" {
		detalles["usuarioEmail"] = "Debes iniciar sesión para comprar"
	}
	if req.Cliente.Nombre == "" {
		detalles["nombre"] = "El nombre es obligatorio"
	}
	if !strings.Contains(req.Cliente.Email, "@") {
		detalles["email"] = "El correo electrónico no es válido"
	}
	if req.Cliente.Direccion == "" {
		detalles["direccion"] = "La dirección es obligatoria"
	}
	if req.Cliente.Ciudad == "" {
		detalles["ciudad"] = "La ciudad es obligatoria"
	}
	if req.Cliente.Telefono == "" {
		detalles["telefono"] = "El teléfono es obligatorio"
	}
	if !metodosPago[req.MetodoPago] {
		detalles["metodoPago"] = "El método de pago no es válido"
	}

	if len(req.Items) == 0 {
		detalles["items"] = "El carrito está vacío"
		return detalles
	}

	cantidades := map[int]int{}
	var orden []int
	for _, item := range req.Items {
		if item.Cantidad <= 0 || item.Cantidad > 100 {
			detalles["items"] = "Cada producto debe tener una cantidad entre 1 y 100"
			return detalles
		}
		if _, existe := cantidades[item.ProductoID]; !existe {
			orden = append(orden, item.ProductoID)
		}
		cantidades[item.ProductoID] += item.Cantidad
	}

	req.Items = req.Items[:0]
	for _, id := range orden {
		req.Items = append(req.Items, models.ItemCheckoutReq{ProductoID: id, Cantidad: cantidades[id]})
	}
	return detalles
}
