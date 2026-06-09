# Gestion Bibliothèque — IFT2935

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql&logoColor=white)
![Fyne](https://img.shields.io/badge/Fyne-5C9BF5?style=flat)
![GORM](https://img.shields.io/badge/GORM-00ACD7?style=flat)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)

Application de gestion de bibliothèque développée dans le cadre du cours **IFT2935 — Bases de données** (Hiver 2025). Elle permet d'interroger une base PostgreSQL via une interface graphique native construite avec Fyne.

---

## Aperçu

| Connexion | Livres en retard |
|:---------:|:----------------:|
| ![Connexion](assets/screenshot_connexion.png) | ![Retard](assets/screenshot_retard.png) |

| Emprunts par livre | Situation des abonnés |
|:-----------------:|:--------------------:|
| ![Emprunts](assets/screenshot_emprunts_livr.png) | ![Situation](assets/screenshot_situation.png) |


---

## Fonctionnalités

- Connexion à une base PostgreSQL via un formulaire
- Affichage des **emprunts en retard** (> 14 jours sans retour)
- Classement des **auteurs les plus populaires** par nombre d'emprunts
- Calcul de la **durée moyenne des emprunts**
- Liste des **livres jamais empruntés**
- Statistiques des **emprunts par année**
- **Répartition des emprunts par genre** littéraire
- Nombre d'**emprunts par livre**
- **Situation de chaque abonné** (livres en cours, retards)

---

## Stack technique

| Composant | Technologie |
|-----------|-------------|
| Langage   | Go 1.24+    |
| Interface | [Fyne v2](https://fyne.io/) |
| ORM       | [GORM v1](https://gorm.io/) |
| Base de données | PostgreSQL 17 |

---

## Prérequis

1. **Go 1.24.2+** — [golang.org/dl](https://golang.org/dl/)
2. **PostgreSQL** — [postgresql.org/download](https://www.postgresql.org/download/)
3. **Compilateur C** (requis par Fyne via CGo)
   - Windows : [MinGW-w64](https://www.mingw-w64.org/)
   - macOS : Xcode Command Line Tools (`xcode-select --install`)
   - Linux : `gcc` (`apt install gcc`)
4. **Dépendances système Fyne** — voir [docs.fyne.io/started](https://docs.fyne.io/started/)

---

## Installation

### 1. Cloner le dépôt

```bash
git clone https://github.com/NizarLakhder/IFT2935Projet.git
cd IFT2935Projet
```

### 2. Télécharger les dépendances Go

```bash
go mod download
```

### 3. Créer et configurer la base de données

```bash
psql -U postgres -c "CREATE DATABASE bibliotheque;"
psql -U postgres -d bibliotheque -f library.sql
psql -U postgres -d bibliotheque -f remplirTables.sql
```

### 4. Lancer l'application

**Méthode recommandée** — compiler puis exécuter (lancement instantané) :

```bash
go build -o bibliotheque.exe .
.\bibliotheque.exe
```

**Alternative rapide** — sans compilation (plus lent au démarrage) :

```bash
go run main.go
```

---

## Connexion

Au lancement, entrer les informations de connexion dans le formulaire :

| Champ          | Valeur par défaut |
|----------------|-------------------|
| Hôte           | `localhost`       |
| Port           | `5432`            |
| Utilisateur    | `postgres`        |
| Mot de passe   | `postgres`        |
| Base de données| `bibliotheque`    |

---

## Structure du projet

```
IFT2935Projet/
├── assets/
│   ├── icon.png                     # Icône de l'application
│   ├── screenshot_connexion.png
│   ├── screenshot_retard.png
│   ├── screenshot_emprunts_livr.png
│   └── screenshot_situation.png
├── main.go                          # Code source principal
├── library.sql                      # Schéma de la base de données
├── remplirTables.sql                # Données de test
├── LICENSE
├── go.mod
└── go.sum
```
