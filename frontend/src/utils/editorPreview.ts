import type { InvoiceLine } from '@/components/invoice/invoiceTypes'
import { formatInvoiceDurationMinutes } from '@/utils/duration'
import { lineTotalMinor } from '@/utils/money'

export function formatEditorPreviewLineMeta(line: InvoiceLine): string {
    if (line.pricingMode !== 'hourly' || line.minutesWorked == null) {
        return line.pricingMode
    }

    return `${line.pricingMode} · ${formatInvoiceDurationMinutes(line.minutesWorked)}`
}

export function editorPreviewLineTotalMinor(line: InvoiceLine): number {
    return lineTotalMinor(line)
}
