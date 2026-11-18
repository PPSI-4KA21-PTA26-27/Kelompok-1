-- Skema Sprint 1 (SDD Bagian 10.2): 11 tabel inti sesuai kamus data SDD Bagian 5.3.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE cart_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    quantity integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_cart_items_quantity CHECK ((quantity > 0))
);

CREATE TABLE categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid NOT NULL,
    product_name text NOT NULL,
    size text NOT NULL,
    unit_price numeric NOT NULL,
    quantity integer NOT NULL,
    subtotal numeric NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    color text,
    CONSTRAINT ck_order_items_quantity CHECK ((quantity > 0)),
    CONSTRAINT ck_order_items_subtotal CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT ck_order_items_unit_price CHECK ((unit_price >= (0)::numeric))
);

CREATE TABLE orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_number text NOT NULL,
    user_id uuid NOT NULL,
    recipient_name text NOT NULL,
    recipient_phone text NOT NULL,
    shipping_address text NOT NULL,
    shipping_city text NOT NULL,
    shipping_kelurahan text NOT NULL,
    subtotal numeric NOT NULL,
    shipping_cost numeric NOT NULL,
    total_amount numeric NOT NULL,
    order_status text NOT NULL,
    shipping_district text NOT NULL,
    shipping_subdistrict_id integer NOT NULL,
    shipping_zip_code text,
    courier_code text NOT NULL,
    courier_service text NOT NULL,
    shipping_etd text,
    shipping_weight_gram integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    payment_due_at timestamp with time zone,
    cancel_reason text,
    CONSTRAINT ck_orders_shipping_cost CHECK ((shipping_cost >= (0)::numeric)),
    CONSTRAINT ck_orders_shipping_weight CHECK (((shipping_weight_gram IS NULL) OR (shipping_weight_gram >= 0))),
    CONSTRAINT ck_orders_subtotal CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT ck_orders_total_amount CHECK ((total_amount >= (0)::numeric))
);

CREATE TABLE payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    method text NOT NULL,
    amount numeric NOT NULL,
    unique_code integer NOT NULL,
    status text NOT NULL,
    paid_at timestamp with time zone,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    proof_url text,
    CONSTRAINT ck_payments_amount CHECK ((amount >= (0)::numeric)),
    CONSTRAINT ck_payments_status CHECK ((status = ANY (ARRAY['menunggu'::text, 'valid'::text, 'ditolak'::text]))),
    CONSTRAINT ck_payments_unique_code CHECK ((unique_code >= 0))
);

CREATE TABLE product_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    image_url text NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE product_variants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    sku text NOT NULL,
    size text NOT NULL,
    color text,
    stock integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_product_variants_stock CHECK ((stock >= 0))
);

CREATE TABLE products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid,
    name text NOT NULL,
    description text,
    price numeric NOT NULL,
    discount_percent numeric DEFAULT 0 NOT NULL,
    weight_gram integer NOT NULL,
    availability_type text NOT NULL,
    po_estimate_days integer,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_products_discount CHECK (((discount_percent >= (0)::numeric) AND (discount_percent <= (100)::numeric))),
    CONSTRAINT ck_products_po_days CHECK (((po_estimate_days IS NULL) OR (po_estimate_days >= 0))),
    CONSTRAINT ck_products_price CHECK ((price >= (0)::numeric)),
    CONSTRAINT ck_products_weight CHECK ((weight_gram > 0))
);

CREATE TABLE roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    CONSTRAINT ck_roles_code CHECK ((code = ANY (ARRAY['customer'::text, 'staff'::text, 'owner'::text])))
);

CREATE TABLE shipments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    tracking_number text,
    coverage_status text NOT NULL,
    shipping_status text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    phone text,
    status text DEFAULT 'active'::text NOT NULL,
    email_verified_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    role_id uuid NOT NULL,
    CONSTRAINT ck_users_status CHECK ((status = ANY (ARRAY['active'::text, 'inactive'::text])))
);

ALTER TABLE ONLY cart_items
    ADD CONSTRAINT pk_cart_items PRIMARY KEY (id);

ALTER TABLE ONLY categories
    ADD CONSTRAINT pk_categories PRIMARY KEY (id);

ALTER TABLE ONLY order_items
    ADD CONSTRAINT pk_order_items PRIMARY KEY (id);

ALTER TABLE ONLY orders
    ADD CONSTRAINT pk_orders PRIMARY KEY (id);

ALTER TABLE ONLY payments
    ADD CONSTRAINT pk_payments PRIMARY KEY (id);

ALTER TABLE ONLY product_images
    ADD CONSTRAINT pk_product_images PRIMARY KEY (id);

ALTER TABLE ONLY product_variants
    ADD CONSTRAINT pk_product_variants PRIMARY KEY (id);

ALTER TABLE ONLY products
    ADD CONSTRAINT pk_products PRIMARY KEY (id);

ALTER TABLE ONLY roles
    ADD CONSTRAINT pk_roles PRIMARY KEY (id);

ALTER TABLE ONLY shipments
    ADD CONSTRAINT pk_shipments PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT pk_users PRIMARY KEY (id);

ALTER TABLE ONLY cart_items
    ADD CONSTRAINT uq_cart_items_user_variant UNIQUE (user_id, variant_id);

ALTER TABLE ONLY categories
    ADD CONSTRAINT uq_categories_name UNIQUE (name);

ALTER TABLE ONLY orders
    ADD CONSTRAINT uq_orders_order_number UNIQUE (order_number);

ALTER TABLE ONLY product_variants
    ADD CONSTRAINT uq_product_variants_sku UNIQUE (sku);

ALTER TABLE ONLY roles
    ADD CONSTRAINT uq_roles_code UNIQUE (code);

ALTER TABLE ONLY shipments
    ADD CONSTRAINT uq_shipments_order UNIQUE (order_id);

ALTER TABLE ONLY users
    ADD CONSTRAINT uq_users_email UNIQUE (email);

CREATE INDEX idx_cart_items_user_id ON cart_items USING btree (user_id);

CREATE INDEX idx_cart_items_variant_id ON cart_items USING btree (variant_id);

CREATE INDEX idx_order_items_order_id ON order_items USING btree (order_id);

CREATE INDEX idx_order_items_product_id ON order_items USING btree (product_id);

CREATE INDEX idx_order_items_variant_id ON order_items USING btree (variant_id);

CREATE INDEX idx_orders_created ON orders USING btree (created_at DESC);

CREATE INDEX idx_orders_status_due ON orders USING btree (order_status, payment_due_at);

CREATE INDEX idx_orders_user_number ON orders USING btree (user_id, order_number DESC);

CREATE INDEX idx_payments_order_id ON payments USING btree (order_id);

CREATE INDEX idx_payments_reviewed_by ON payments USING btree (reviewed_by);

CREATE INDEX idx_payments_status ON payments USING btree (status);

CREATE INDEX idx_product_images_primary ON product_images USING btree (product_id, is_primary DESC, sort_order, id);

CREATE INDEX idx_product_variants_product_id ON product_variants USING btree (product_id);

CREATE INDEX idx_products_category_id ON products USING btree (category_id);

CREATE INDEX idx_products_name ON products USING btree (name, id);

CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);

CREATE INDEX idx_products_status_created ON products USING btree (status, created_at DESC, id);

CREATE INDEX idx_products_status_name ON products USING btree (status, name, id);

CREATE INDEX idx_users_role_id ON users USING btree (role_id);

ALTER TABLE ONLY cart_items
    ADD CONSTRAINT fk_cart_items_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY cart_items
    ADD CONSTRAINT fk_cart_items_variant FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE;

ALTER TABLE ONLY order_items
    ADD CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;

ALTER TABLE ONLY order_items
    ADD CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT;

ALTER TABLE ONLY order_items
    ADD CONSTRAINT fk_order_items_variant FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE RESTRICT;

ALTER TABLE ONLY orders
    ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY payments
    ADD CONSTRAINT fk_payments_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;

ALTER TABLE ONLY payments
    ADD CONSTRAINT fk_payments_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE ONLY product_images
    ADD CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE ONLY product_variants
    ADD CONSTRAINT fk_product_variants_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE ONLY products
    ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

ALTER TABLE ONLY shipments
    ADD CONSTRAINT fk_shipments_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;

ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT;

INSERT INTO roles (code, name) VALUES
    ('customer', 'Pelanggan'),
    ('staff', 'Staff'),
    ('owner', 'Owner');
