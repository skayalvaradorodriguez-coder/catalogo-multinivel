package controllers

import (
	"strings"

	"multicatalogo-backend/cache"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/validacion"

	"github.com/gofiber/fiber/v2"
)

// ProductoController maneja el catálogo público y su administración.
type ProductoController struct {
	base
	repo  repository.DBRepository
	cache cache.CacheRepository
}

// GetProductos devuelve el catálogo público (solo activos) con Cache-Aside.
// Con ?q= hace búsqueda server-side (ILIKE) sin caché.
func (ctl *ProductoController) GetProductos(c *fiber.Ctx) error {
	ctx := c.Context()

	if q := strings.TrimSpace(c.Query("q")); q != "" {
		productos, err := ctl.repo.ObtenerProductos(ctx, q)
		if err != nil {
			return ctl.errorInterno(c, err)
		}
		return c.JSON(productos)
	}

	var productos []models.Producto
	if ctl.cache.Obtener(ctx, cache.ClaveProductos, &productos) {
		c.Set("X-Cache", "HIT")
		return c.JSON(productos)
	}

	productos, err := ctl.repo.ObtenerProductos(ctx, "")
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	ctl.cache.Guardar(ctx, cache.ClaveProductos, productos, cache.TTLProductos)
	c.Set("X-Cache", "MISS")
	return c.JSON(productos)
}

// GetProductoPorID devuelve un producto activo (404 si no existe o está desactivado).
func (ctl *ProductoController) GetProductoPorID(c *fiber.Ctx) error {
	id, err := ctl.paramID(c)
	if err != nil {
		return err
	}
	producto, err := ctl.repo.ObtenerProductoPorID(c.Context(), id, false)
	if err != nil {
		return ctl.errorRepo(c, err, "Producto no encontrado")
	}
	return c.JSON(producto)
}

// ListarAdmin: GET /api/admin/productos?page&limit&q — incluye inactivos.
func (ctl *ProductoController) ListarAdmin(c *fiber.Ctx) error {
	page, limit := paginacion(c)
	pagina, err := ctl.repo.ListarProductosAdmin(c.Context(), c.Query("q"), page, limit)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	return c.JSON(pagina)
}

// Crear: POST /api/productos
func (ctl *ProductoController) Crear(c *fiber.Ctx) error {
	var in models.ProductoInput
	if err := c.BodyParser(&in); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}
	p := models.Producto{Galeria: []string{}}
	if in.Nombre != nil {
		p.Nombre = strings.TrimSpace(*in.Nombre)
	}
	if in.Descripcion != nil {
		p.Descripcion = strings.TrimSpace(*in.Descripcion)
	}
	if in.Precio != nil {
		p.Precio = *in.Precio
	}
	if in.Categoria != nil {
		p.Categoria = strings.TrimSpace(*in.Categoria)
	}
	if in.Img != nil {
		p.Img = strings.TrimSpace(*in.Img)
	}
	if in.Galeria != nil {
		p.Galeria = *in.Galeria
	}
	if in.Stock != nil {
		p.Stock = *in.Stock
	}

	if errs := validarProducto(p); errs != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Revisa los datos del producto", errs)
	}

	creado, err := ctl.repo.CrearProducto(c.Context(), p)
	if err != nil {
		return ctl.errorInterno(c, err)
	}
	ctl.cache.Eliminar(c.Context(), cache.ClaveProductos)
	return c.Status(fiber.StatusCreated).JSON(creado)
}

// Actualizar: PUT /api/productos/:id — edición parcial.
func (ctl *ProductoController) Actualizar(c *fiber.Ctx) error {
	id, err := ctl.paramID(c)
	if err != nil {
		return err
	}
	var in models.ProductoInput
	if err := c.BodyParser(&in); err != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Cuerpo de petición inválido", nil)
	}

	// Validamos el resultado final: el producto actual con los cambios aplicados.
	actual, err := ctl.repo.ObtenerProductoPorID(c.Context(), id, true)
	if err != nil {
		return ctl.errorRepo(c, err, "Producto no encontrado")
	}
	if in.Nombre != nil {
		*in.Nombre = strings.TrimSpace(*in.Nombre)
		actual.Nombre = *in.Nombre
	}
	if in.Categoria != nil {
		*in.Categoria = strings.TrimSpace(*in.Categoria)
		actual.Categoria = *in.Categoria
	}
	if in.Precio != nil {
		actual.Precio = *in.Precio
	}
	if in.Stock != nil {
		actual.Stock = *in.Stock
	}
	if errs := validarProducto(*actual); errs != nil {
		return ctl.errorJSON(c, fiber.StatusBadRequest, "Revisa los datos del producto", errs)
	}

	producto, err := ctl.repo.ActualizarProducto(c.Context(), id, in)
	if err != nil {
		return ctl.errorRepo(c, err, "Producto no encontrado")
	}
	ctl.cache.Eliminar(c.Context(), cache.ClaveProductos)
	return c.JSON(producto)
}

// Desactivar: DELETE /api/productos/:id — borrado lógico (activo = false).
func (ctl *ProductoController) Desactivar(c *fiber.Ctx) error {
	id, err := ctl.paramID(c)
	if err != nil {
		return err
	}
	if err := ctl.repo.DesactivarProducto(c.Context(), id); err != nil {
		return ctl.errorRepo(c, err, "Producto no encontrado")
	}
	ctl.cache.Eliminar(c.Context(), cache.ClaveProductos)
	return c.JSON(fiber.Map{"message": "Producto desactivado", "id": id})
}

func validarProducto(p models.Producto) map[string]string {
	v := validacion.Errores{}
	v.Requerido("nombre", p.Nombre, "El nombre es obligatorio")
	v.Requerido("categoria", p.Categoria, "La categoría es obligatoria")
	v.Positivo("precio", p.Precio, "El precio debe ser mayor que cero")
	v.NoNegativo("stock", p.Stock, "El stock no puede ser negativo")
	return v.Validar()
}
