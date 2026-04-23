# WhatsApp com Evolution API

Este documento registra o contrato usado pelo backend para envio de WhatsApp via Evolution API.

## Visão Geral

O Unicast usa instâncias da Evolution para enviar mensagens pelo WhatsApp. A instância local fica salva em `whatsapp_instances`, e o campo `instance_name` é usado no path dos endpoints da Evolution.

Formato atual de `instance_name` criado pelo backend:

```txt
email-do-professor:telefone-da-instancia
```

Exemplo:

```txt
professor@example.com:5500000000000
```

## Endpoints da Evolution Usados

Envio de texto:

```txt
POST /message/sendText/{instanceName}
```

Envio de mídia:

```txt
POST /message/sendMedia/{instanceName}
```

O backend envia o header:

```txt
apikey: <AUTHENTICATION_API_KEY>
Content-Type: application/json
```

## Formato do Destinatário

A Evolution espera o destinatário como JID:

```txt
5500000000001@s.whatsapp.net
```

O backend aceita telefones como `+5500000000001`, `5500000000001` ou formatos com pontuação, e converte para JID antes de chamar a Evolution.

## Texto

Payload enviado para a Evolution:

```json
{
  "number": "5500000000001@s.whatsapp.net",
  "text": "*Aviso importante*\n\nA aula foi remarcada para sexta-feira."
}
```

No `POST /message/send`, o campo `subject` vira título em negrito no WhatsApp, e `body` fica abaixo:

```txt
*Assunto*

Corpo da mensagem
```

## Mídia

Payload base:

```json
{
  "number": "5500000000001@s.whatsapp.net",
  "mediatype": "image",
  "mimetype": "image/webp",
  "caption": "*Aviso importante*\n\nSegue o arquivo.",
  "media": "base64-ou-url",
  "fileName": "arquivo.webp"
}
```

Campos:

- `number`: JID do destinatário.
- `mediatype`: `image`, `video`, `audio` ou `document`.
- `mimetype`: MIME detectado pelo backend a partir da extensão e/ou conteúdo.
- `caption`: opcional. No fluxo atual do backend, a mensagem de texto é enviada antes e os anexos seguem sem legenda.
- `media`: base64 do arquivo ou URL pública.
- `fileName`: nome do arquivo exibido/associado pela Evolution.

## Exemplos por Tipo

Imagem `.webp`:

```json
{
  "number": "5500000000001@s.whatsapp.net",
  "mediatype": "image",
  "mimetype": "image/webp",
  "caption": "texto com imagem",
  "media": "UklGR...",
  "fileName": "3.webp"
}
```

Vídeo `.mp4`:

```json
{
  "number": "5500000000001@s.whatsapp.net",
  "mediatype": "video",
  "mimetype": "video/mp4",
  "caption": "Video teste com texto",
  "media": "AAAAIGZ0eXBpc29t...",
  "fileName": "Clair_Obscure_Expedition_.mp4"
}
```

Documento `.pdf`:

```json
{
  "number": "5500000000001@s.whatsapp.net",
  "mediatype": "document",
  "mimetype": "application/pdf",
  "caption": "envio de arquivos como documento",
  "media": "JVBERi0xLjcK...",
  "fileName": "ML-INFORME-RENDIMENTOS-2025 (2).pdf"
}
```

## Respostas da Evolution

A Evolution retorna `status: "PENDING"` quando aceita o envio. O campo `message` varia conforme o tipo:

- texto: `message.conversation`
- imagem: `message.imageMessage`
- vídeo: `message.videoMessage`
- documento: `message.documentMessage`

Exemplo simplificado:

```json
{
  "key": {
    "remoteJid": "5500000000001@s.whatsapp.net",
    "fromMe": true,
    "id": "3EB020BC7FEB8824CB9BC4"
  },
  "status": "PENDING",
  "message": {
    "documentMessage": {
      "mimetype": "application/pdf",
      "fileName": "arquivo.pdf",
      "caption": "texto"
    }
  },
  "messageType": "documentMessage",
  "messageTimestamp": 1776300800
}
```

O backend trata `message` e `messageTimestamp` de forma flexível, porque a Evolution retorna objetos diferentes por tipo de mídia e timestamp numérico.

## Entrada pelo Backend

Exemplo de `POST /message/send` usando WhatsApp:

```json
{
  "whatsapp_id": "uuid-da-instancia-whatsapp",
  "subject": "Aviso importante",
  "body": "Segue o informe de rendimentos.",
  "to": ["uuid-do-aluno"],
  "attachments": [
    {
      "fileName": "informe.pdf",
      "data": "JVBERi0xLjcK..."
    }
  ]
}
```

Também é possível usar URL em anexos:

```json
{
  "fileName": "arquivo.pdf",
  "url": "https://exemplo.edu.br/arquivo.pdf"
}
```

Para anexos por URL, a URL precisa estar acessível pela Evolution API.
