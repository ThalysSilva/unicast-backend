# Procedimento: OAuth de Email com Gmail no Unicast

Este documento explica como configurar o envio de email por OAuth usando Gmail/Google no Unicast.

Para a POC/TCC, o fluxo validado é:

- `Gmail OAuth`: recomendado para contas pessoais Google.
- `SMTP manual`: mantido como alternativa para emails institucionais ou provedores compatíveis.

O backend foi modelado com `provider` e `auth_mode`, então outros provedores podem ser adicionados no futuro por fork ou evolução do projeto. Nesta versão, apenas Google/Gmail fica exposto na aplicação.

## Como o Fluxo Funciona

1. O professor acessa `Integrações`.
2. Clica em `Conectar Google`.
3. O backend gera uma URL de autorização.
4. O navegador abre a tela do Google.
5. O professor autoriza o Unicast.
6. O Google redireciona para o callback do backend.
7. O backend troca o `code` por tokens.
8. O backend salva a integração criptografada.
9. O envio usa Gmail API.

## Redirect URI

Para desenvolvimento local:

```txt
http://localhost:8070/smtp/oauth/google/callback
```

Para produção:

```txt
https://api.seu-dominio.edu.br/smtp/oauth/google/callback
```

O redirect URI cadastrado no Google precisa ser exatamente igual ao valor usado no `.env`.

Confira:

- protocolo: `http` ou `https`
- host: `localhost`, domínio ou subdomínio
- porta: `8070` em dev local
- caminho completo: `/smtp/oauth/google/callback`
- ausência/presença de barra final

## Variáveis de Ambiente do Backend

Adicione ao `.env`:

```env
FRONTEND_BASE_URL=http://localhost:3000

GOOGLE_OAUTH_CLIENT_ID=
GOOGLE_OAUTH_CLIENT_SECRET=
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:8070/smtp/oauth/google/callback
```

Em produção:

```env
FRONTEND_BASE_URL=https://unicast.seu-dominio.edu.br
GOOGLE_OAUTH_REDIRECT_URL=https://api.seu-dominio.edu.br/smtp/oauth/google/callback
```

## Configuração no Google

### 1. Criar ou Selecionar o Projeto

1. Acesse o Google Cloud Console.
2. Crie um projeto ou selecione um projeto existente.
3. Use um nome identificável, por exemplo `unicast-demo`.

### 2. Habilitar a Gmail API

1. Acesse `APIs e serviços`.
2. Abra `Biblioteca`.
3. Pesquise por `Gmail API`.
4. Clique em `Ativar`.

Sem a Gmail API ativada, o token pode até ser gerado, mas o envio pode falhar.

### 3. Configurar a Tela de Consentimento

No Google Auth Platform:

1. Acesse `Branding`.
2. Preencha nome do app, email de suporte e dados obrigatórios.
3. Acesse `Público-alvo`.
4. Escolha `Externo`.

Use `Externo` para contas Google pessoais ou usuários fora de uma organização Google Workspace.

Em modo de teste, o app só funciona para usuários adicionados como testadores.

### 4. Adicionar Usuários de Teste

No Google Auth Platform:

1. Acesse `Público-alvo`.
2. Procure `Usuários de teste`.
3. Adicione o email Google que será usado no teste.

Exemplo:

```txt
seu-email@gmail.com
```

Se você não fizer isso, o Google pode mostrar:

```txt
Erro 403: access_denied
O app não concluiu o processo de verificação do Google.
```

### 5. Configurar Acesso a Dados

Este passo é obrigatório para o envio funcionar.

No Google Auth Platform:

1. Acesse `Acesso a dados`.
2. Clique em `Adicionar ou remover escopos`.
3. Adicione o escopo:

```txt
https://www.googleapis.com/auth/gmail.send
```

4. Salve.

Esse escopo permite que o Unicast envie emails pela conta conectada.

Se esse escopo não estiver configurado, o envio pode falhar com:

```txt
Request had insufficient authentication scopes.
ACCESS_TOKEN_SCOPE_INSUFFICIENT
```

### 6. Criar o Cliente OAuth

No Google Auth Platform:

1. Acesse `Clientes`.
2. Clique em `Criar um cliente OAuth`.
3. Em tipo de aplicativo, escolha `Aplicativo da Web`.
4. Nome sugerido: `Unicast Local`.
5. Em `URIs de redirecionamento autorizados`, adicione:

```txt
http://localhost:8070/smtp/oauth/google/callback
```

6. Se houver campo de origens JavaScript autorizadas, adicione:

```txt
http://localhost:3000
```

7. Salve.
8. Copie `Client ID` e `Client secret`.

### 7. Preencher o .env

```env
GOOGLE_OAUTH_CLIENT_ID=client-id-do-google
GOOGLE_OAUTH_CLIENT_SECRET=client-secret-do-google
GOOGLE_OAUTH_REDIRECT_URL=http://localhost:8070/smtp/oauth/google/callback
```

### 8. Reconectar Depois de Alterar Escopos

Importante: tokens já emitidos não recebem escopos novos automaticamente.

Se você adicionou `gmail.send` depois de já ter conectado a conta:

1. Remova a integração Google no Unicast.
2. Conecte novamente com `Conectar Google`.
3. Confira se a tela de consentimento menciona permissão para enviar email.
4. Teste o envio novamente.

## Escopos Usados pelo Backend

```txt
openid
email
https://www.googleapis.com/auth/gmail.send
```

## Rodando as Migrations

Depois de configurar o ambiente, aplique as migrations:

```bash
cd ~/github/unicast-backend

set -a
source .env
set +a

migrate \
  -path migrations \
  -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" \
  up
```

A migration que adiciona suporte a OAuth é:

```txt
000020_add_oauth_email_support
```

Ela mantém o SMTP manual e adiciona suporte a:

- `auth_mode = password`
- `auth_mode = oauth`
- `provider = custom_smtp`
- `provider = google`

## Subindo o Projeto

Backend:

```bash
cd ~/github/unicast-backend
./run.sh
```

Frontend:

```bash
cd ~/github/unicast-frontend
npm run dev
```

## Testando

1. Acesse o frontend.
2. Faça login.
3. Vá em `Integrações`.
4. Clique em `Conectar Google`.
5. Autorize a conta.
6. O Google deve redirecionar para o backend.
7. O backend deve redirecionar para `/integrations`.
8. A conta deve aparecer na lista de integrações de email com badge `OAuth`.
9. Vá em `Mensagens`.
10. Selecione aluno(s), assunto e corpo.
11. Escolha a conta Google conectada.
12. Envie.

Se funcionar, o email será enviado pela Gmail API e aparecerá como remetente da conta Google conectada.

## Problemas Comuns

### access_denied em app de teste

Mensagem típica:

```txt
Erro 403: access_denied
O app não concluiu o processo de verificação do Google.
```

Causa provável:

- O app está em modo de teste.
- O usuário ainda não foi adicionado em `Público-alvo > Usuários de teste`.

Correção:

1. Adicione o email em `Usuários de teste`.
2. Tente conectar novamente.

### insufficient authentication scopes

Mensagem típica:

```txt
Request had insufficient authentication scopes.
ACCESS_TOKEN_SCOPE_INSUFFICIENT
```

Causa provável:

- O escopo `gmail.send` não foi adicionado em `Acesso a dados`.
- A conta foi conectada antes de o escopo ser configurado.

Correção:

1. Vá em `Google Auth Platform > Acesso a dados`.
2. Adicione:

```txt
https://www.googleapis.com/auth/gmail.send
```

3. Remova a integração Google no Unicast.
4. Conecte a conta novamente.
5. Tente enviar outra mensagem.

### redirect_uri_mismatch

Causa provável:

- Redirect URI no Google diferente do `.env`.

Verifique:

```txt
http://localhost:8070/smtp/oauth/google/callback
```

### Callback falha ou token não salva

Verifique:

- `JWE_SECRET` tem 32 bytes em hex.
- `FRONTEND_BASE_URL` está correto.
- `GOOGLE_OAUTH_CLIENT_SECRET` está correto.
- A migration `000020` foi aplicada.
- O backend foi reiniciado depois de alterar `.env`.

## Removendo e Reconectando uma Integração

Use a tela de `Integrações` e clique em `Apagar` na conta conectada.

Isso remove o registro em `smtp_instances`.

Depois clique novamente em `Conectar Google`.

Remover e reconectar é necessário quando:

- escopos foram alterados
- token ficou inválido
- app OAuth foi recriado
- client secret foi trocado

## Extensibilidade para Outros Provedores

O backend foi estruturado para permitir novos provedores futuramente:

- `smtp_instances.provider` identifica o provedor.
- `smtp_instances.auth_mode` diferencia senha SMTP de OAuth.
- tokens OAuth ficam criptografados em `oauth_payload` e `oauth_iv`.
- o envio decide o caminho com base em `auth_mode` e `provider`.

Para adicionar outro provedor, será necessário:

1. Criar uma configuração de provedor OAuth no backend.
2. Adicionar as variáveis de ambiente necessárias.
3. Criar uma rota de callback.
4. Implementar o envio por API ou protocolo aceito pelo provedor.
5. Expor o botão no frontend.

## Referências Oficiais

- Google OAuth 2.0 for Web Server Applications: https://developers.google.com/identity/protocols/oauth2/web-server
- Gmail API - Sending Email: https://developers.google.com/gmail/api/guides/sending
- Google OAuth Client redirect URI rules: https://support.google.com/cloud/answer/6158849
