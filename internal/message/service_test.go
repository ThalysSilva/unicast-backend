package message

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ThalysSilva/unicast-backend/pkg/mailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestBuildEmailAttachmentsSupportsBase64DataAndURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/arquivo.pdf", r.URL.Path)
		_, _ = w.Write([]byte("arquivo remoto"))
	}))
	defer server.Close()

	attachments, err := buildEmailAttachments(context.Background(), &Message{
		Attachments: &[]Attachment{
			{
				FileName: "local.txt",
				Data:     []byte("arquivo local"),
			},
			{
				FileName: "arquivo.pdf",
				URL:      server.URL + "/arquivo.pdf",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, []mailer.Attachment{
		{
			FileName: "local.txt",
			Data:     []byte("arquivo local"),
		},
		{
			FileName: "arquivo.pdf",
			Data:     []byte("arquivo remoto"),
		},
	}, attachments)
}

func TestBuildEmailAttachmentsReturnsErrorWhenURLDownloadFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "erro", http.StatusBadGateway)
	}))
	defer server.Close()

	attachments, err := buildEmailAttachments(context.Background(), &Message{
		Attachments: &[]Attachment{
			{
				FileName: "arquivo.pdf",
				URL:      server.URL + "/arquivo.pdf",
			},
		},
	})

	assert.Nil(t, attachments)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao baixar anexo para email")
}

func TestBuildEmailAttachmentsRejectsBlockedExtension(t *testing.T) {
	attachments, err := buildEmailAttachments(context.Background(), &Message{
		Attachments: &[]Attachment{
			{
				FileName: "virus.exe",
				Data:     []byte("fake exe"),
			},
		},
	})

	assert.Nil(t, attachments)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tipo de arquivo não permitido")
}

func TestValidateAttachmentCountRejectsMoreThanFiveFiles(t *testing.T) {
	attachments := []Attachment{
		{FileName: "1.pdf", Data: []byte("1")},
		{FileName: "2.pdf", Data: []byte("2")},
		{FileName: "3.pdf", Data: []byte("3")},
		{FileName: "4.pdf", Data: []byte("4")},
		{FileName: "5.pdf", Data: []byte("5")},
		{FileName: "6.pdf", Data: []byte("6")},
	}

	err := validateAttachmentCount(&attachments)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantidade de anexos excede o limite permitido")
}

func TestBuildWhatsAppAttachmentsRejectsPayloadAboveLimit(t *testing.T) {
	large := make([]byte, 9*1024*1024)
	attachments, names, err := buildWhatsAppAttachments(&Message{
		Attachments: &[]Attachment{
			{FileName: "a.pdf", Data: large},
			{FileName: "b.pdf", Data: large},
		},
	})

	assert.Nil(t, attachments)
	assert.Empty(t, names)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anexos excedem o limite total do WhatsApp")
}
