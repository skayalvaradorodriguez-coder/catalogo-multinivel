// Package notificaciones define el envío de correos transaccionales.
// En el curso no hay servidor SMTP: LogMailer solo escribe en la consola, pero como los
// controladores dependen de la interfaz Mailer, cambiarlo por un envío real no toca su código.
package notificaciones

import "log"

// Mailer envía un correo.
type Mailer interface {
	Enviar(para, asunto, cuerpo string) error
}

// LogMailer simula el envío escribiendo el correo en el log.
type LogMailer struct{}

func (LogMailer) Enviar(para, asunto, cuerpo string) error {
	log.Printf("[correo] para=%s asunto=%q: %s", para, asunto, cuerpo)
	return nil
}
