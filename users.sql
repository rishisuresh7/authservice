CREATE TABLE IF NOT EXISTS users (
    p_id serial NOT NULL,
    id character varying(100) NOT NULL,
    first_name character varying(50) DEFAULT '',
    last_name character varying(50) DEFAULT '',
    phone character varying(15),
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_phone_key UNIQUE (phone)
)
