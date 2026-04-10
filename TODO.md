## Support

Add support contact after account creation:
invoiceandgo@gmail.com

---

## Core Model

- Invoice → what is owed
- Payment Receipt → what has been paid
- Revision → changes to invoice content

## Behaviour By Invoice Status

### Draft

- Full freedom
- No revisions

### Issued

#### 1. Creates revision

- items / qty / price
- discount
- VAT
- client details
- wording changes
- deposit (requested amount)

#### 2. Does NOT create revision

- adding payment
- editing payment metadata

## Individual Features

### Payments

- Reduce balance due
- Stored as separate entries (not revisions)
- Appear in invoice history alongside revisions
- Generate Payment Receipt (printable)

### Deposits

- Represent requested upfront amount
- Do not reduce balance
- Do not create payment entries
- Deposit allowed anytime
- But after issue → editing it creates a revision

### Discounts

- Reduce total due
- If added/edited after issue → creates revision

### Supply Date

- Optional field
- Only required if different from invoice date

## Balance Logic

- balanceDue = invoice_total - sum(payments)
- Deposit does NOT affect this
- Only payments affect this

## Numbering

- Invoice + revisions
- INV-3
- INV-3.1
- INV-3.2

#### Used for:

- commercial edits only

## Payment receipts

- INV-3-PR-1
- INV-3-PR-2

#### Used for:

- payments only

---

## Leaking information for `api/clients/{id}/products` API Fix

- A user can see `null` when they try to access a client that they do not have access to.
  - This is a security concern because it allows users to infer the existence of client IDs that they do not own

- For a user accessing a client they do not own or a client that doesnt exist, we want the same outward result as for a non-existent client.

### What is wrong:

1. a client lookup path that returns:
   `{"error":{"code":"NOT_FOUND","message":"client not found"}}`
2. a product list query like:
   `WHERE account_id = ? AND client_id = ?`

That means:

- non-existent client → explicit not found
- other tenant’s client → empty result / null

### Before listing products, do a tenant-scoped client existence check:

` SELECT 1 FROM clients WHERE id = ? AND account_id = ?`

If that fails, return:

`{"error":{"code":"NOT_FOUND","message":"client not found"}}`

That way:

- client doesn’t exist → 404
- client belongs to another account → 404
- client exists in your account but has no products → []

Then and only then run the product query.
