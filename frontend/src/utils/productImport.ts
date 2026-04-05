import type { APIFieldError } from '@/utils/apiErrors'

export type ProductImportKind = 'style' | 'sample_flat' | 'sample_hourly'

export type ProductImportResult = {
    createdCount: number
    clientId: number
    importKind: ProductImportKind
}

export type ProductImportKindOption = {
    id: ProductImportKind
    name: string
    productType: 'style' | 'sample'
    pricingMode: 'flat' | 'hourly'
    summary: string
    csvHeader: string
}

export type FormattedProductImportError = {
    id: string
    message: string
    row: number | null
    column: string | null
}

export const PRODUCT_IMPORT_MAX_BYTES = 50 * 1024
export const PRODUCT_IMPORT_MAX_ROWS = 400
export const PRODUCT_IMPORT_ACCEPT = '.csv,text/csv'

export const PRODUCT_IMPORT_KIND_OPTIONS: ProductImportKindOption[] = [
    {
        id: 'style',
        name: 'Styles',
        productType: 'style',
        pricingMode: 'flat',
        summary: 'Create flat-priced styles for the selected client.',
        csvHeader: 'name,unit price',
    },
    {
        id: 'sample_flat',
        name: 'Samples (Flat)',
        productType: 'sample',
        pricingMode: 'flat',
        summary: 'Create flat-priced samples for the selected client.',
        csvHeader: 'name,unit price',
    },
    {
        id: 'sample_hourly',
        name: 'Samples (Hourly)',
        productType: 'sample',
        pricingMode: 'hourly',
        summary: 'Create hourly samples with default minutes for the selected client.',
        csvHeader: 'name,time to produce (in minutes),unit price',
    },
]

export function getProductImportOption(kind: ProductImportKind): ProductImportKindOption {
    return (
        PRODUCT_IMPORT_KIND_OPTIONS.find((option) => option.id === kind) ??
        PRODUCT_IMPORT_KIND_OPTIONS[0]!
    )
}

export function validateProductImportFile(file: unknown): File {
    if (!(file instanceof File)) {
        throw new Error('No CSV file selected.')
    }

    if (file.size <= 0) {
        throw new Error('Selected CSV is empty.')
    }

    if (file.size > PRODUCT_IMPORT_MAX_BYTES) {
        throw new Error('CSV is too large. Maximum size is 50KB.')
    }

    const name = file.name.trim().toLowerCase()
    const type = file.type.trim().toLowerCase()
    const nameLooksCSV = name.endsWith('.csv')
    const typeLooksCSV =
        type === '' ||
        type === 'text/csv' ||
        type === 'application/csv' ||
        type === 'text/plain' ||
        type === 'application/vnd.ms-excel'

    if (!nameLooksCSV && !typeLooksCSV) {
        throw new Error('Upload a CSV file.')
    }

    return file
}

export function formatProductImportErrors(fields: APIFieldError[]): FormattedProductImportError[] {
    return fields.map((fieldError, index) => {
        const rowMeta = fieldError.meta?.row
        const columnMeta = fieldError.meta?.column

        return {
            id: `${fieldError.field || 'import'}:${fieldError.code}:${index}`,
            message:
                typeof fieldError.message === 'string' && fieldError.message.trim().length > 0
                    ? fieldError.message
                    : 'Invalid CSV value.',
            row: typeof rowMeta === 'number' && Number.isFinite(rowMeta) ? rowMeta : null,
            column:
                typeof columnMeta === 'string' && columnMeta.trim().length > 0 ? columnMeta : null,
        }
    })
}
