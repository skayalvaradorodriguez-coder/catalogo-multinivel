// Package seguridad agrupa el hash de contraseñas (bcrypt) y los tokens JWT.
package seguridad

import "golang.org/x/crypto/bcrypt"

// HashPassword genera el hash bcrypt de una contraseña.
// Es compatible con los hashes que crea pgcrypto (crypt + gen_salt('bf')) en las migraciones.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerificarPassword compara la contraseña escrita con el hash guardado.
func VerificarPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
