
### Support
Add support contact after account creation:
invoiceandgo@gmail.com
-----
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
-----

Implementation details:
 1) Overall 
- Draft behaviour stays the same. You can edit as much as you want and it will not create revisions. Once the invoice becomes issued changes to payments - create payment receipt.
 - Changes to items, qty, pricing, adding new items and etc create a revision.
 - Payments no longer create revisions. They create a separate entry in the invoice history along revisions called - payment receipt. 
 - Revisions only get created when there are edits on the quantity, price or a discount gets input after an invoice has been issued.
 - Discounts reduce due amount and create a revision.

2) Payments: 
- reduce balance due
- are separate entries not revisions
- sit in invoice history alongside revisions
- create a printable entry called payment receipt with different template than invoice. Can be printed like invoice in pdf or docx

3) Deposits:
- do not reduce balance due
- do not create payment receipt entry
- allowed only after 

4) Discount
- changes the balance due
- if added after issue, that is an invoice edit, so that creates a revision, not a payment entry

5) Supply date (optional) HMRC requires the invoice date and the supply date. The supply date is the date when the goods or services were supplied. 

For commercial edits (same without payments) - [InvPrefix]-3-Rev-1 [InvPrefix]-3-Rev-2 [InvPrefix]-3-Rev-[InvPrefix]-3-Rev-3 .... 
qty
lines
price
discount
VAT
client details
wording

Payments / receipts - 3-PR-1, 3-PR-2, 3-PR-3 ....




