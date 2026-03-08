# Invoice Calculation Flow (Frontend + Backend)

This document describes the process for calculating invoice totals during **invoice creation** and **editing**.

The goal is:

- Fast UI updates
- Backend as the source of truth
- No silent corrections
- No calculation drift

---

# Core Principle

Always calculate from **raw invoice input**.

Never calculate from previously calculated totals.

Raw input is the only source of truth for calculations.

---

# Data Types in the Frontend

## 1. Raw Input (Editable)

This is the only state the user can change.

Examples:

- lines
- quantity
- unitPriceMinor
- minutesWorked
- vatRate
- discountType
- discountMinor
- depositType
- depositMinor
- paidMinor

These values are called **Draft Input**.

## 2. Provisional Totals

Calculated locally on the frontend.

Purpose:

- instant UI feedback
- show totals while user edits

These values are **not authoritative**.

## 3. Server Confirmed Totals

Returned from backend calculation endpoint.

Purpose:

- authoritative totals
- verification

These are the values that will be stored when the invoice is created.

---

# Calculation Flow

## Step 1 — User edits invoice

Example actions:

- add product
- change quantity
- change price
- change VAT
- apply discount
- apply deposit

Update **Draft Input**.

## Step 2 — Local calculation

Frontend recalculates totals from **Draft Input**.

Store results as:

`provisionalTotals`

UI updates immediately.

## Step 3 — Request server calculation

Send **Draft Input** to:

`POST /invoices/calculate`

Include:

- lines
- quantities
- prices
- VAT
- discount
- deposit
- paid

Do NOT send totals as input.

## Step 4 — Receive server response

Server recalculates totals from the same raw input.

Response returns:

- lineTotals
- subtotal
- discountAmount
- vatAmount
- total
- balanceDue

Store results as:

`serverConfirmedTotals`

## Step 5 — Update UI

UI shows:

- provisional totals immediately
- server-confirmed totals when response arrives

If they differ:

- replace provisional totals with server totals

---

# Handling Request Races

Each calculation request must include a **revision number**.

Example:

```
draftRevision++
```

Send revision with request.

When response arrives:

Only apply it if:

```
response.revision === currentDraftRevision
```

Otherwise discard the response.

This prevents stale responses overwriting new edits.

---

# Invoice Creation

When user presses **Create Invoice**:

1. Send **Draft Input** to backend
2. Backend recalculates totals again
3. Backend stores canonical totals
4. Backend returns created invoice

Frontend replaces local totals with returned totals.

---

# Important Rules

## Rule 1

Never use totals as input to future calculations.

Totals are always **derived**.

## Rule 2

All calculations must start from **Draft Input**.

## Rule 3

Backend calculations are authoritative.

## Rule 4

Frontend calculations are only for UX.

## Rule 5

Always recalculate on backend during create/update.

---

# Mental Model

```
Draft Input
     |
     |
     v
Local Calculator  ---> provisionalTotals
     |
     |
     v
Server Calculator ---> serverConfirmedTotals
```

Both calculators use the **same raw input**.

Neither calculator uses previously calculated totals.

---

# Summary

1. User edits Draft Input
2. FE calculates provisional totals
3. FE sends Draft Input to server
4. Server calculates authoritative totals
5. FE updates confirmed totals
6. Create endpoint recalculates again before pe
