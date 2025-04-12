package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

var jsonFunc = json.Marshal

func httpClientEvolution[responseType any](method, uri string, payload *bytes.Buffer) (*responseType, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	evolutionPort := os.Getenv("EVOLUTION_PORT")
	evolutionHost := os.Getenv("EVOLUTION_HOST")
	evolutionUrl := fmt.Sprintf("http://%s:%s", evolutionHost, evolutionPort) + uri

	evolutionApiKey := os.Getenv("AUTHENTICATION_API_KEY")
	req, err := http.NewRequest(method, evolutionUrl, payload)
	if err != nil {
		err := customerror.Make("Falha ao criar a requisição", http.StatusInternalServerError, err)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", evolutionApiKey)
	resp, err := client.Do(req)
	if err != nil {
		err := customerror.Make("Falha ao fazer a requisição", resp.StatusCode, err)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Println("body", string(body))
	fmt.Println("resp.StatusCode", resp.StatusCode)
	if err != nil {
		err := customerror.Make("Falha ao ler o corpo da resposta", http.StatusInternalServerError, err)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	var responseData responseType
	err = json.Unmarshal(body, &responseData)
	if err != nil {

		err := customerror.Make("Falha ao decodificar a resposta", http.StatusInternalServerError, err)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	return &responseData, nil
}

func createEvolutionInstance(phone, instanceName string, qrCode bool) (instanceId, qrCodeString string, err error) {
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
	fmt.Println("resp", resp)
	if err != nil {
		return "", "", customerror.Trace("createEvolutionInstance: ", err)
	}

	return resp.Instance.InstanceID, resp.Qrcode.Code, nil
}
