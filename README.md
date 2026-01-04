
## Unicast

Backend em Go para auxiliar docentes na gestão e comunicação de disciplinas. Inclui autenticação, cadastro de campus/cursos/disciplinas, alunos pré-cadastrados por disciplina (enrollments), convites públicos com código curto para auto-cadastro do aluno, integrações de SMTP e WhatsApp, e documentação Swagger.

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
- **Invites**: professor cria código curto para a disciplina (`POST /invite/:courseId`); aluno usa `POST /invite/:code/self-register` com `studentId`, `name`, `phone`, `email`. Backend valida vínculo (enrollment) e status `PENDING` antes de ativar.
- **SMTP/WhatsApp**: criação/listagem de instâncias de envio.

### Migrations
Arquivo SQL em `migrations/`. Exemplo com golang-migrate:
```
migrate -path migrations -database "$POSTGRES_DATABASE_URL" up
```

### Referências úteis
- Swagger gerado em `docs/` (origem: `cmd/main/main.go` via `swag init`).
- Banco: migrations incluem `invites`, `enrollments`, `students`, `courses`, `programs`, `campuses`, `users`, `smtp_instances`, `whatsapp_instances`.

### To-do / Roadmap
- Verifique o board no [Notion](https://www.notion.so/1c702239900d80b7b24dc911e23ed2a4?v=1c702239900d8012923e000c184e26af).
