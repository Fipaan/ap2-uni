CREATE OR REPLACE FUNCTION notify_order_status_change()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify(
        'order_updates',
        json_build_object(
            'order_id', NEW.id,
            'status', NEW.status,
            'updated_at', to_char(NOW(), 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS orders_status_notify ON orders;

CREATE TRIGGER orders_status_notify
AFTER INSERT OR UPDATE OF status ON orders
FOR EACH ROW
EXECUTE FUNCTION notify_order_status_change();
