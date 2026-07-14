package whatsapp

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ThalysSilva/unicast-backend/internal/config/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendEvolutionTextUsesInstancePathAndJIDPayload(t *testing.T) {
	var gotPath string
	var gotAPIKey string
	var gotPayload sendTextPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("apikey")

		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"key":{"remoteJid":"5500000000001@s.whatsapp.net","fromMe":true,"id":"3EB0313C9EA80A7ED95190"},
			"status":"PENDING",
			"message":{"conversation":"enviando mensagem de teste"}
		}`))
	}))
	defer server.Close()

	setEvolutionTestConfig(t, server.URL, "test-api-key")

	err := sendEvolutionText("professor@example.com:5500000000000", "+5500000000001", "enviando mensagem de teste")

	require.NoError(t, err)
	assert.Equal(t, "/message/sendText/professor@example.com:5500000000000", gotPath)
	assert.Equal(t, "test-api-key", gotAPIKey)
	assert.Equal(t, sendTextPayload{
		Number: "5500000000001@s.whatsapp.net",
		Text:   "enviando mensagem de teste",
	}, gotPayload)
}

func TestEvolutionRecipientJID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "digits only", in: "5500000000001", want: "5500000000001@s.whatsapp.net"},
		{name: "plus prefixed", in: "+5500000000001", want: "5500000000001@s.whatsapp.net"},
		{name: "formatted", in: "+55 (00) 00000-0001", want: "5500000000001@s.whatsapp.net"},
		{name: "already jid", in: "5500000000001@s.whatsapp.net", want: "5500000000001@s.whatsapp.net"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, evolutionRecipientJID(tt.in))
		})
	}
}

func TestBuildEvolutionURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{
			name: "docker service uses configured port",
			host: "evolution-api-unicast",
			port: "8080",
			want: "http://evolution-api-unicast:8080/instance/create",
		},
		{
			name: "host with scheme uses configured port",
			host: "http://localhost",
			port: "8081",
			want: "http://localhost:8081/instance/create",
		},
		{
			name: "explicit host port is preserved",
			host: "http://localhost:9090",
			port: "8081",
			want: "http://localhost:9090/instance/create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildEvolutionURL(tt.host, tt.port, "/instance/create")

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSendMediaUsesEvolutionMediaContract(t *testing.T) {
	var gotPath string
	var gotPayload sendMediaPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"key":{"remoteJid":"5500000000001@s.whatsapp.net","fromMe":true,"id":"3EB045EE1466BB3440F04F"},
			"status":"PENDING",
			"message":{"imageMessage":{"mimetype":"image/webp","caption":"texto com imagem"}},
			"messageTimestamp":1776300381
		}`))
	}))
	defer server.Close()

	setEvolutionTestConfig(t, server.URL, "test-api-key")

	resp, err := SendMedia(
		"professor@example.com:5500000000000",
		"+5500000000001",
		"3.webp",
		[]byte("test image data"),
		"texto com imagem",
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "/message/sendMedia/professor@example.com:5500000000000", gotPath)
	assert.Equal(t, sendMediaPayload{
		Number:    "5500000000001@s.whatsapp.net",
		Media:     "dGVzdCBpbWFnZSBkYXRh",
		MediaType: "image",
		MimeType:  "image/webp",
		FileName:  "3.webp",
		Caption:   "texto com imagem",
	}, gotPayload)
	assert.JSONEq(t, `1776300381`, string(resp.MessageTimestamp))
}

func TestSendMediaUsesEvolutionVideoContract(t *testing.T) {
	var gotPath string
	var gotPayload sendMediaPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"key":{"remoteJid":"5500000000001@s.whatsapp.net","fromMe":true,"id":"3EB0673A06EC8F9FB8D059"},
			"status":"PENDING",
			"message":{"videoMessage":{"mimetype":"video/mp4","caption":"Video teste com texto","gifPlayback":false}},
			"messageType":"videoMessage",
			"messageTimestamp":1776300648
		}`))
	}))
	defer server.Close()

	setEvolutionTestConfig(t, server.URL, "test-api-key")

	resp, err := SendMedia(
		"professor@example.com:5500000000000",
		"+5500000000001",
		"Clair_Obscure_Expedition_.mp4",
		[]byte("test video data"),
		"Video teste com texto",
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "/message/sendMedia/professor@example.com:5500000000000", gotPath)
	assert.Equal(t, sendMediaPayload{
		Number:    "5500000000001@s.whatsapp.net",
		Media:     "dGVzdCB2aWRlbyBkYXRh",
		MediaType: "video",
		MimeType:  "video/mp4",
		FileName:  "Clair_Obscure_Expedition_.mp4",
		Caption:   "Video teste com texto",
	}, gotPayload)
	assert.JSONEq(t, `1776300648`, string(resp.MessageTimestamp))
}

func TestSendMediaUsesEvolutionDocumentContract(t *testing.T) {
	var gotPath string
	var gotPayload sendMediaPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"key":{"remoteJid":"5500000000001@s.whatsapp.net","fromMe":true,"id":"3EB020BC7FEB8824CB9BC4"},
			"status":"PENDING",
			"message":{"documentMessage":{"mimetype":"application/pdf","fileName":"ML-INFORME-RENDIMENTOS-2025 (2).pdf","caption":"envio de arquivos como documento"}},
			"messageType":"documentMessage",
			"messageTimestamp":1776300800
		}`))
	}))
	defer server.Close()

	setEvolutionTestConfig(t, server.URL, "test-api-key")

	resp, err := SendMedia(
		"professor@example.com:5500000000000",
		"+5500000000001",
		"ML-INFORME-RENDIMENTOS-2025 (2).pdf",
		[]byte("test pdf data"),
		"envio de arquivos como documento",
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "/message/sendMedia/professor@example.com:5500000000000", gotPath)
	assert.Equal(t, sendMediaPayload{
		Number:    "5500000000001@s.whatsapp.net",
		Media:     "dGVzdCBwZGYgZGF0YQ==",
		MediaType: "document",
		MimeType:  "application/pdf",
		FileName:  "ML-INFORME-RENDIMENTOS-2025 (2).pdf",
		Caption:   "envio de arquivos como documento",
	}, gotPayload)
	assert.JSONEq(t, `1776300800`, string(resp.MessageTimestamp))
}

func TestSendEvolutionTextIncludesEvolutionBodyOnHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"response":{"message":["Unauthorized"]}}`))
	}))
	defer server.Close()

	setEvolutionTestConfig(t, server.URL, "test-api-key")

	err := sendEvolutionText("professor@example.com:5500000000000", "+5500000000001", "enviando mensagem de teste")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Evolution API retornou status 403 em POST /message/sendText/professor@example.com:5500000000000")
	assert.Contains(t, err.Error(), `body="{\"response\":{\"message\":[\"Unauthorized\"]}}"`)
}

func setEvolutionTestConfig(t *testing.T, rawURL, apiKey string) {
	t.Helper()

	previousConfig := cachedConfig
	t.Cleanup(func() {
		cachedConfig = previousConfig
	})

	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)

	host, port, err := net.SplitHostPort(parsed.Host)
	require.NoError(t, err)

	cachedConfig = &env.Config{
		Evolution: env.Evolution{
			Host:   host,
			Port:   port,
			APIKey: apiKey,
		},
	}
}
