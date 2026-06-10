# Library Management

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne-5C9BF5?style=flat)
![GORM](https://img.shields.io/badge/GORM-00ACD7?style=flat)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)

A desktop library management application built with Go and Fyne, using PostgreSQL as the database backend.

---

## Screenshots

| Login | Overdue Books |
|:---------:|:----------------:|
| ![Login](assets/screenshot_connexion.png) | ![Overdue](assets/screenshot_retard.png) |

| Loans per Book | Member Status |
|:-----------------:|:--------------------:|
| ![Loans](assets/screenshot_emprunts_livr.png) | ![Status](assets/screenshot_situation.png) |

---

## Features

- Connect to a PostgreSQL database via a login form
- View **overdue loans** (unreturned after 14 days)
- Rank **most popular authors** by number of loans
- Calculate **average loan duration**
- List **books never borrowed**
- **Loans by year** statistics
- **Loans by literary genre** breakdown
- **Loans per book** count
- **Member status** overview (active loans, overdue)

---

## Tech Stack

| Component | Technology |
|-----------|-------------|
| Language  | Go 1.24+    |
| UI        | [Fyne v2](https://fyne.io/) |
| ORM       | [GORM v1](https://gorm.io/) |
| Database  | PostgreSQL 17 |

---

## Prerequisites

1. **Go 1.24.2+** — [golang.org/dl](https://golang.org/dl/)
2. **PostgreSQL** — [postgresql.org/download](https://www.postgresql.org/download/)
3. **C compiler** (required by Fyne via CGo)
   - Windows: [MinGW-w64](https://www.mingw-w64.org/)
   - macOS: Xcode Command Line Tools (`xcode-select --install`)
   - Linux: `gcc` (`apt install gcc`)
4. **Fyne system dependencies** — see [docs.fyne.io/started](https://docs.fyne.io/started/)

---

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/NizarLakhder/library-management.git
cd library-management
```

### 2. Download Go dependencies

```bash
go mod download
```

### 3. Set up the database

```bash
psql -U postgres -c "CREATE DATABASE bibliotheque;"
psql -U postgres -d bibliotheque -f library.sql
psql -U postgres -d bibliotheque -f remplirTables.sql
```

### 4. Run the application

**Recommended** — build then run (faster startup):

```bash
go build -o bibliotheque.exe .
.\bibliotheque.exe
```

**Quick alternative** — run without building (slower startup):

```bash
go run main.go
```

---

## Login

On startup, enter your database credentials in the login form:

| Field    | Default value  |
|----------|----------------|
| Host     | `localhost`    |
| Port     | `5432`         |
| User     | `postgres`     |
| Password | `postgres`     |
| Database | `bibliotheque` |

---

## Project Structure

```
library-management/
├── assets/
│   ├── icon.png
│   ├── screenshot_connexion.png
│   ├── screenshot_retard.png
│   ├── screenshot_emprunts_livr.png
│   └── screenshot_situation.png
├── main.go
├── library.sql
├── remplirTables.sql
├── LICENSE
├── go.mod
└── go.sum
```
