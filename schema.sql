-- =============================================================
-- Schéma de la base de données — Gestion Bibliothèque
-- Cours  : IFT2935 — Bases de données (Hiver 2025)
-- =============================================================

-- Un livre peut avoir plusieurs auteurs (relation N:N via livre_auteur).
-- Un livre peut avoir plusieurs exemplaires physiques.
-- Un emprunt porte sur un exemplaire précis, pas sur un titre.
-- La contrainte CHECK garantit qu'un livre ne peut pas être rendu
-- le jour même de son emprunt (date_retour > date_pret).

CREATE TABLE IF NOT EXISTS livreinfo (
    isbn      VARCHAR PRIMARY KEY,
    titre     VARCHAR NOT NULL,
    genre     VARCHAR
);

CREATE TABLE IF NOT EXISTS auteur (
    auteur_id SERIAL  PRIMARY KEY,
    nom       VARCHAR NOT NULL,
    prenom    VARCHAR
);

-- Table de jonction N:N entre livreinfo et auteur
CREATE TABLE IF NOT EXISTS livre_auteur (
    isbn      VARCHAR REFERENCES livreinfo(isbn),
    auteur_id INT     REFERENCES auteur(auteur_id),
    PRIMARY KEY (isbn, auteur_id)
);

CREATE TABLE IF NOT EXISTS exemplaire (
    exemplaire_id SERIAL  PRIMARY KEY,
    isbn          VARCHAR NOT NULL REFERENCES livreinfo(isbn),
    status        VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS adherant (
    code_adherant SERIAL  PRIMARY KEY,
    status        VARCHAR NOT NULL,
    nom           VARCHAR NOT NULL,
    prenom        VARCHAR NOT NULL
);

-- Clé primaire composite : un même adhérent peut emprunter le même
-- exemplaire à des dates différentes (historique complet des emprunts).
CREATE TABLE IF NOT EXISTS emprunts (
    code_adherant INT  REFERENCES adherant(code_adherant),
    exemplaire_id INT  REFERENCES exemplaire(exemplaire_id),
    date_pret     DATE NOT NULL,
    date_retour   DATE CHECK (date_retour > date_pret),
    PRIMARY KEY (code_adherant, exemplaire_id, date_pret)
);

-- Un exemplaire ne peut avoir qu'un seul emprunt OUVERT à la fois (date_retour
-- NULL). Cet index unique partiel garantit l'invariant au niveau de la base,
-- et pas seulement dans le code applicatif — y compris en cas d'accès concurrent.
CREATE UNIQUE INDEX IF NOT EXISTS uniq_emprunt_ouvert
    ON emprunts (exemplaire_id)
    WHERE date_retour IS NULL;
