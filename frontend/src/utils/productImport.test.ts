import { describe, expect, it } from 'vitest'
import {
    PRODUCT_IMPORT_MAX_BYTES,
    formatProductImportErrors,
    getProductImportOption,
    validateProductImportFile,
} from '@/utils/productImport'

describe('product import helpers', () => {
    it('returns the correct instructions for each import kind', () => {
        expect(getProductImportOption('style').csvHeader).toBe('name,unit price')
        expect(getProductImportOption('sample_flat').summary).toContain('flat-priced samples')
        expect(getProductImportOption('sample_hourly').csvHeader).toBe(
            'name,time to produce (in minutes),unit price',
        )
    })

    it('accepts csv files and rejects invalid uploads early', () => {
        const valid = new File(['name,unit price\nHemline,12.50\n'], 'styles.csv', {
            type: 'text/csv',
        })

        expect(validateProductImportFile(valid)).toBe(valid)

        const oversize = new File([new Uint8Array(PRODUCT_IMPORT_MAX_BYTES + 1)], 'big.csv', {
            type: 'text/csv',
        })
        expect(() => validateProductImportFile(oversize)).toThrow('Maximum size is 50KB.')

        const wrongType = new File(['not csv'], 'image.png', {
            type: 'image/png',
        })
        expect(() => validateProductImportFile(wrongType)).toThrow('Upload a CSV file.')
    })

    it('formats row-level api errors into a readable list model', () => {
        const formatted = formatProductImportErrors([
            {
                field: 'rows[2].unitPrice',
                code: 'REQUIRED',
                message: 'Row 2: unit price is required',
                meta: {
                    row: 2,
                    column: 'unit price',
                },
            },
            {
                field: 'header',
                code: 'INVALID',
                message: 'columns must exactly be: name, unit price',
            },
        ])

        expect(formatted).toEqual([
            {
                id: 'rows[2].unitPrice:REQUIRED:0',
                message: 'Row 2: unit price is required',
                row: 2,
                column: 'unit price',
            },
            {
                id: 'header:INVALID:1',
                message: 'columns must exactly be: name, unit price',
                row: null,
                column: null,
            },
        ])
    })
})
