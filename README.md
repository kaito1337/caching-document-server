# Документация по запуску приложения / Application Launch Guide

---
## English

### Requirements

- Docker and Docker Compose must be installed and running on your system.  
- `make` must be installed (on Windows, it is recommended to use Git Bash or WSL).

### Quick Start

1. Start the PostgreSQL database in Docker:

   ```bash
   make db-up```
2. Apply database migrations:
   ```bash
   make migrate-up
   ```
3. Run the application server:
   ```bash
   make run
   ```
## Русский

### Требования

- Docker и Docker Compose должны быть установлены и запущены на вашей системе.
- Установлен `make` (на Windows рекомендуется использовать Git Bash или WSL).

### Быстрый старт

1. Запустить базу данных PostgreSQL в Docker:

   ```bash
   make db-up
   ```
2. Применить миграции к базе данных:
   ```bash
   make migrate-up
   ```
3. Запустить сервер приложения:
   ```bash
   make run
   ```