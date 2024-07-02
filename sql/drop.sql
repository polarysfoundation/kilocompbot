DROP TABLE IF EXISTS promo CASCADE;

-- Eliminar las tablas que dependen de 'groups' primero
DROP TABLE IF EXISTS order_buy CASCADE;
DROP TABLE IF EXISTS order_sell CASCADE;
-- Finalmente, eliminar la tabla 'active_groups'
DROP TABLE IF EXISTS end_time CASCADE;
DROP TABLE IF EXISTS groups CASCADE;