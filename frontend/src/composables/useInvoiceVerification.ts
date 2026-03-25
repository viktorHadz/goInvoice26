import { ref, type Ref } from 'vue'
import type { Invoice, Totals, MoneyMinor } from '@/components/invoice/invoiceTypes'
import { verifyInvoiceHandler } from '@/utils/invoiceHttpHandler'
import { apiDTO } from '@/utils/invoiceDto'
import { lineTotalMinor } from '@/utils/money'
import { isApiError, hasFieldErrors, toFieldErrorMap } from '@/utils/apiErrors'
import { NetworkError } from '@/utils/fetchHelper'
import { emitToastError } from '@/utils/toast'

type VerifyStatus = 'idle' | 'checking' | 'ok' | 'mismatch' | 'invalid' | 'error'

export function useInvoiceVerification(
    invoice: Ref<Invoice | null>,
    clientId: Ref<number | null | undefined>,
    totals: Ref<Totals | null>,
    serverFieldErrors: Ref<Record<string, string>>,
) {
    const verifyStatus = ref<VerifyStatus>('idle')
    const lastVerifyAt = ref<number | null>(null)
    const serverCanonicalTotals = ref<Totals | null>(null)
    const serverCanonicalLineTotals = ref<Record<number, MoneyMinor>>({})
    const lastVerifyFailureToastedAt = ref<number | null>(null)

    let verifyTimer: number | null = null
    let verifyAbort: AbortController | null = null

    function clearVerifyTimer() {
        if (verifyTimer != null) {
            window.clearTimeout(verifyTimer)
            verifyTimer = null
        }
    }

    function abortVerify() {
        if (verifyAbort) {
            verifyAbort.abort()
            verifyAbort = null
        }
    }

    function clearVerifyState() {
        verifyStatus.value = 'idle'
        lastVerifyAt.value = null
        serverCanonicalTotals.value = null
        serverCanonicalLineTotals.value = {}
        clearVerifyTimer()
        abortVerify()
    }

    async function runServerVerify() {
        const inv = invoice.value
        const cid = clientId.value
        if (!inv || !cid) return

        const dto = apiDTO(inv)
        abortVerify()
        verifyAbort = new AbortController()

        verifyStatus.value = 'checking'

        try {
            const res = await verifyInvoiceHandler(dto.overview.clientId, inv.baseNumber, dto, {
                signal: verifyAbort.signal,
            })

            const canonical = (res as any)?.invoice as any
            const canonicalLines: any[] = Array.isArray(canonical?.lines) ? canonical.lines : []
            const canonicalTotals = canonical?.totals

            const canonicalBySort: Record<number, MoneyMinor> = {}
            for (const ln of canonicalLines) {
                const so = Number(ln?.sortOrder)
                const lt = Number(ln?.lineTotalMinor)
                if (Number.isFinite(so) && Number.isFinite(lt)) {
                    canonicalBySort[so] = Math.round(lt) as MoneyMinor
                }
            }

            const parsedTotals: Totals | null =
                canonicalTotals &&
                typeof canonicalTotals === 'object' &&
                Number.isFinite((canonicalTotals as any).subtotalMinor) &&
                Number.isFinite((canonicalTotals as any).discountMinor) &&
                Number.isFinite((canonicalTotals as any).subtotalAfterDiscountMinor) &&
                Number.isFinite((canonicalTotals as any).vatMinor) &&
                Number.isFinite((canonicalTotals as any).totalMinor)
                    ? {
                          subtotalMinor: Math.round(
                              (canonicalTotals as any).subtotalMinor,
                          ) as MoneyMinor,
                          discountMinor: Math.round(
                              (canonicalTotals as any).discountMinor,
                          ) as MoneyMinor,
                          subtotalAfterDiscountMinor: Math.round(
                              (canonicalTotals as any).subtotalAfterDiscountMinor,
                          ) as MoneyMinor,
                          vatMinor: Math.round((canonicalTotals as any).vatMinor) as MoneyMinor,
                          totalMinor: Math.round((canonicalTotals as any).totalMinor) as MoneyMinor,
                      }
                    : null

            serverCanonicalLineTotals.value = canonicalBySort
            serverCanonicalTotals.value = parsedTotals
            lastVerifyAt.value = Date.now()
            lastVerifyFailureToastedAt.value = null

            const optimisticTotals = totals.value
            let mismatch = false

            if (optimisticTotals && parsedTotals) {
                mismatch =
                    optimisticTotals.subtotalMinor !== parsedTotals.subtotalMinor ||
                    optimisticTotals.discountMinor !== parsedTotals.discountMinor ||
                    optimisticTotals.subtotalAfterDiscountMinor !==
                        parsedTotals.subtotalAfterDiscountMinor ||
                    optimisticTotals.vatMinor !== parsedTotals.vatMinor ||
                    optimisticTotals.totalMinor !== parsedTotals.totalMinor
            }

            for (const line of inv.lines) {
                const serverLT = canonicalBySort[line.sortOrder]
                if (serverLT == null) continue
                const optimisticLT = lineTotalMinor(line)
                if (optimisticLT !== serverLT) {
                    mismatch = true
                    break
                }
            }

            verifyStatus.value = mismatch ? 'mismatch' : 'ok'
        } catch (err: unknown) {
            if (err instanceof NetworkError) {
                verifyStatus.value = 'error'
                if (lastVerifyFailureToastedAt.value == null) {
                    lastVerifyFailureToastedAt.value = Date.now()
                    emitToastError({
                        title: 'Verification unavailable',
                        message: 'Could not verify totals right now. Check your connection and try again.',
                    })
                }
                return
            }

            if (isApiError(err) && err.code === 'VALIDATION_FAILED') {
                verifyStatus.value = 'invalid'
                if (hasFieldErrors(err)) {
                    serverFieldErrors.value = toFieldErrorMap(err.fields)
                }
                return
            }

            verifyStatus.value = 'error'
            if (lastVerifyFailureToastedAt.value == null) {
                lastVerifyFailureToastedAt.value = Date.now()
                emitToastError({
                    title: 'Verification failed',
                    message:
                        isApiError(err) && err.message.trim().length > 0
                            ? err.message
                            : 'Could not verify totals right now. Please try again.',
                })
            }
            console.error('[editor verify]', err)
        }
    }

    function scheduleServerVerify(debounceDur = 1000) {
        if (typeof window === 'undefined') return

        const inv = invoice.value
        if (!inv) return
        if (inv.lines.length <= 0) return
        if (!inv.issueDate) return

        clearVerifyTimer()
        verifyTimer = window.setTimeout(() => {
            runServerVerify()
        }, debounceDur)
    }

    return {
        verifyStatus,
        lastVerifyAt,
        serverCanonicalTotals,
        serverCanonicalLineTotals,
        runServerVerify,
        scheduleServerVerify,
        clearVerifyState,
        abortVerify,
    }
}
