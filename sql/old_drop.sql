-- Eliminar las tablas que dependen de 'groups' primero
DROP TABLE IF EXISTS comp_list CASCADE;
DROP TABLE IF EXISTS blacklist CASCADE;
DROP TABLE IF EXISTS events CASCADE;
-- Luego eliminar la tabla 'groups'
DROP TABLE IF EXISTS groups CASCADE;
-- Finalmente, eliminar la tabla 'active_groups'
DROP TABLE IF EXISTS active_groups CASCADE;
DROP TABLE IF EXISTS admin_panel CASCADE;
DROP TABLE IF EXISTS promotions CASCADE;