CREATE SCHEMA IF NOT EXISTS identity_;
CREATE SCHEMA IF NOT EXISTS logistics_;
CREATE SCHEMA IF NOT EXISTS mobility_;
CREATE SCHEMA IF NOT EXISTS payment_;
CREATE SCHEMA IF NOT EXISTS operations_;

DO
$$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'identity_user') THEN
        CREATE USER identity_user WITH PASSWORD 'identity_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'logistics_user') THEN
        CREATE USER logistics_user WITH PASSWORD 'logistics_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'mobility_user') THEN
        CREATE USER mobility_user WITH PASSWORD 'mobility_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'payment_user') THEN
        CREATE USER payment_user WITH PASSWORD 'payment_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'operations_user') THEN
        CREATE USER operations_user WITH PASSWORD 'operations_pass';
    END IF;
END
$$;

GRANT ALL ON SCHEMA identity_ TO identity_user;
GRANT ALL ON SCHEMA logistics_ TO logistics_user;
GRANT ALL ON SCHEMA mobility_ TO mobility_user;
GRANT ALL ON SCHEMA payment_ TO payment_user;
GRANT ALL ON SCHEMA operations_ TO operations_user;

ALTER USER identity_user SET search_path = identity_;
ALTER USER logistics_user SET search_path = logistics_;
ALTER USER mobility_user SET search_path = mobility_;
ALTER USER payment_user SET search_path = payment_;
ALTER USER operations_user SET search_path = operations_;
