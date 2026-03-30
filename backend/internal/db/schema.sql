-- GoInvoicer (SQLite) — clean, minimal
-- Goals:
-- - money stored as INTEGER minor units (pence)
-- - editable invoices via revisions: 123.1, 123.2, ...
-- - styles + samples supported as "products" with flat or hourly pricing
--   styles/samples, invoices, invoice items, deposits/payments
--
-- Notes:
-- - SQLite enforces FKs only when enabled:
PRAGMA foreign_keys = ON;

-- -----------------------
-- Auth / access
-- -----------------------

CREATE TABLE IF NOT EXISTS accounts (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS allowed_users (
  id INTEGER PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id),
  invited_by_user_id INTEGER,
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY,
  name TEXT,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id),
  google_sub TEXT,
  avatar_url TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT 'member',
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS auth_sessions (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  last_seen_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE IF NOT EXISTS stored_files (
  id INTEGER PRIMARY KEY,
  account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('logo')),
  storage_key TEXT NOT NULL UNIQUE,
  content_type TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  delete_pending_at TEXT
);

CREATE TABLE IF NOT EXISTS account_settings (
  account_id INTEGER PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  company_name TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL DEFAULT '',
  phone TEXT NOT NULL DEFAULT '',
  company_address TEXT NOT NULL DEFAULT '',
  invoice_prefix TEXT NOT NULL DEFAULT 'INV-',
  currency TEXT NOT NULL DEFAULT 'GBP',
  date_format TEXT NOT NULL DEFAULT 'dd/mm/yyyy',
  payment_terms TEXT NOT NULL DEFAULT 'Please make payment within 14 days.',
  payment_details TEXT NOT NULL DEFAULT '',
  notes_footer TEXT NOT NULL DEFAULT '',
  logo_asset_id INTEGER REFERENCES stored_files(id) ON DELETE SET NULL,
  show_item_type_headers INTEGER NOT NULL DEFAULT 1,
  updated_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

-- Legacy settings table retained for migration from older installs.
CREATE TABLE IF NOT EXISTS user_settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  company_name TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL DEFAULT '',
  phone TEXT NOT NULL DEFAULT '',
  company_address TEXT NOT NULL DEFAULT '',
  invoice_prefix TEXT NOT NULL DEFAULT 'INV-',
  currency TEXT NOT NULL DEFAULT 'GBP',
  date_format TEXT NOT NULL DEFAULT 'dd/mm/yyyy',
  payment_terms TEXT NOT NULL DEFAULT 'Please make payment within 14 days.',
  payment_details TEXT NOT NULL DEFAULT '',
  notes_footer TEXT NOT NULL DEFAULT '',
  logo_url TEXT NOT NULL DEFAULT '',
  show_item_type_headers INTEGER NOT NULL DEFAULT 1
);

-- -----------------------
-- Clients
-- -----------------------

CREATE TABLE IF NOT EXISTS clients (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  company_name TEXT,
  address TEXT,
  email TEXT,
  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at TEXT
);

-- -----------------------
-- Products (replaces styles + samples tables)
-- - styles: flat price only
-- - samples: flat OR hourly (hours * rate)
--
-- Previously had client-specific styles/samples via client_id.
-- (If client_id is NULL => global product visible to all clients.)
-- -----------------------

CREATE TABLE IF NOT EXISTS products (
  id INTEGER PRIMARY KEY,

  -- 'style' | 'sample'
  product_type TEXT NOT NULL CHECK (product_type IN ('style','sample')),

  -- 'flat' | 'hourly'
  -- CHECK allows only these values, and a second CHECK enforces style=flat.
  pricing_mode TEXT NOT NULL CHECK (pricing_mode IN ('flat','hourly')),

  name TEXT NOT NULL,

  -- If pricing_mode='flat' use flat_price_minor
  flat_price_minor INTEGER CHECK (flat_price_minor IS NULL OR flat_price_minor >= 0),

  -- If pricing_mode='hourly' use hourly_rate_minor
  hourly_rate_minor INTEGER CHECK (hourly_rate_minor IS NULL OR hourly_rate_minor >= 0),

  client_id INTEGER,
  is_active INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0,1)),

  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at TEXT,

  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL,

  -- Enforce: styles must be flat priced
  CHECK (NOT (product_type = 'style' AND pricing_mode <> 'flat')),

  -- Enforce: correct price fields are present depending on pricing_mode
  CHECK (
    (pricing_mode = 'flat'  AND flat_price_minor IS NOT NULL) OR
    (pricing_mode = 'hourly' AND hourly_rate_minor IS NOT NULL)
  )
);

-- -----------------------
-- Invoices
-- -----------------------

CREATE TABLE IF NOT EXISTS invoices (
  id INTEGER PRIMARY KEY,
  client_id INTEGER NOT NULL,

  -- invoice number base (the "123" in 123.1)
  base_number INTEGER NOT NULL UNIQUE CHECK (base_number > 0),

  -- 'draft' | 'issued' | 'paid' | 'void'
  status TEXT NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft','issued','paid','void')),

  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  updated_at TEXT,

  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE RESTRICT
);

-- Each edit creates a new revision row.
-- TODO: Print "base_number.revision_no" for the user.
CREATE TABLE IF NOT EXISTS invoice_revisions (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,
  revision_no INTEGER NOT NULL CHECK (revision_no >= 1),

  -- Snapshot-ish header fields (minimal)
  issue_date TEXT NOT NULL,     -- store ISO8601 date 'YYYY-MM-DD' or timestamp
  due_by_date TEXT NOT NULL,

  note TEXT,

  -- VAT: keep one global rate (20% default)
  -- Stored as basis points: 2000 = 20.00%
  vat_rate_bps INTEGER NOT NULL DEFAULT 2000 CHECK (vat_rate_bps >= 0),

  -- Discount: minimal and non-redundant
  -- type: 'none' | 'percent' | 'fixed'
  discount_type TEXT NOT NULL DEFAULT 'none'
    CHECK (discount_type IN ('none','percent','fixed')),

  -- If percent: value is bps (0..10000). If fixed: minor units.
  discount_value INTEGER NOT NULL DEFAULT 0 CHECK (discount_value >= 0),

  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
  UNIQUE (invoice_id, revision_no),

  CHECK (
    (discount_type = 'none'    AND discount_value = 0) OR
    (discount_type = 'percent' AND discount_value BETWEEN 0 AND 10000) OR
    (discount_type = 'fixed'   AND discount_value >= 0)
  )
);

-- -----------------------
-- Invoice items (lines)
-- Snapshot values live here to preserve history even if products change.
-- For hourly samples:
--   minutes_worked NOT NULL
--   unit_price_minor = hourly rate (minor)
--   total computed as minutes_worked * rate / 60
-- For flat items:
--   quantity (REAL) used; minutes_worked NULL
-- -----------------------

CREATE TABLE IF NOT EXISTS invoice_items (
  id INTEGER PRIMARY KEY,
  invoice_revision_id INTEGER NOT NULL,

  -- Optional link to product for convenience/reporting
  product_id INTEGER,

  name TEXT NOT NULL,                 -- snapshot description
  line_type TEXT NOT NULL DEFAULT 'custom'
    CHECK (line_type IN ('style','sample','custom')),

  -- For flat-priced items: quantity and unit_price_minor
  quantity REAL NOT NULL DEFAULT 1 CHECK (quantity > 0),
  unit_price_minor INTEGER NOT NULL CHECK (unit_price_minor >= 0),

  -- For hourly items (samples): store minutes worked (integer)
  minutes_worked INTEGER CHECK (minutes_worked IS NULL OR minutes_worked >= 0),

  -- Optional: keep a stable sort order
  sort_order INTEGER NOT NULL DEFAULT 1 CHECK (sort_order >= 1),

  FOREIGN KEY (invoice_revision_id) REFERENCES invoice_revisions(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL
);

-- -----------------------
-- Payments 
-- -----------------------

CREATE TABLE IF NOT EXISTS payments (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,

  -- 'deposit' | 'payment'
  kind TEXT NOT NULL DEFAULT 'payment' CHECK (kind IN ('deposit','payment')),

  amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),

  label TEXT,

  created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),

  FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
);

-- -----------------------
-- Indexes
-- -----------------------

CREATE INDEX IF NOT EXISTS idx_invoices_client_id ON invoices(client_id);
CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id ON invoice_revisions(invoice_id);
CREATE INDEX IF NOT EXISTS idx_items_revision_id ON invoice_items(invoice_revision_id);
CREATE INDEX IF NOT EXISTS idx_items_product_id ON invoice_items(product_id);
CREATE INDEX IF NOT EXISTS idx_products_client_id ON products(client_id);
CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);
