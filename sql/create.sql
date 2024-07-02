CREATE TABLE groups(
    id TEXT UNIQUE PRIMARY KEY,
    comp_active BOOLEAN DEFAULT FALSE,
    jetton_address TEXT NOT NULL,
    dedust_address TEXT NOT NULL,
    stonfi_address TEXT NOT NULL,
    emoji TEXT NOT NULL
);
CREATE TABLE order_buy(
    id SERIAL PRIMARY KEY,
    group_id TEXT REFERENCES groups(id),
    jetton_address TEXT NOT NULL,
    jetton_name TEXT NOT NULL,
    jetton_symbol TEXT NOT NULL,
    jetton_decimal NUMERIC NOT NULL,
    buyer_address TEXT NOT NULL,
    ton_amount NUMERIC NOT NULL,
    token_amount NUMERIC NOT NULL
);
CREATE TABLE order_sell(
    id SERIAL PRIMARY KEY,
    group_id TEXT REFERENCES groups(id),
    jetton_address TEXT NOT NULL,
    jetton_name TEXT NOT NULL,
    jetton_symbol TEXT NOT NULL,
    jetton_decimal TEXT NOT NULL,
    seller_address TEXT NOT NULL,
    ton_amount NUMERIC NOT NULL,
    token_amount NUMERIC NOT NULL
);
CREATE TABLE promo(
    id TEXT UNIQUE NOT NULL,
    ad_text TEXT NOT NULL,
    button_name TEXT NOT NULL,
    button_link TEXT NOT NULL,
    media TEXT NOT NULL
);
CREATE TABLE end_time(
    id TEXT UNIQUE PRIMARY KEY REFERENCES groups(id),
    timestamp NUMERIC NOT NULL
);