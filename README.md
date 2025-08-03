# Application Launch Guide / Документация по запуску приложения

---
## English

### Requirements

- Docker and Docker Compose must be installed and running on your system.  
- `make` must be installed (on Windows, it is recommended to use Git Bash or WSL).

### Quick Start

1. Start all services (PostgreSQL, Backend):

   ```bash
   docker compose -f deployments/docker-compose.yml up -d
   ```
2. Apply database migrations:
   ```bash
   make migrate-up
   ```
3. Access the application:
   API: http://localhost:8080
## Русский

### Требования

- Docker и Docker Compose должны быть установлены и запущены на вашей системе.
- Установлен `make` (на Windows рекомендуется использовать Git Bash или WSL).

### Быстрый старт

1. Запустите все сервисы (PostgreSQL, Backend):

   ```bash
   docker compose -f deployments/docker-compose.yml up -d
   ```
2. Применить миграции к базе данных:
   ```bash
   make migrate-up
   ```
3. Доступ к сервисам:
   API: http://localhost:8080