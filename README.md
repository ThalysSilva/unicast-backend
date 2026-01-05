
## Unicast

Backend em Go para fortalecer a comunicação docente–discente. Permite ao professor cadastrar disciplinas e alunos (pré-cadastrados por matrícula) e enviar mensagens que chegam por múltiplos canais (WhatsApp e e-mail), reduzindo o risco de a informação passar despercebida. Inclui autenticação, gestão de campus/curso/disciplina, convites públicos com código curto para o auto-cadastro do aluno, e integrações de SMTP e WhatsApp.

### Stack
- Go 1.24.x (Gin, Swagger, JWT/JWE, PQ)
- PostgreSQL (persistência principal)
- Redis + Evolution API (para WhatsApp)
- Docker Compose para ambiente local (postgres, redis, mongo, evolution, pgadmin)

### Estrutura (resumo)
- `cmd/main/main.go`: inicialização, DI dos repositórios/serviços e rotas.
- `internal/*`: módulos de domínio (auth, campus, program, course, student, enrollment, invite, smtp, whatsapp, user).
- `pkg/database`: abstrações de transação e helpers SQL.
- `migrations/`: migrações SQL (Postgres).
- `docs/`: documentação Swagger gerada pelo `swag`.

### Pré-requisitos
- Go 1.24.x instalado.
- `swag` CLI para gerar Swagger: `go install github.com/swaggo/swag/cmd/swag@latest`.
- (Opcional, hot reload) `air`: `go install github.com/air-verse/air@latest`.
- Docker e Docker Compose se for subir os serviços auxiliares.

### Configuração de ambiente
Crie um `.env` (ou `.env.development`) na raiz seguindo o `example.env`

> Dica: converta o `.env` para formato Unix se estiver no WSL: `dos2unix .env`.

### Subir dependências com Docker Compose (dev)
```
docker-compose -f docker-compose-dev.yaml up -d
```
Sobe Postgres, Redis, Evolution API, Mongo e PgAdmin com base nas variáveis do `.env`.

### Rodando a API
Opção 1) Script padrão (gera Swagger e executa):
```
./run.sh
```

Opção 2) Hot reload com Air (carrega .env/.env.development):
```
air
```
Swagger disponível em `http://localhost:${API_PORT}/swagger/index.html`.

### Fluxos principais
- **Auth**: `/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout` (Bearer).
- **Campus/Program/Course**: CRUD protegido; ownership validado por usuário.
- **Students**: pré-cadastro com status (PENDING, ACTIVE, etc.).
- **Enrollments**: vínculo aluno ↔ disciplina.
- **Invites**: professor cria código curto para a disciplina (`POST /invite/:courseId`); aluno usa `POST /invite/self-register/:code` com `studentId`, `name`, `phone`, `email`. Backend valida vínculo (enrollment) e status `PENDING` antes de ativar.
- **Importação de alunos**: `POST /course/:courseId/students/import?mode=upsert|clean` (CSV multipart em `file`). Colunas aceitas: `studentId` (obrigatória), `name`, `phone`, `email`, `status` (1/2/3/4/5 ou ACTIVE/LOCKED/GRADUATED/CANCELED/PENDING). `mode=clean` remove matrículas do curso antes de inserir. Regras: se o aluno não existir, apenas o `studentId` é salvo com status `PENDING`; status pode ser atualizado sempre; dados de contato só são atualizados se o aluno já tiver algum contato salvo (cadastro próprio); contatos enviados para quem nunca se cadastrou são ignorados e logados.
- **SMTP/WhatsApp**: criação/listagem de instâncias de envio.
- **WhatsApp Instâncias**: além do CRUD de instâncias, expõe connect/status/logout/restart; criação já retorna QR/pairing code para parear.
- **Mensagens**: `POST /message/send` envia e-mail e WhatsApp para alunos; logs de entrega ficam em `message_logs`.
- **Backdoor admin**: `POST /backdoor/reset-password` com `ADMIN_SECRET` permite reset de senha por `userId` ou `email` para recuperar acesso.

### Segurança e credenciais
- **Tokens**: JWT para acesso/refresh; JWE com chave de 32 bytes hex para proteger tokens sensíveis.
- **SMTP**: credenciais armazenadas com criptografia (ver `internal/encryption` / `smtp`), usando `JWE_SECRET` para cifrar dados sensíveis antes de persistir.
- **Env vars**: segredos ficam no `.env`/`.env.development`. Não commitá-los; use `example.env` como base.
- **Ownership**: operações sensíveis (campus/program/course/invite) conferem o `userID` do token ao dono do recurso.
- **Invite codes**: códigos curtos únicos por disciplina; validados como ativos/não expirados e vinculados ao enrollment, garantindo que apenas alunos pré-cadastrados possam ativar seus dados.
- **Backdoor**: rota administrativa protegida por `ADMIN_SECRET`; trate essa chave como segredo crítico.
  
#### Modelo de criptografia SMTP
- Cada instância SMTP é criada pelo usuário fornecendo um segredo (`smtpSecret`) próprio; esse segredo não é armazenado em texto plano.
- As credenciais SMTP (email, senha, host/porta, IV) são cifradas com JWE usando o `smtpSecret` do usuário, de modo que um vazamento de banco afeta apenas a instância/usuário que teve o segredo comprometido.
- O `JWE_SECRET` global serve apenas para proteger dados sensíveis de tokens/JWE do sistema; o segredo específico de SMTP é fornecido pelo usuário, reduzindo o blast radius.
- Logs não carregam dados sensíveis; recomenda-se nunca registrar host/usuário/senha do SMTP em claro.

### Diagramas (Mermaid)

**Arquitetura geral**
```mermaid
flowchart LR
    subgraph Frontend
        FE["Cliente Web"]
    end
    subgraph Backend
        API["UniCast API (Gin)"]
        Auth["Auth/JWT/JWE"]
        Msg["Message Service"]
        Inv["Invite/Enrollment"]
        SMTP["SMTP Service"]
        WA["WhatsApp Service"]
        Log["Message Logs"]
    end
    subgraph Infra
        PG["PostgreSQL"]
        Evo["Evolution API\n(+ Redis interno)"]
        Mail["SMTP Provider"]
    end

    FE -->|HTTP/JSON| API
    API --> Auth
    API --> Inv
    API --> Msg
    API --> SMTP
    API --> WA
    Msg --> Log
    Auth --> PG
    Inv --> PG
    Msg --> PG
    Log --> PG
    SMTP --> PG
    WA --> PG
    WA -->|send/receive| Evo
    SMTP -->|send| Mail
```

**Fluxo de auto-cadastro via invite**
```mermaid
sequenceDiagram
    participant Professor
    participant API
    participant DB as Postgres
    participant Aluno

    Professor->>API: POST /invite/:courseId (Bearer)
    API->>DB: cria invite (code, courseId, expiração)
    API-->>Professor: code
    Aluno->>API: POST /invite/self-register/:code {studentId, name, phone, email}
    API->>DB: valida invite + enrollment PENDING
    API->>DB: atualiza aluno (contatos, status ACTIVE)
    API-->>Aluno: mensagem de sucesso
```

**Fluxo de envio de mensagem**
```mermaid
sequenceDiagram
    participant Professor
    participant API
    participant DB as Postgres
    participant Evo as Evolution API
    participant Mail as SMTP Provider

    Professor->>API: POST /message/send (Bearer)
    API->>DB: resolve alunos/contatos e instâncias (SMTP/WA)
    API->>Mail: envia email
    API->>Evo: envia WhatsApp (texto/media)
    API->>DB: grava message_logs (status, canais, destinatários)
    API-->>Professor: resposta com falhas por canal (se houver)
```

**Criptografia de credenciais SMTP**
```mermaid
flowchart TD
    subgraph Entrada
      SmtpSecret["smtpSecret (fornecido pelo usuário)"]
      Creds["Credenciais SMTP (email, senha, host, port)"]
    end

    subgraph Criptografia
      JWE["JWE Encrypt/Decrypt"]
    end

    subgraph Persistência
      DB["smtp_instances (ciphertext + iv)"]
    end

    subgraph Uso
      SMTPClient["SMTP Client (em memória)"]
      Mail["SMTP Provider"]
    end

    Creds --> JWE
    SmtpSecret --> JWE
    JWE -->|gera ciphertext| DB
    DB -->|ciphertext| JWE
    SmtpSecret --> JWE
    JWE -->|dados em claro em memória| SMTPClient
    SMTPClient --> Mail
```

**Diagrama de classes/serviços (alto nível)**
```mermaid
classDiagram
    class AuthService {
      +Register(email, password, name)
      +Login(email, password)
      +Refresh(token)
    }
    class InviteService {
      +Create(courseId, userId, expiresAt)
      +SelfRegister(code, studentId, name, phone, email)
    }
    class MessageService {
      +Send(message)
    }
    class WhatsAppService {
      +CreateInstance(userId, phone)
      +ConnectInstance(userId, instanceId)
      +SendText(to, body, instanceName)
      +SendMedia(to, caption, mimetype, mediatype, media, filename, instanceName)
    }
    class SMTPService {
      +CreateInstance(userId, jweSecret, smtpSecret, email, password, host, port)
      +GetInstances(userId)
    }
    class StudentService {
      +Create(studentId, name, phone, email, annotation, status)
      +Update(id, fields)
      +ImportForCourse(courseId, mode, records)
      +GetStudents(filters)
    }
    class EnrollmentRepo {
      +FindByCourseAndStudent(...)
      +DeleteByCourseID(...)
      +Create(...)
    }
    class StudentRepo {
      +Create(...)
      +Update(...)
      +FindByID(...)
      +FindByStudentID(...)
    }
    class CourseRepo {
      +FindByProgramID(...)
      +Update(...)
      +Delete(...)
    }

    MessageService --> WhatsAppService
    MessageService --> SMTPService
    InviteService --> EnrollmentRepo
    InviteService --> StudentService
    StudentService --> StudentRepo
    StudentService --> EnrollmentRepo
```

**Estados de Invite e Student (simplificado)**
```mermaid
stateDiagram-v2
    [*] --> InviteActive
    InviteActive --> InviteUsed: self-register
    InviteActive --> InviteExpired: expiresAt (opcional)
    InviteUsed --> [*]
    InviteExpired --> [*]

    [*] --> StudentPending
    StudentPending --> StudentActive: self-register + consent
    StudentActive --> StudentLocked: admin change / status import
    StudentActive --> StudentCanceled: admin change / status import
    StudentLocked --> StudentActive: admin reativação
```

**Entidades principais (ER simplificado)**
```mermaid
erDiagram
    USER ||--o{ CAMPUS : owns
    USER ||--o{ WHATSAPP_INSTANCE : owns
    USER ||--o{ SMTP_INSTANCE : owns
    USER ||--o{ MESSAGE_LOG : sends

    CAMPUS ||--o{ PROGRAM : contains
    PROGRAM ||--o{ COURSE : contains
    COURSE ||--o{ INVITE : issues
    COURSE ||--o{ ENROLLMENT : has

    STUDENT ||--o{ ENROLLMENT : participates
    STUDENT ||--o{ MESSAGE_LOG : receives

    WHATSAPP_INSTANCE ||--o{ MESSAGE_LOG : delivers
    SMTP_INSTANCE ||--o{ MESSAGE_LOG : delivers

    USER {
        string id
        string name
        string email
    }
    CAMPUS {
        string id
        string name
    }
    PROGRAM {
        string id
        string name
        string campus_id
    }
    COURSE {
        string id
        string name
        string program_id
        int    year
        int    semester
    }
    STUDENT {
        string id
        string student_id
        string status
    }
    ENROLLMENT {
        string id
        string course_id
        string student_id
    }
    INVITE {
        string code
        string course_id
        datetime expires_at
    }
    WHATSAPP_INSTANCE {
        string id
        string instance_name
        string phone
    }
    SMTP_INSTANCE {
        string id
        string email
        string host
        int    port
        string iv
    }
    MESSAGE_LOG {
        string id
        string channel
        string status
        string recipient
    }
```

**Atores e fluxos principais**
```mermaid
flowchart TB
    Professor["Professor/Coordenador"]
    Aluno["Aluno"]
    Admin["Backdoor (Admin)"]

    C1["Gerir campus/program/curso"]
    C2["Importar alunos / criar invite"]
    C3["Auto-cadastro via invite"]
    C4["Enviar mensagens (email/WhatsApp)"]
    C5["Gerir instâncias SMTP/WhatsApp"]
    C6["Reset de senha (backdoor)"]

    Professor --> C1
    Professor --> C2
    Professor --> C4
    Professor --> C5
    Aluno --> C3
    Admin --> C6
```

### Migrations
Arquivo SQL em `migrations/`. Exemplo com golang-migrate:
```
migrate -path migrations -database "$POSTGRES_DATABASE_URL" up
```

### Referências úteis
- Swagger gerado em `docs/` (origem: `cmd/main/main.go` via `swag init`).
- Banco: migrations incluem `invites`, `enrollments`, `students`, `courses`, `programs`, `campuses`, `users`, `smtp_instances`, `whatsapp_instances`.

### Fluxo de uso rápido
1. Preencha `.env`/`.env.development` conforme `example.env`.
2. `docker-compose -f docker-compose-dev.yaml up -d` para dependências (Postgres, Redis, Evolution, Mongo, PgAdmin).
3. `./run.sh` ou `air` (hot reload) para subir a API.
4. Gere/swagger se precisar: `swag init -g cmd/main/main.go` (ou use o gerado em `docs/`).
5. Use `/auth/register` e `/auth/login` para obter tokens e chamar os demais endpoints protegidos (Bearer).
6. Cadastre campus/program/course; crie instâncias de SMTP/WhatsApp; crie invites para disciplinas; importe matrículas por curso se quiser (`/course/:courseId/students/import`); alunos finalizam o cadastro via invite `POST /invite/self-register/:code`.

### To-do / Roadmap
- Verifique o board no [Notion](https://www.notion.so/1c702239900d80b7b24dc911e23ed2a4?v=1c702239900d8012923e000c184e26af).
