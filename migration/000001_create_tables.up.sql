
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    balance NUMERIC(19,4) NOT NULL DEFAULT 0,
    withdrawn NUMERIC(19,4) NOT NULL DEFAULT 0
);

CREATE INDEX user_login_index ON users(login);

CREATE TABLE order_statuses (
    id SERIAL PRIMARY KEY,
    status VARCHAR(15)
);

INSERT INTO order_statuses(id, status) VALUES
    (0, 'NEW'),
    (1, 'PROCESSING'),
    (2, 'INVALID'),
    (3, 'PROCESSED');



CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    number VARCHAR(32) NOT NULL UNIQUE,
    status INT NOT NULL REFERENCES order_statuses(id) DEFAULT 0,
    accrual NUMERIC(19,4) NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX orders_number_index ON orders(number);

CREATE INDEX orders_sorted_index ON orders(user_id, uploaded_at);

CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    order VARCHAR(32) NOT NULL,
    sum NUMERIC(19, 4) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX withdrawals_sorted_index ON withdrawals(user_id, processed_at);