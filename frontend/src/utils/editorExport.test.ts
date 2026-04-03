import { describe, expect, it } from 'vitest'
import { resolveEditorExportRevisionNo } from '@/utils/editorExport'

describe('resolveEditorExportRevisionNo', () => {
    it('uses the explicit revision when a revision node is selected', () => {
        expect(
            resolveEditorExportRevisionNo(
                {
                    type: 'revision',
                    clientId: 1,
                    id: 20,
                    invoiceId: 10,
                    baseNo: 100,
                    revisionNo: 4,
                },
                2,
            ),
        ).toBe(4)
    })

    it('uses the loaded active revision for invoice-level selections', () => {
        expect(
            resolveEditorExportRevisionNo(
                {
                    type: 'invoice',
                    clientId: 1,
                    id: 10,
                    baseNo: 100,
                },
                3,
            ),
        ).toBe(3)
    })

    it('falls back to revision one when no active revision is known yet', () => {
        expect(resolveEditorExportRevisionNo(null, 0)).toBe(1)
    })
})
