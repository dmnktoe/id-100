-- deriven_rows.sql
-- Placeholder file for Supabase export of deriven table
-- 
-- INSTRUCTIONS:
-- 1. Export your deriven table from Supabase
-- 2. Replace this file with your export
-- 3. Run: ./scripts/convert-deriven-export.sh deriven_rows.sql
-- 4. Restart Docker: docker-compose down -v && docker-compose up -d --build
--
-- EXPECTED FORMAT:
-- Your Supabase export should contain INSERT statements like:
--
-- INSERT INTO "public"."deriven" ("id", "number", "title", "description", "created_at", "points") VALUES
-- ('1', '1', 'Derive #001', 'Dokumentiere ein Objekt, das deiner aktuellen Stimmung entspricht.', '2025-12-30 12:17:45.375781+00', '1'),
-- ('2', '2', 'Derive #002', 'Miss die Höhe von fünf Bordsteinkanten.', '2025-12-30 12:17:45.375781+00', '2'),
-- ('3', '3', 'Derive #003', 'Gestalte eine Postkarte für deine Innenstadt.', '2025-12-30 12:17:45.375781+00', '3');
--
-- NOTE: The conversion script will automatically:
-- - Convert "public"."deriven" to deriven
-- - Wrap the INSERT in a conditional block to prevent duplicates
-- - Generate the proper migration file
--
-- After replacing this file, the script is ready to run!

-- PLACEHOLDER DATA (for testing only - replace with your actual Supabase export)
INSERT INTO "public"."deriven" ("id", "number", "title", "description", "created_at", "points") VALUES
('1', '1', 'Derive #001', 'Placeholder derive - replace with your data', '2025-12-30 12:17:45.375781+00', '1');
