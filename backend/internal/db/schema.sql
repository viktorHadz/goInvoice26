PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS accounts (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL DEFAULT '',
  stripe_customer_id TEXT NOT NULL DEFAULT '',
  stripe_subscription_id TEXT NOT NULL DEFAULT '',
  billing_price_id TEXT NOT NULL DEFAULT '',
  billing_plan TEXT NOT NULL DEFAULT '',
  billing_interval TEXT NOT NULL DEFAULT '',
  billing_email TEXT NOT NULL DEFAULT '',
  billing_status TEXT NOT NULL DEFAULT 'inactive',
  billing_current_period_end TEXT NOT NULL DEFAULT '',
  billing_cancel_at_period_end INTEGER NOT NULL DEFAULT 0,
  billing_updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

INSERT OR IGNORE INTO accounts (id, name) VALUES (1, 'Default account');

CREATE TABLE IF NOT EXISTS allowed_users (
  id INTEGER PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id),
  invited_by_user_id INTEGER,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS direct_access_grants (
  id INTEGER PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  plan TEXT NOT NULL DEFAULT 'single',
  note TEXT NOT NULL DEFAULT '',
  created_by_user_id INTEGER,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS promo_codes (
  id INTEGER PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  duration_days INTEGER NOT NULL CHECK (duration_days > 0),
  active INTEGER NOT NULL DEFAULT 1,
  created_by_user_id INTEGER,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS promo_code_redemptions (
  id INTEGER PRIMARY KEY,
  promo_code_id INTEGER NOT NULL REFERENCES promo_codes(id) ON DELETE CASCADE,
  account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  redeemed_by_user_id INTEGER,
  redeemed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  expires_at TEXT NOT NULL,
  UNIQUE (promo_code_id, account_id)
);

CREATE TABLE IF NOT EXISTS promo_code_redemption_claims (
  id INTEGER PRIMARY KEY,
  promo_code_id INTEGER NOT NULL REFERENCES promo_codes(id) ON DELETE CASCADE,
  owner_email_hmac TEXT NOT NULL,
  redeemed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  retention_until TEXT NOT NULL,
  UNIQUE (promo_code_id, owner_email_hmac)
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
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS auth_sessions (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  last_seen_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS stored_files (
  id INTEGER PRIMARY KEY,
  account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('logo')),
  storage_key TEXT NOT NULL UNIQUE,
  content_type TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
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
  legacy_logo_url TEXT NOT NULL DEFAULT '',
  show_item_type_headers INTEGER NOT NULL DEFAULT 1,
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE IF NOT EXISTS clients (
  id INTEGER PRIMARY KEY,
  account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  company_name TEXT,
  address TEXT,
  email TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT,
  UNIQUE (account_id, id)
);

CREATE TABLE IF NOT EXISTS products (
  id INTEGER PRIMARY KEY,
  account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id) ON DELETE CASCADE,
  product_type TEXT NOT NULL CHECK (product_type IN ('style','sample')),
  pricing_mode TEXT NOT NULL CHECK (pricing_mode IN ('flat','hourly')),
  name TEXT NOT NULL,
  flat_price_minor INTEGER CHECK (flat_price_minor IS NULL OR flat_price_minor >= 0),
  hourly_rate_minor INTEGER CHECK (hourly_rate_minor IS NULL OR hourly_rate_minor >= 0),
  default_minutes_worked INTEGER CHECK (default_minutes_worked IS NULL OR default_minutes_worked >= 0),
  client_id INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT,
  FOREIGN KEY (account_id, client_id) REFERENCES clients(account_id, id) ON DELETE CASCADE,
  UNIQUE (account_id, client_id, id),
  CHECK (NOT (product_type = 'style' AND pricing_mode <> 'flat')),
  CHECK (
    (pricing_mode = 'flat' AND flat_price_minor IS NOT NULL) OR
    (pricing_mode = 'hourly' AND hourly_rate_minor IS NOT NULL)
  ),
  CHECK (
    (product_type = 'sample' AND pricing_mode = 'hourly' AND default_minutes_worked IS NOT NULL) OR
    (NOT (product_type = 'sample' AND pricing_mode = 'hourly') AND default_minutes_worked IS NULL)
  )
);

CREATE TABLE IF NOT EXISTS invoices (
  id INTEGER PRIMARY KEY,
  account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id) ON DELETE CASCADE,
  client_id INTEGER NOT NULL,
  current_revision_id INTEGER
    REFERENCES invoice_revisions(id)
    ON DELETE SET NULL
    DEFERRABLE INITIALLY DEFERRED,
  base_number INTEGER NOT NULL CHECK (base_number > 0),
  status TEXT NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft','issued','paid','void')),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (account_id, client_id) REFERENCES clients(account_id, id) ON DELETE RESTRICT,
  UNIQUE (account_id, base_number),
  UNIQUE (id, account_id, client_id)
);

CREATE TABLE IF NOT EXISTS invoice_number_seq (
  account_id INTEGER PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  next_base_number INTEGER NOT NULL CHECK (next_base_number > 0)
);

INSERT OR IGNORE INTO invoice_number_seq (account_id, next_base_number) VALUES (1, 1);

CREATE TABLE IF NOT EXISTS invoice_revisions (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,
  revision_no INTEGER NOT NULL CHECK (revision_no >= 1),
  issue_date TEXT NOT NULL,
  supply_date TEXT,
  due_by_date TEXT,
  updated_at TEXT,
  client_name TEXT NOT NULL,
  client_company_name TEXT NOT NULL DEFAULT '',
  client_address TEXT NOT NULL DEFAULT '',
  client_email TEXT NOT NULL DEFAULT '',
  note TEXT,
  vat_rate INTEGER NOT NULL DEFAULT 2000 CHECK (vat_rate BETWEEN 0 AND 10000),
  discount_type TEXT NOT NULL DEFAULT 'none'
    CHECK (discount_type IN ('none','percent','fixed')),
  discount_rate INTEGER NOT NULL DEFAULT 0 CHECK (discount_rate BETWEEN 0 AND 10000),
  discount_minor INTEGER NOT NULL DEFAULT 0 CHECK (discount_minor >= 0),
  deposit_type TEXT NOT NULL DEFAULT 'none'
    CHECK (deposit_type IN ('none','percent','fixed')),
  deposit_rate INTEGER NOT NULL DEFAULT 0 CHECK (deposit_rate BETWEEN 0 AND 10000),
  deposit_minor INTEGER NOT NULL DEFAULT 0 CHECK (deposit_minor >= 0),
  subtotal_minor INTEGER NOT NULL CHECK (subtotal_minor >= 0),
  vat_amount_minor INTEGER NOT NULL CHECK (vat_amount_minor >= 0),
  total_minor INTEGER NOT NULL CHECK (total_minor >= 0),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
  UNIQUE (id, invoice_id),
  UNIQUE (invoice_id, revision_no),
  CHECK (
    (discount_type = 'none' AND discount_rate = 0 AND discount_minor = 0) OR
    (discount_type = 'percent' AND discount_rate BETWEEN 0 AND 10000) OR
    (discount_type = 'fixed' AND discount_rate = 0)
  ),
  CHECK (
    (deposit_type = 'none' AND deposit_rate = 0 AND deposit_minor = 0) OR
    (deposit_type = 'percent' AND deposit_rate BETWEEN 0 AND 10000) OR
    (deposit_type = 'fixed' AND deposit_rate = 0)
  )
);

CREATE TABLE IF NOT EXISTS invoice_items (
  id INTEGER PRIMARY KEY,
  invoice_revision_id INTEGER NOT NULL,
  product_id INTEGER,
  name TEXT NOT NULL,
  line_type TEXT NOT NULL DEFAULT 'custom'
    CHECK (line_type IN ('style','sample','custom')),
  pricing_mode TEXT NOT NULL DEFAULT 'flat'
    CHECK (pricing_mode IN ('flat','hourly')),
  quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
  unit_price_minor INTEGER NOT NULL CHECK (unit_price_minor >= 0),
  line_total_minor INTEGER NOT NULL DEFAULT 0 CHECK (line_total_minor >= 0),
  minutes_worked INTEGER CHECK (minutes_worked IS NULL OR minutes_worked >= 0),
  sort_order INTEGER NOT NULL DEFAULT 1 CHECK (sort_order >= 1),
  FOREIGN KEY (invoice_revision_id) REFERENCES invoice_revisions(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL,
  UNIQUE (invoice_revision_id, sort_order),
  CHECK (
    (pricing_mode = 'flat' AND minutes_worked IS NULL) OR
    (pricing_mode = 'hourly' AND minutes_worked IS NOT NULL)
  )
);

CREATE TABLE IF NOT EXISTS payments (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,
  receipt_no INTEGER NOT NULL DEFAULT 0 CHECK (receipt_no >= 0),
  payment_type TEXT NOT NULL DEFAULT 'payment'
    CHECK (payment_type IN ('deposit','payment')),
  amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),
  payment_date TEXT NOT NULL,
  applied_in_revision_id INTEGER
    REFERENCES invoice_revisions(id)
    ON DELETE SET NULL
    DEFERRABLE INITIALLY DEFERRED,
  label TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trg_accounts_id_immutable
BEFORE UPDATE OF id ON accounts
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'account identity is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_users_scope_immutable
BEFORE UPDATE OF id, account_id ON users
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'user ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_allowed_users_scope_immutable
BEFORE UPDATE OF id, account_id ON allowed_users
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'invite ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_direct_access_grants_scope_immutable
BEFORE UPDATE OF id ON direct_access_grants
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'direct access grant identity is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_promo_codes_scope_immutable
BEFORE UPDATE OF id ON promo_codes
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'promo code identity is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_promo_code_redemptions_scope_immutable
BEFORE UPDATE OF id, promo_code_id, account_id ON promo_code_redemptions
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'promo code redemption ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_promo_code_redemption_claims_scope_immutable
BEFORE UPDATE OF id, promo_code_id, owner_email_hmac ON promo_code_redemption_claims
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'promo code redemption claim ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_auth_sessions_scope_immutable
BEFORE UPDATE OF id, account_id, user_id ON auth_sessions
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'session ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_stored_files_scope_immutable
BEFORE UPDATE OF id, account_id ON stored_files
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'stored file ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_account_settings_scope_immutable
BEFORE UPDATE OF account_id ON account_settings
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'settings ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoice_number_seq_scope_immutable
BEFORE UPDATE OF account_id ON invoice_number_seq
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'invoice sequence ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_clients_account_immutable
BEFORE UPDATE OF id, account_id ON clients
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'client ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_products_scope_immutable
BEFORE UPDATE OF id, account_id, client_id ON products
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'product ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoices_scope_immutable
BEFORE UPDATE OF id, account_id, client_id ON invoices
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'invoice ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoice_revisions_scope_immutable
BEFORE UPDATE OF id, invoice_id ON invoice_revisions
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'invoice revision ownership is immutable');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoices_current_revision_matches_invoice_insert
AFTER INSERT ON invoices
FOR EACH ROW
WHEN NEW.current_revision_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invoice_revisions r
    WHERE r.id = NEW.current_revision_id
      AND r.invoice_id = NEW.id
  )
BEGIN
  SELECT RAISE(ABORT, 'invoice current revision must belong to same invoice');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoices_current_revision_matches_invoice_update
AFTER UPDATE OF current_revision_id ON invoices
FOR EACH ROW
WHEN NEW.current_revision_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invoice_revisions r
    WHERE r.id = NEW.current_revision_id
      AND r.invoice_id = NEW.id
  )
BEGIN
  SELECT RAISE(ABORT, 'invoice current revision must belong to same invoice');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoice_items_product_scope_insert
BEFORE INSERT ON invoice_items
FOR EACH ROW
WHEN NEW.product_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invoice_revisions r
    JOIN invoices i
      ON i.id = r.invoice_id
    JOIN products p
      ON p.id = NEW.product_id
    WHERE r.id = NEW.invoice_revision_id
      AND p.account_id = i.account_id
      AND p.client_id = i.client_id
  )
BEGIN
  SELECT RAISE(ABORT, 'invoice item product must belong to same account and client');
END;

CREATE TRIGGER IF NOT EXISTS trg_invoice_items_product_scope_update
BEFORE UPDATE OF invoice_revision_id, product_id ON invoice_items
FOR EACH ROW
WHEN NEW.product_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invoice_revisions r
    JOIN invoices i
      ON i.id = r.invoice_id
    JOIN products p
      ON p.id = NEW.product_id
    WHERE r.id = NEW.invoice_revision_id
      AND p.account_id = i.account_id
      AND p.client_id = i.client_id
  )
BEGIN
  SELECT RAISE(ABORT, 'invoice item product must belong to same account and client');
END;

CREATE TRIGGER IF NOT EXISTS trg_payments_applied_revision_matches_invoice_insert
BEFORE INSERT ON payments
FOR EACH ROW
WHEN NEW.applied_in_revision_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invoice_revisions r
    WHERE r.id = NEW.applied_in_revision_id
      AND r.invoice_id = NEW.invoice_id
  )
BEGIN
  SELECT RAISE(ABORT, 'payment applied revision must belong to same invoice');
END;

CREATE TRIGGER IF NOT EXISTS trg_payments_applied_revision_matches_invoice_update
BEFORE UPDATE OF invoice_id, applied_in_revision_id ON payments
FOR EACH ROW
WHEN NEW.applied_in_revision_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM invoice_revisions r
    WHERE r.id = NEW.applied_in_revision_id
      AND r.invoice_id = NEW.invoice_id
  )
BEGIN
  SELECT RAISE(ABORT, 'payment applied revision must belong to same invoice');
END;

CREATE INDEX IF NOT EXISTS idx_invoices_current_revision_id ON invoices(current_revision_id);
CREATE INDEX IF NOT EXISTS idx_promo_code_redemption_claims_retention_until ON promo_code_redemption_claims(retention_until);

CREATE VIEW IF NOT EXISTS invoice_current_items AS
SELECT
  i.id AS invoice_id,
  i.current_revision_id AS revision_id,
  r.revision_no,
  it.id AS item_id,
  it.sort_order,
  it.product_id,
  it.name,
  it.line_type,
  it.pricing_mode,
  it.quantity,
  it.unit_price_minor,
  it.minutes_worked,
  it.line_total_minor
FROM invoices i
JOIN invoice_revisions r
  ON r.id = i.current_revision_id
JOIN invoice_items it
  ON it.invoice_revision_id = r.id
ORDER BY i.id, it.sort_order;

CREATE VIEW IF NOT EXISTS invoice_book_rows AS
SELECT
  i.id,
  i.client_id,
  i.base_number,
  i.status,
  i.current_revision_id,
  r.revision_no,
  r.issue_date,
  r.due_by_date,
  r.updated_at
FROM invoices i
JOIN invoice_revisions r
  ON r.id = i.current_revision_id;

CREATE VIEW IF NOT EXISTS invoice_revision_items AS
SELECT
  r.invoice_id,
  r.id AS revision_id,
  r.revision_no,
  it.id AS item_id,
  it.sort_order,
  it.product_id,
  it.name,
  it.line_type,
  it.pricing_mode,
  it.quantity,
  it.unit_price_minor,
  it.minutes_worked,
  it.line_total_minor
FROM invoice_revisions r
JOIN invoice_items it
  ON it.invoice_revision_id = r.id;

CREATE INDEX IF NOT EXISTS idx_invoices_client_id ON invoices(client_id);
CREATE INDEX IF NOT EXISTS idx_clients_account_id ON clients(account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_account_id_id ON clients(account_id, id);
CREATE INDEX IF NOT EXISTS idx_invoices_account_id ON invoices(account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_id_account_client ON invoices(id, account_id, client_id);
CREATE INDEX IF NOT EXISTS idx_invoices_client_base ON invoices(account_id, client_id, base_number DESC);
CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id ON invoice_revisions(invoice_id);
CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id_revno ON invoice_revisions(invoice_id, revision_no);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoice_revisions_id_invoice ON invoice_revisions(id, invoice_id);
CREATE INDEX IF NOT EXISTS idx_items_revision_id ON invoice_items(invoice_revision_id);
CREATE INDEX IF NOT EXISTS idx_items_product_id ON invoice_items(product_id);
CREATE INDEX IF NOT EXISTS idx_products_account_client ON products(account_id, client_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_account_client_id ON products(account_id, client_id, id);
CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payments_invoice_revision ON payments(invoice_id, applied_in_revision_id);
-- Keep indexes for newly introduced columns in targeted migrations so legacy DBs can
-- add the column before bootstrap tries to reference it.
CREATE INDEX IF NOT EXISTS idx_stored_files_account_id ON stored_files(account_id);
CREATE INDEX IF NOT EXISTS idx_stored_files_delete_pending ON stored_files(delete_pending_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_sub ON users(google_sub) WHERE google_sub IS NOT NULL AND google_sub <> '';
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_users_account_id ON users(account_id);
CREATE INDEX IF NOT EXISTS idx_allowed_users_account_id ON allowed_users(account_id);
