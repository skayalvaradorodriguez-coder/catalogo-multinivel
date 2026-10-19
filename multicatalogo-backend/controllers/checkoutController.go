package controllers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/notificaciones"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/validacion"

	"github.com/gofiber/fiber/v2"
)

// metodosPago son los mismos que ofrece Checkout.jsx.
var metodosPago = map[string]bool{
	"Tarjeta de crédito":      true,
	"Transferencia bancaria":  true,
	"Efectivo contra entrega": true,
}

// CheckoutController procesa las compras.
type CheckoutController struct {
	base
	repo   repository.DBRepository
	cache  cache.CacheRepository
	mailer notificaciones.Mailer
	// pendientes cuenta las goroutines en curso, para no cortarlas al apagar el servidor.
	pendientes sync.WaitGroup
}

// Checkout: POST /api/checkout (requiere token). El comprador sale del TOKEN, no del cuerpo.
// Guarda pedido + detalle + ledger de comisiones en una transacción ACID y responde de inmediato;
// el contador de ventas y la caché se actualizan después en una goroutine.
func (ctl *CheckoutController) Checkout(c *fiber.Ctx) error {
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}

	// 1. Validaciones antes de abrir la transacción.
	if mensaje, detalles := validarCheckout(&req); mensaje != "" {
		if detalles == nil {
			return ctl.errorJSON(c, fiber.StatusBadRequest, mensaje, nil)
		}
		return ctl.errorJSON(c, fiber.StatusBadRequest, mensaje, detalles)
	}

	ctx := c.Context()
	usuario, err := ctl.repo.BuscarUsuarioPorID(ctx, usuarioToken(c).UsuarioID())
	if err != nil {
		return ctl.errorRepo(c, err, "Usuario no encontrado")
	}
	if !usuario.Activo {
		return ctl.errorJSON(c, fiber.StatusForbidden, "Tu cuenta está desactivada", nil)
	}

	// 2. Transacción ACID: BEGIN → pedidos → pedido_items + stock → comisiones → COMMIT (o ROLLBACK).
	pedido, err := ctl.repo.CrearPedido(ctx, models.NuevoPedido{
		UsuarioID:  usuario.ID,
		Cliente:    req.Cliente,
		MetodoPago: req.MetodoPago,
		Items:      req.Items,
	})
	if err != nil {
		if _, esValidacion := err.(*repository.ErrValidacion); esValidacion || err == repository.ErrCantidadInvalida {
			return ctl.errorRepo(c, err, "")
		}
		log.Printf("Checkout fallido, se aplicó ROLLBACK: %v", err)
		return ctl.errorJSON(c, fiber.StatusInternalServerError,
			"No se pudo procesar el pedido. No se realizó ningún cobro; intenta de nuevo.", err.Error())
	}

	// El stock cambió: el catálogo en caché ya no es válido.
	ctl.cache.Eliminar(ctx, cache.ClaveProductos)

	// 3. COMMIT exitoso: lo no crítico va en segundo plano.
	ctl.pendientes.Add(1)
	go ctl.procesarComisiones(pedido.ID, usuario.ID, pedido.Total, req.Cliente.Email)

	return c.Status(fiber.StatusCreated).JSON(pedido)
}

// procesarComisiones corre en una goroutine con su propio contexto (el de la petición
// deja de ser válido apenas se responde). Actualiza el contador de ventas del comprador,
// invalida los árboles en caché y envía el correo de confirmación.
func (ctl *CheckoutController) procesarComisiones(pedidoID, usuarioID int, total float64, email string) {
	defer ctl.pendientes.Done()
	inicio := time.Now()
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	if err := ctl.repo.SumarVentas(ctx, usuarioID, total); err != nil {
		log.Printf("Goroutine de comisiones: error en el pedido %d: %v", pedidoID, err)
		return
	}
	ctl.cache.EliminarPatron(ctx, cache.PatronRed)
	_ = ctl.mailer.Enviar(email, "Confirmación de pedido",
		fmt.Sprintf("Tu pedido MC-%06d por $%.2f fue recibido", pedidoID, total))

	log.Printf("Goroutine de comisiones: pedido %d procesado en %s; caché %s invalidada",
		pedidoID, time.Since(inicio).Round(time.Millisecond), cache.PatronRed)
}

// EsperarComisiones bloquea hasta que terminen las goroutines en curso (se usa al apagar).
func (ctl *CheckoutController) EsperarComisiones() {
	ctl.pendientes.Wait()
}

// validarCheckout revisa los datos y agrupa cantidades si un producto viene repetido.
// Devuelve el mensaje principal y el detalle por campo ("" si todo está bien).
func validarCheckout(req *models.CheckoutRequest) (string, map[string]string) {
	req.Cliente.Nombre = strings.TrimSpace(req.Cliente.Nombre)
	req.Cliente.Email = strings.TrimSpace(req.Cliente.Email)
	req.Cliente.Direccion = strings.TrimSpace(req.Cliente.Direccion)
	req.Cliente.Ciudad = strings.TrimSpace(req.Cliente.Ciudad)
	req.Cliente.Telefono = strings.TrimSpace(req.Cliente.Telefono)

	v := validacion.Errores{}
	v.Requerido("nombre", req.Cliente.Nombre, "El nombre es obligatorio")
	v.Email("email", req.Cliente.Email)
	v.Requerido("direccion", req.Cliente.Direccion, "La dirección es obligatoria")
	v.Requerido("ciudad", req.Cliente.Ciudad, "La ciudad es obligatoria")
	v.Requerido("telefono", req.Cliente.Telefono, "El teléfono es obligatorio")
	if !metodosPago[req.MetodoPago] {
		v["metodoPago"] = "El método de pago no es válido"
	}
	if len(req.Items) == 0 {
		v["items"] = "El carrito está vacío"
	}
	if errs := v.Validar(); errs != nil {
		return "Los datos del pedido no son válidos", errs
	}

	cantidades := map[int]int{}
	var orden []int
	for _, item := range req.Items {
		if item.Cantidad <= 0 {
			return "La cantidad debe ser mayor que cero", nil
		}
		if item.Cantidad > 100 {
			return "No puedes comprar más de 100 unidades de un producto", nil
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
	return "", nil
}
