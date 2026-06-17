
DROP TRIGGER before_update_orders ON orders;

DROP FUNCTION set_updated_at;


DROP INDEX IF EXISTS orders_sorted_updated_at_index;

ALTER TABLE orders
DROP COLUMN updated_at;