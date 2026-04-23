package whatsapp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ThalysSilva/unicast-backend/internal/config/env"
	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type newEvolutionInstanceReturn struct {
	Instance struct {
		InstanceName          string      `json:"instanceName"`
		InstanceID            string      `json:"instanceId"`
		Integration           string      `json:"integration"`
		WebhookWaBusiness     interface{} `json:"webhookWaBusiness"`
		AccessTokenWaBusiness string      `json:"accessTokenWaBusiness"`
		Status                string      `json:"status"`
	} `json:"instance"`
	Hash    string `json:"hash"`
	Webhook struct {
	} `json:"webhook"`
	Websocket struct {
	} `json:"websocket"`
	Rabbitmq struct {
	} `json:"rabbitmq"`
	Sqs struct {
	} `json:"sqs"`
	Settings struct {
		RejectCall      bool   `json:"rejectCall"`
		MsgCall         string `json:"msgCall"`
		GroupsIgnore    bool   `json:"groupsIgnore"`
		AlwaysOnline    bool   `json:"alwaysOnline"`
		ReadMessages    bool   `json:"readMessages"`
		ReadStatus      bool   `json:"readStatus"`
		SyncFullHistory bool   `json:"syncFullHistory"`
	} `json:"settings"`
	Qrcode struct {
		PairingCode interface{} `json:"pairingCode"`
		Code        string      `json:"code"`
		Base64      string      `json:"base64"`
		Count       int         `json:"count"`
	} `json:"qrcode"`
}

type newEvolutionPayload struct {
	Phone        string `json:"phone" validate:"required"`
	InstanceName string `json:"instanceName"`
	QrCode       bool   `json:"qrcode"`
	Integration  string `json:"integration"`
}

type sendTextPayload struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

type sendTextResponse struct {
	Key struct {
		RemoteJid string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
		ID        string `json:"id"`
	} `json:"key"`
	Status  string          `json:"status"`
	Message json.RawMessage `json:"message"`
}

type deleteInstanceResponse struct {
	Message string `json:"message"`
}

type sendMediaPayload struct {
	Number    string `json:"number"`
	Media     string `json:"media"` // url ou base64
	MediaType string `json:"mediatype"`
	MimeType  string `json:"mimetype,omitempty"`
	FileName  string `json:"fileName,omitempty"`
	Caption   string `json:"caption,omitempty"`
}

type sendMediaResponse struct {
	Status  string          `json:"status"`
	Message json.RawMessage `json:"message"`
	Key     struct {
		RemoteJid string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
		ID        string `json:"id"`
	} `json:"key"`
	MessageTimestamp json.RawMessage `json:"messageTimestamp"`
}

type connectResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	PairingCode string `json:"pairingCode"`
	Code        string `json:"code"`
	Base64      string `json:"base64"`
	Count       int    `json:"count"`
	Qrcode      struct {
		Code   string `json:"code"`
		Base64 string `json:"base64"`
	} `json:"qrcode"`
}

type statusResponse struct {
	Instance struct {
		Status string `json:"status"`
		State  string `json:"state"`
	} `json:"instance"`
}

var jsonFunc = json.Marshal
var evolutionBaseURL = ""
var cachedConfig *env.Config

func httpClientEvolution[responseType any](method, uri string, payload *bytes.Buffer) (*responseType, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	if cachedConfig == nil {
		cfg, err := env.Load()
		if err != nil {
			return nil, customerror.Trace("HTTPClientEvolution", err)
		}
		cachedConfig = cfg
	}
	evolutionUrl := fmt.Sprintf("http://%s:%s", cachedConfig.Evolution.Host, cachedConfig.Evolution.Port) + uri

	evolutionApiKey := cachedConfig.Evolution.APIKey
	req, err := http.NewRequest(method, evolutionUrl, payload)
	if err != nil {
		err := customerror.Make("Falha ao criar a requisição", http.StatusInternalServerError, err)
		return nil, customerror.Trace("HTTPClientEvolution", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", evolutionApiKey)
	resp, err := client.Do(req)
	if err != nil {
		err := customerror.Make("Falha ao fazer a requisição", http.StatusBadGateway, err)
		return nil, customerror.Trace("HTTPClientEvolution", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err := customerror.Make("Falha ao ler o corpo da resposta", http.StatusInternalServerError, err)
		return nil, customerror.Trace("HTTPClientEvolution", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseSnippet := compactResponseBody(body)
		detail := fmt.Sprintf("%s %s -> status %d", method, uri, resp.StatusCode)
		if responseSnippet != "" {
			detail += fmt.Sprintf(", body=%q", responseSnippet)
		}
		err := customerror.Make(
			fmt.Sprintf("Evolution API retornou status %d em %s %s", resp.StatusCode, method, uri),
			resp.StatusCode,
			errors.New(detail),
		)
		return nil, customerror.Trace("HTTPClientEvolution", err)
	}
	var responseData responseType
	err = json.Unmarshal(body, &responseData)
	if err != nil {

		err := customerror.Make("Falha ao decodificar a resposta", http.StatusInternalServerError, err)
		return nil, customerror.Trace("HTTPClientEvolution", err)
	}
	return &responseData, nil
}

func compactResponseBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}
	text = strings.Join(strings.Fields(text), " ")
	const maxLen = 300
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	}
	return text
}

func createEvolutionInstance(phone, instanceName string, qrCode bool) (createdName, qrCodeString string, err error) {
	jsonData, err := jsonFunc(newEvolutionPayload{
		Phone:        phone,
		InstanceName: instanceName,
		QrCode:       qrCode,
		Integration:  "WHATSAPP-BAILEYS",
	})
	if err != nil {
		return "", "", customerror.Make("createEvolutionInstance: Falha ao codificar o payload", http.StatusInternalServerError, err)
	}
	payload := bytes.NewBuffer(jsonData)
	resp, err := httpClientEvolution[newEvolutionInstanceReturn]("POST", "/instance/create", payload)
	if err != nil {
		return "", "", customerror.Trace("createEvolutionInstance", err)
	}

	createdName = resp.Instance.InstanceName
	if createdName == "" {
		createdName = instanceName
	}
	return createdName, resp.Qrcode.Code, nil
}

// sendEvolutionText envia uma mensagem de texto simples usando a Evolution API.
func sendEvolutionText(instanceName, number, text string) error {
	body, err := jsonFunc(sendTextPayload{
		Number: evolutionRecipientJID(number),
		Text:   text,
	})
	if err != nil {
		return customerror.Trace("sendEvolutionText: marshal", err)
	}

	payload := bytes.NewBuffer(body)
	resp, err := httpClientEvolution[sendTextResponse]("POST", "/message/sendText/"+instanceName, payload)
	if err != nil {
		return err
	}

	if resp == nil {
		return customerror.Make("resposta vazia da Evolution API", http.StatusBadGateway, fmt.Errorf("empty response"))
	}

	return nil
}

// sendEvolutionMedia envia mídia/base64 ou URL via Evolution API.
func sendEvolutionMedia(instanceName string, payload sendMediaPayload) (*sendMediaResponse, error) {
	body, err := jsonFunc(payload)
	if err != nil {
		return nil, customerror.Trace("sendEvolutionMedia: marshal", err)
	}

	buf := bytes.NewBuffer(body)
	resp, err := httpClientEvolution[sendMediaResponse]("POST", "/message/sendMedia/"+instanceName, buf)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, customerror.Make("resposta vazia da Evolution API", http.StatusBadGateway, fmt.Errorf("empty response"))
	}

	return resp, nil
}

// deleteEvolutionInstance remove uma instância na Evolution API.
func deleteEvolutionInstance(instanceName string) error {
	payload := bytes.NewBuffer(nil)
	encodedName := url.PathEscape(instanceName)
	_, err := httpClientEvolution[deleteInstanceResponse]("DELETE", fmt.Sprintf("/instance/delete/%s", encodedName), payload)
	return err
}

// connectEvolutionInstance dispara a conexão/pareamento (precisa do número).
func connectEvolutionInstance(instanceName, number string) (*connectResponse, error) {
	payload := bytes.NewBuffer(nil)
	encodedName := url.PathEscape(instanceName)
	resp, err := httpClientEvolution[connectResponse]("GET", fmt.Sprintf("/instance/connect/%s", encodedName), payload)
	if err != nil && number != "" {
		resp, err = httpClientEvolution[connectResponse]("GET", fmt.Sprintf("/instance/connect/%s?number=%s", encodedName, number), payload)
		if err != nil {
			return nil, err
		}
	}
	if resp == nil {
		return nil, customerror.Make("resposta vazia da Evolution API (connect)", http.StatusBadGateway, fmt.Errorf("empty response"))
	}
	if resp.Base64 == "" && resp.Qrcode.Base64 != "" {
		resp.Base64 = resp.Qrcode.Base64
	}
	if resp.Code == "" && resp.Qrcode.Code != "" {
		resp.Code = resp.Qrcode.Code
	}
	if resp.Qrcode.Base64 == "" && resp.Base64 != "" {
		resp.Qrcode.Base64 = resp.Base64
	}
	if resp.Qrcode.Code == "" && resp.Code != "" {
		resp.Qrcode.Code = resp.Code
	}
	return resp, nil
}

// connectionStateEvolution retorna o status da instância.
func connectionStateEvolution(instanceName string) (string, error) {
	payload := bytes.NewBuffer(nil)
	encodedName := url.PathEscape(instanceName)
	resp, err := httpClientEvolution[statusResponse]("GET", fmt.Sprintf("/instance/connectionState/%s", encodedName), payload)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", customerror.Make("resposta vazia da Evolution API (connectionState)", http.StatusBadGateway, fmt.Errorf("empty response"))
	}
	if resp.Instance.Status == "" && resp.Instance.State != "" {
		return resp.Instance.State, nil
	}
	return resp.Instance.Status, nil
}

func logoutEvolutionInstance(instanceName string) error {
	payload := bytes.NewBuffer(nil)
	encodedName := url.PathEscape(instanceName)
	_, err := httpClientEvolution[deleteInstanceResponse]("DELETE", fmt.Sprintf("/instance/logout/%s", encodedName), payload)
	return err
}

func restartEvolutionInstance(instanceName string) error {
	payload := bytes.NewBuffer(nil)
	encodedName := url.PathEscape(instanceName)
	_, err := httpClientEvolution[deleteInstanceResponse]("POST", fmt.Sprintf("/instance/restart/%s", encodedName), payload)
	return err
}
