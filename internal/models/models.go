// Package models contains the GORM entities mapped to the PostgreSQL schema
// defined in schema.sql. Each struct mirrors one table; the TableName methods
// pin the (lowercase) table names so GORM does not pluralize them.
package models

import "time"

// Adherant represents a library member.
type Adherant struct {
	CodeAdherant int        `gorm:"primaryKey;column:code_adherant"`
	Status       string     `gorm:"column:status;not null"`
	Nom          string     `gorm:"column:nom;not null"`
	Prenom       string     `gorm:"column:prenom;not null"`
	Emprunts     []Emprunts `gorm:"foreignKey:CodeAdherant"`
}

func (Adherant) TableName() string { return "adherant" }

// LivreInfo represents a book title in the library.
type LivreInfo struct {
	Isbn        string       `gorm:"primaryKey;column:isbn"`
	Titre       string       `gorm:"column:titre;not null"`
	Genre       string       `gorm:"column:genre"`
	Auteurs     []*Auteur    `gorm:"many2many:livre_auteur;"`
	Exemplaires []Exemplaire `gorm:"foreignKey:Isbn"`
}

func (LivreInfo) TableName() string { return "livreinfo" }

// Exemplaire represents a physical copy of a book in the library.
type Exemplaire struct {
	ExemplaireID int        `gorm:"primaryKey;column:exemplaire_id"`
	Isbn         string     `gorm:"column:isbn;not null"`
	Status       string     `gorm:"column:status;not null"`
	LivreInfo    LivreInfo  `gorm:"foreignKey:Isbn"`
	Emprunts     []Emprunts `gorm:"foreignKey:ExemplaireID"`
}

func (Exemplaire) TableName() string { return "exemplaire" }

// Auteur represents a book author.
type Auteur struct {
	AuteurID int          `gorm:"primaryKey;column:auteur_id"`
	Nom      string       `gorm:"column:nom;not null"`
	Prenom   string       `gorm:"column:prenom"`
	Livres   []*LivreInfo `gorm:"many2many:livre_auteur;"`
}

func (Auteur) TableName() string { return "auteur" }

// Emprunts represents a loan record. The primary key is composite
// (code_adherant, exemplaire_id, date_pret) so the same member can borrow the
// same copy on different dates, keeping a full loan history.
type Emprunts struct {
	CodeAdherant int        `gorm:"primaryKey;column:code_adherant"`
	ExemplaireID int        `gorm:"primaryKey;column:exemplaire_id"`
	DatePret     time.Time  `gorm:"primaryKey;column:date_pret;type:date"`
	DateRetour   *time.Time `gorm:"column:date_retour;type:date"`
	Adherant     Adherant   `gorm:"foreignKey:CodeAdherant"`
	Exemplaire   Exemplaire `gorm:"foreignKey:ExemplaireID"`
}

func (Emprunts) TableName() string { return "emprunts" }
