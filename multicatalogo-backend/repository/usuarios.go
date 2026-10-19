package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"multicatalogo-backend/models"

	"github.com/jackc/pgx/v5"
)

const columnasUsuario = `u.id, u.email, u.password, u.rol, u.nombre, u.activo, u.referido_id,
	COALESCE(r.codigo, ''), u.creado_en`

func escanearUsuario(fila pgx.Row) (*models.Usuario, error) {
	var u models.Usuario
	err := fila.Scan(&u.ID, &u.Email, &u.Password, &u.Rol, &u.Nombre, &u.Activo, &u.ReferidoID, &u.Codigo, &u.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// BuscarUsuarioPorEmail devuelve el usuario o ErrNoEncontrado.
func (r *PostgresRepository) BuscarUsuarioPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	return escanearUsuario(r.db.QueryRow(ctx, `SELECT `+columnasUsuario+`
		FROM usuarios u LEFT JOIN referidos r ON r.id = u.referido_id
		WHERE u.email = $1`, strings.ToLower(email)))
}

// BuscarUsuarioPorID devuelve el usuario o ErrNoEncontrado.
func (r *PostgresRepository) BuscarUsuarioPorID(ctx context.Context, id int) (*models.Usuario, error) {
	return escanearUsuario(r.db.QueryRow(ctx, `SELECT `+columnasUsuario+`
		FROM usuarios u LEFT JOIN referidos r ON r.id = u.referido_id
		WHERE u.id = $1`, id))
}

// RegistrarUsuario crea la cuenta (rol cliente) y su nodo en la red en UNA transacción.
// Con un código ref válido el nodo cuelga del referente; sin código cuelga de la raíz (la empresa).
func (r *PostgresRepository) RegistrarUsuario(ctx context.Context, nombre, email, hash, ref string) (*models.Usuario, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	email = strings.ToLower(email)
	var existe bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM usuarios WHERE email = $1)", email).Scan(&existe); err != nil {
		return nil, err
	}
	if existe {
		return nil, ErrDuplicado
	}

	// Nodo padre: el del código de invitación, o la raíz de la red.
	var padreID, padreNivel int
	if ref = strings.ToUpper(strings.TrimSpace(ref)); ref != "" {
		err = tx.QueryRow(ctx, "SELECT id, nivel FROM referidos WHERE codigo = $1", ref).Scan(&padreID, &padreNivel)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errValidacion("El código de referido %q no existe", ref)
		}
	} else {
		err = tx.QueryRow(ctx, "SELECT id, nivel FROM referidos WHERE parent_id IS NULL ORDER BY id LIMIT 1").
			Scan(&padreID, &padreNivel)
	}
	if err != nil {
		return nil, err
	}

	codigo, err := generarCodigo(ctx, tx, nombre)
	if err != nil {
		return nil, err
	}

	var referidoID int
	err = tx.QueryRow(ctx, `
		INSERT INTO referidos (nombre, nivel, ventas, parent_id, codigo)
		VALUES ($1, $2, 0, $3, $4) RETURNING id`,
		nombre, padreNivel+1, padreID, codigo).Scan(&referidoID)
	if err != nil {
		return nil, err
	}

	var usuarioID int
	err = tx.QueryRow(ctx, `
		INSERT INTO usuarios (email, password, rol, nombre, referido_id)
		VALUES ($1, $2, 'cliente', $3, $4) RETURNING id`,
		email, hash, nombre, referidoID).Scan(&usuarioID)
	if codigoPG(err) == "23505" {
		return nil, ErrDuplicado
	}
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.BuscarUsuarioPorID(ctx, usuarioID)
}

// generarCodigo arma un código de invitación legible y único: letras del nombre + 3 dígitos.
func generarCodigo(ctx context.Context, tx pgx.Tx, nombre string) (string, error) {
	sinTildes := strings.NewReplacer("Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ü", "U", "Ñ", "N")
	base := strings.Builder{}
	for _, letra := range sinTildes.Replace(strings.ToUpper(nombre)) {
		if letra >= 'A' && letra <= 'Z' {
			base.WriteRune(letra)
		} else if letra == ' ' && base.Len() > 0 {
			break // solo el primer nombre
		}
		if base.Len() == 8 {
			break
		}
	}
	prefijo := base.String()
	if prefijo == "" {
		prefijo = "REF"
	}

	for intento := 0; intento < 10; intento++ {
		n, err := rand.Int(rand.Reader, big.NewInt(1000))
		if err != nil {
			return "", err
		}
		codigo := fmt.Sprintf("%s%03d", prefijo, n.Int64())
		var existe bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM referidos WHERE codigo = $1)", codigo).Scan(&existe); err != nil {
			return "", err
		}
		if !existe {
			return codigo, nil
		}
	}
	return "", errors.New("no se pudo generar un código de referido único")
}

// ListarUsuarios devuelve las cuentas paginadas, con búsqueda opcional por nombre o email.
func (r *PostgresRepository) ListarUsuarios(ctx context.Context, q string, page, limit int) (models.Pagina[models.Usuario], error) {
	page, limit, offset := paginar(page, limit)
	pagina := models.Pagina[models.Usuario]{Data: []models.Usuario{}, Page: page, Limit: limit}
	filtro := "%" + strings.TrimSpace(q) + "%"

	err := r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM usuarios WHERE nombre ILIKE $1 OR email ILIKE $1", filtro).Scan(&pagina.Total)
	if err != nil {
		return pagina, err
	}

	filas, err := r.db.Query(ctx, `SELECT `+columnasUsuario+`
		FROM usuarios u LEFT JOIN referidos r ON r.id = u.referido_id
		WHERE u.nombre ILIKE $1 OR u.email ILIKE $1
		ORDER BY u.id LIMIT $2 OFFSET $3`, filtro, limit, offset)
	if err != nil {
		return pagina, err
	}
	defer filas.Close()
	for filas.Next() {
		u, err := escanearUsuario(filas)
		if err != nil {
			return pagina, err
		}
		pagina.Data = append(pagina.Data, *u)
	}
	return pagina, filas.Err()
}

// ActualizarUsuario cambia rol y/o estado activo.
func (r *PostgresRepository) ActualizarUsuario(ctx context.Context, id int, cambios models.CambiosUsuario) (*models.Usuario, error) {
	etiqueta, err := r.db.Exec(ctx, `
		UPDATE usuarios SET rol = COALESCE($2, rol), activo = COALESCE($3, activo)
		WHERE id = $1`, id, cambios.Rol, cambios.Activo)
	if err != nil {
		return nil, err
	}
	if etiqueta.RowsAffected() == 0 {
		return nil, ErrNoEncontrado
	}
	return r.BuscarUsuarioPorID(ctx, id)
}
