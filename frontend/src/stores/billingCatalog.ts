import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
    formatTrialLabel,
    getBillingPrice,
    type BillingInterval,
    type BillingPlan,
    WORKSPACE_TRIAL_DAYS,
} from '@/constants/billing'
import { request } from '@/utils/fetchHelper'

export type PublicBillingCatalog = {
    configured: boolean
    trialDays: number
    singleMonthlyAvailable: boolean
    singleYearlyAvailable: boolean
    teamMonthlyAvailable: boolean
    teamYearlyAvailable: boolean
    singleMonthlyPriceLabel?: string
    singleYearlyPriceLabel?: string
    teamMonthlyPriceLabel?: string
    teamYearlyPriceLabel?: string
}

export const useBillingCatalogStore = defineStore('billingCatalog', () => {
    const catalog = ref<PublicBillingCatalog | null>(null)
    const isLoading = ref(false)
    const hasLoaded = ref(false)

    const trialDays = computed(() => Math.max(catalog.value?.trialDays ?? WORKSPACE_TRIAL_DAYS, 0))
    const trialLabel = computed(() => formatTrialLabel(trialDays.value))

    async function fetchCatalog(force = false) {
        if (hasLoaded.value && !force && catalog.value) {
            return catalog.value
        }

        isLoading.value = true
        try {
            const data = await request<PublicBillingCatalog>('/api/billing/public')
            catalog.value = data
            hasLoaded.value = true
            return data
        } finally {
            isLoading.value = false
        }
    }

    function isSelectionAvailable(plan: BillingPlan, interval: BillingInterval) {
        if (!catalog.value) return true

        switch (`${plan}:${interval}`) {
            case 'single:monthly':
                return catalog.value.singleMonthlyAvailable
            case 'single:yearly':
                return catalog.value.singleYearlyAvailable
            case 'team:monthly':
                return catalog.value.teamMonthlyAvailable
            case 'team:yearly':
                return catalog.value.teamYearlyAvailable
            default:
                return false
        }
    }

    function isPlanAvailable(plan: BillingPlan) {
        return isSelectionAvailable(plan, 'monthly') || isSelectionAvailable(plan, 'yearly')
    }

    function getPriceLabel(plan: BillingPlan, interval: BillingInterval) {
        const configuredLabel = (() => {
            if (!catalog.value) return ''

            switch (`${plan}:${interval}`) {
                case 'single:monthly':
                    return catalog.value.singleMonthlyPriceLabel ?? ''
                case 'single:yearly':
                    return catalog.value.singleYearlyPriceLabel ?? ''
                case 'team:monthly':
                    return catalog.value.teamMonthlyPriceLabel ?? ''
                case 'team:yearly':
                    return catalog.value.teamYearlyPriceLabel ?? ''
                default:
                    return ''
            }
        })()

        return configuredLabel || getBillingPrice(plan, interval).priceLabel
    }

    function listPlanPriceLabels(plan: BillingPlan) {
        const intervals: BillingInterval[] = ['monthly', 'yearly']
        const labels = intervals
            .filter((interval) => isSelectionAvailable(plan, interval))
            .map((interval) => getPriceLabel(plan, interval))

        if (labels.length > 0) {
            return labels
        }

        return intervals.map((interval) => getPriceLabel(plan, interval))
    }

    function getPlanPricingSummary(plan: BillingPlan) {
        return listPlanPriceLabels(plan).join(' or ')
    }

    function getPlanStartingPriceLabel(plan: BillingPlan) {
        return listPlanPriceLabels(plan)[0] ?? getPriceLabel(plan, 'monthly')
    }

    return {
        catalog,
        isLoading,
        hasLoaded,
        trialDays,
        trialLabel,
        fetchCatalog,
        isSelectionAvailable,
        isPlanAvailable,
        getPriceLabel,
        listPlanPriceLabels,
        getPlanPricingSummary,
        getPlanStartingPriceLabel,
    }
})
