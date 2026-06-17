

ALTER TABLE orders
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX orders_sorted_updated_at_index ON orders(updated_at ASC) WHERE status = 0 OR status = 1;


CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER before_update_orders
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();