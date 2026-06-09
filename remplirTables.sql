-- =============================================================
-- Données de test — Gestion Bibliothèque
-- Cours  : IFT2935 — Bases de données (Hiver 2025)
-- =============================================================
-- Scénarios couverts par ces données :
--   - 2 emprunts en retard (date_pret > 14 jours, date_retour NULL)
--   - 1 livre jamais emprunté (Notre-Dame de Paris)
--   - 1 emprunt retourné (pour la durée moyenne)
--   - Victor Hugo associé à deux livres (test auteurs populaires)
-- =============================================================

INSERT INTO livreinfo (isbn, titre, genre) VALUES
    ('978-2-07-036024-5', 'Les Miserables',      'Roman historique'),
    ('978-2-07-040850-4', 'Le Petit Prince',      'Conte philosophique'),
    ('978-2-07-036822-7', 'Madame Bovary',        'Roman realiste'),
    ('978-2-07-041239-6', 'L Etranger',           'Roman philosophique'),
    ('978-2-07-036025-2', 'Notre-Dame de Paris',  'Roman historique');

INSERT INTO auteur (nom, prenom) VALUES
    ('Hugo',          'Victor'),
    ('Saint-Exupery', 'Antoine de'),
    ('Flaubert',      'Gustave'),
    ('Camus',         'Albert');

INSERT INTO livre_auteur (isbn, auteur_id) VALUES
    ('978-2-07-036024-5', 1),
    ('978-2-07-040850-4', 2),
    ('978-2-07-036822-7', 3),
    ('978-2-07-041239-6', 4),
    ('978-2-07-036025-2', 1);

INSERT INTO exemplaire (isbn, status) VALUES
    ('978-2-07-036024-5', 'disponible'),
    ('978-2-07-036024-5', 'emprunte'),
    ('978-2-07-040850-4', 'disponible'),
    ('978-2-07-036822-7', 'emprunte'),
    ('978-2-07-041239-6', 'disponible'),
    ('978-2-07-036025-2', 'disponible');

INSERT INTO adherant (status, nom, prenom) VALUES
    ('actif', 'Tremblay', 'Marie'),
    ('actif', 'Gagnon',   'Pierre'),
    ('actif', 'Lakhder',  'Nizar');

INSERT INTO emprunts (code_adherant, exemplaire_id, date_pret, date_retour) VALUES
    (1, 2, '2026-05-01', '2026-05-10'),  -- retourné (durée : 9 jours)
    (2, 4, '2026-04-01', NULL),          -- en retard
    (3, 2, '2026-05-20', NULL),          -- en retard
    (1, 3, '2026-05-25', NULL);          -- en cours
