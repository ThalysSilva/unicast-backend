package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ThalysSilva/unicast-backend/pkg/customerror"
)

type newEvolutionInstanceReturn struct {
	QrCode   string `json:"qrcode"`
	Instance struct {
		InstanceName          string `json:"instanceName"`
		InstanceID            string `json:"instanceId"`
		WebhookWaBusiness     any    `json:"webhook_wa_business"`
		AccessTokenWaBusiness string `json:"access_token_wa_business"`
		Status                string `json:"status"`
	} `json:"instance"`
	Hash struct {
		Apikey string `json:"apikey"`
	} `json:"hash"`
	Settings struct {
		RejectCall      bool   `json:"reject_call"`
		MsgCall         string `json:"msg_call"`
		GroupsIgnore    bool   `json:"groups_ignore"`
		AlwaysOnline    bool   `json:"always_online"`
		ReadMessages    bool   `json:"read_messages"`
		ReadStatus      bool   `json:"read_status"`
		SyncFullHistory bool   `json:"sync_full_history"`
	} `json:"settings"`
}

type newEvolutionPayload struct {
	Phone        string `json:"phone" validate:"required"`
	InstanceName string `json:"instanceName"`
	QrCode       bool   `json:"qrCode"`
}

var jsonFunc = json.Marshal

func httpClientEvolution[responseType any](method, uri string, payload *bytes.Buffer) (*responseType, error) {
	client := &http.Client{}
	evolutionPort := os.Getenv("EVOLUTION_PORT")
	evolutionUrl := fmt.Sprintf("http://evolution-api-unicast:%s", evolutionPort) + uri

	evolutionApiKey := os.Getenv("AUTHENTICATION_API_KEY")
	req, err := http.NewRequest(method, evolutionUrl, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", evolutionApiKey)
	resp, err := client.Do(req)
	if err != nil {
		err := customerror.Make("Falha ao fazer a requisição", resp.StatusCode)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err := customerror.Make("Falha ao ler o corpo da resposta", http.StatusInternalServerError)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	var responseData responseType
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		err := customerror.Make("Falha ao decodificar a resposta", http.StatusInternalServerError)
		return nil, customerror.Trace("HTTPClientEvolution: ", err)
	}
	return &responseData, nil
}

func createEvolutionInstance(phone, instanceName string, qrCode bool) (instanceId, qrCodeString string, err error) {
	jsonData, err := jsonFunc(newEvolutionPayload{
		Phone:        phone,
		InstanceName: instanceName,
		QrCode:       qrCode,
	})
	if err != nil {
		return "", "", customerror.Trace("createEvolutionInstance: ", err)
	}
	payload := bytes.NewBuffer(jsonData)
	resp, err := httpClientEvolution[newEvolutionInstanceReturn]("POST", "/instance/create", payload)
	if err != nil {
		return "", "", customerror.Trace("createEvolutionInstance: ", err)
	}

	return resp.Instance.InstanceName, resp.QrCode, nil
}
