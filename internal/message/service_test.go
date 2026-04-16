package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatWhatsAppBodyUsesSubjectAsBoldTitle(t *testing.T) {
	got := formatWhatsAppBody("Aviso importante", "A aula foi remarcada.")

	assert.Equal(t, "*Aviso importante*\n\nA aula foi remarcada.", got)
}

func TestFormatWhatsAppBodyKeepsBodyWhenSubjectIsBlank(t *testing.T) {
	got := formatWhatsAppBody("  ", "Mensagem sem assunto.")

	assert.Equal(t, "Mensagem sem assunto.", got)
}

func TestFormatWhatsAppBodyNormalizesSubjectNewlines(t *testing.T) {
	got := formatWhatsAppBody("Aviso\nurgente", "Verifique o portal.")

	assert.Equal(t, "*Aviso urgente*\n\nVerifique o portal.", got)
}
