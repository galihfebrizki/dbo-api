-- DROP SCHEMA public;

COMMENT ON SCHEMA public IS 'standard public schema';
-- public.customer_data definition

-- Drop table

-- DROP TABLE public.customer_data;

CREATE TABLE public.customer_data (
	user_id varchar(50) NOT NULL,
	dob date NULL,
	phone_number varchar(20) NULL,
	gender bpchar(1) NULL,
	marital_status varchar(10) NULL,
	address varchar(200) NULL,
	district_address varchar(30) NULL,
	city_address varchar(30) NULL,
	province_address varchar(30) NULL,
	postal_code int4 NULL,
	latitude_address varchar(30) NULL,
	longitude_address varchar(30) NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL
);


-- public.items definition

-- Drop table

-- DROP TABLE public.items;

CREATE TABLE public.items (
	id varchar(50) NOT NULL,
	item_name varchar(100) NOT NULL,
	sku varchar(30) NOT NULL,
	price int8 NOT NULL DEFAULT 0,
	quantity_type int4 NOT NULL,
	stock int4 NOT NULL DEFAULT 0,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT items_pkey PRIMARY KEY (id)
);


-- public.order_items definition

-- Drop table

-- DROP TABLE public.order_items;

CREATE TABLE public.order_items (
	id varchar(50) NOT NULL,
	order_id varchar(50) NOT NULL,
	item_id varchar(50) NOT NULL,
	quantity int4 NOT NULL,
	item_price int8 NOT NULL,
	discount_amount int8 NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT order_items_pkey PRIMARY KEY (id)
);


-- public.order_logs definition

-- Drop table

-- DROP TABLE public.order_logs;

CREATE TABLE public.order_logs (
	order_id varchar(50) NOT NULL,
	order_status int4 NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL
);


-- public.order_status definition

-- Drop table

-- DROP TABLE public.order_status;

CREATE TABLE public.order_status (
	id int4 NOT NULL,
	"name" varchar(50) NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT order_status_pkey PRIMARY KEY (id)
);


-- public.orders definition

-- Drop table

-- DROP TABLE public.orders;

CREATE TABLE public.orders (
	id varchar(50) NOT NULL,
	user_id varchar(50) NOT NULL,
	status int4 NOT NULL DEFAULT 0,
	total_amount int8 NOT NULL,
	total_quantity int4 NOT NULL,
	total_discount_amount int8 NULL DEFAULT 0,
	payment_method varchar(30) NOT NULL,
	payment_acquirement_id varchar(50) NULL,
	payment_date timestamptz NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT orders_pkey PRIMARY KEY (id)
);
CREATE INDEX orders_created_at_idx ON public.orders USING btree (created_at);
CREATE INDEX orders_user_id_idx ON public.orders USING btree (user_id);


-- public.quantity_type definition

-- Drop table

-- DROP TABLE public.quantity_type;

CREATE TABLE public.quantity_type (
	id int4 NOT NULL,
	"name" varchar(50) NOT NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT quantity_type_pkey PRIMARY KEY (id)
);


-- public.user_sessions definition

-- Drop table

-- DROP TABLE public.user_sessions;

CREATE TABLE public.user_sessions (
	user_id varchar(50) NOT NULL,
	"token" text NOT NULL,
	login_time timestamptz NULL,
	logout_time timestamptz NULL
);
CREATE INDEX user_sessions_logout_time_idx ON public.user_sessions USING btree (logout_time, user_id);
CREATE INDEX user_sessions_token_idx ON public.user_sessions USING btree (token);
CREATE INDEX user_sessions_user_id_idx ON public.user_sessions USING btree (user_id);
CREATE INDEX user_sessions_user_id_token_idx ON public.user_sessions USING btree (user_id, token);


-- public.user_status definition

-- Drop table

-- DROP TABLE public.user_status;

CREATE TABLE public.user_status (
	id int4 NOT NULL,
	"name" varchar(50) NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT user_status_pkey PRIMARY KEY (id)
);


-- public.users definition

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE public.users (
	id varchar NOT NULL,
	username varchar(50) NOT NULL,
	"password" varchar(50) NOT NULL,
	full_name varchar(100) NOT NULL,
	status int4 NULL DEFAULT 1,
	"level" int4 NULL DEFAULT 0,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);
CREATE INDEX users_full_name_idx ON public.users USING btree (full_name);
CREATE INDEX users_username_password_idx ON public.users USING btree (username, password);


-- public.customer_data foreign keys

-- public.items foreign keys

-- public.order_items foreign keys

-- public.order_logs foreign keys

-- public.order_status foreign keys

-- public.orders foreign keys

-- public.quantity_type foreign keys

-- public.user_sessions foreign keys

-- public.user_status foreign keys

-- public.users foreign keys



-- Permissions;


GRANT ALL ON SCHEMA public TO pg_database_owner;
GRANT USAGE ON SCHEMA public TO public;


INSERT INTO users (id,username,"password",full_name,status,"level",created_at,updated_at) VALUES
	 ('1638070605594742300','galih.febrizki@gmail.com','5f4dcc3b5aa765d61d8327deb882cf99','Galih Febrizki',1,0,'2023-07-19 10:19:03.043387+00',NULL);
