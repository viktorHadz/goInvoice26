export type BillingPlan = 'single' | 'team'
export type BillingInterval = 'monthly' | 'yearly'

export type BillingPriceDetails = {
    priceLabel: string
    publicPriceLabel: string
}

export type BillingPlanOption = {
    id: BillingPlan
    name: string
    summary: string
    description: string
    seatLabel: string
    features: string[]
    pricing: Record<BillingInterval, BillingPriceDetails>
}

export type BillingIntervalOption = {
    id: BillingInterval
    name: string
    helper: string
}

export const DEFAULT_BILLING_PLAN: BillingPlan = 'single'
export const DEFAULT_BILLING_INTERVAL: BillingInterval = 'monthly'
export const WORKSPACE_TRIAL_DAYS = 7
export const WORKSPACE_TRIAL_LABEL = `${WORKSPACE_TRIAL_DAYS}-day free trial`
export const TEAM_PLAN_SEAT_LIMIT = 5

export const SINGLE_MONTHLY_PRICE_LABEL = '£5 / month'
export const SINGLE_MONTHLY_PRICE_LABEL_PUBLIC = '£5 / month'
export const SINGLE_YEARLY_PRICE_LABEL = '£50 / year'
export const SINGLE_YEARLY_PRICE_LABEL_PUBLIC = '£50 / year'
export const TEAM_MONTHLY_PRICE_LABEL = '£10 / month'
export const TEAM_MONTHLY_PRICE_LABEL_PUBLIC = '£10 / month'
export const TEAM_YEARLY_PRICE_LABEL = '£100 / year'
export const TEAM_YEARLY_PRICE_LABEL_PUBLIC = '£100 / year'

export const BILLING_INTERVAL_OPTIONS: BillingIntervalOption[] = [
    {
        id: 'monthly',
        name: 'Monthly',
        helper: 'Pay month to month',
    },
    {
        id: 'yearly',
        name: 'Yearly',
        helper: 'Two months free vs monthly pricing',
    },
]

export const BILLING_PLAN_OPTIONS: BillingPlanOption[] = [
    {
        id: 'single',
        name: 'Single',
        summary: 'Best for one person running their own workspace.',
        description: 'One seat for the workspace owner.',
        seatLabel: '1 person',
        features: ['One workspace owner', 'Full invoicing workspace', 'Upgrade to team later'],
        pricing: {
            monthly: {
                priceLabel: SINGLE_MONTHLY_PRICE_LABEL,
                publicPriceLabel: SINGLE_MONTHLY_PRICE_LABEL_PUBLIC,
            },
            yearly: {
                priceLabel: SINGLE_YEARLY_PRICE_LABEL,
                publicPriceLabel: SINGLE_YEARLY_PRICE_LABEL_PUBLIC,
            },
        },
    },
    {
        id: 'team',
        name: 'Team',
        summary: `Shared workspace for up to ${TEAM_PLAN_SEAT_LIMIT} people.`,
        description: 'Owner plus teammates in one account.',
        seatLabel: `Up to ${TEAM_PLAN_SEAT_LIMIT} people`,
        features: [
            'Invite teammates with Google sign-in',
            'Shared clients, invoices, and settings',
            `Seat cap of ${TEAM_PLAN_SEAT_LIMIT} people total`,
        ],
        pricing: {
            monthly: {
                priceLabel: TEAM_MONTHLY_PRICE_LABEL,
                publicPriceLabel: TEAM_MONTHLY_PRICE_LABEL_PUBLIC,
            },
            yearly: {
                priceLabel: TEAM_YEARLY_PRICE_LABEL,
                publicPriceLabel: TEAM_YEARLY_PRICE_LABEL_PUBLIC,
            },
        },
    },
]

export const WORKSPACE_PRICING_SUMMARY_LABEL = `Solo: ${SINGLE_MONTHLY_PRICE_LABEL_PUBLIC} or ${SINGLE_YEARLY_PRICE_LABEL_PUBLIC}. Teams up to ${TEAM_PLAN_SEAT_LIMIT}: ${TEAM_MONTHLY_PRICE_LABEL_PUBLIC} or ${TEAM_YEARLY_PRICE_LABEL_PUBLIC}.`

export function isBillingPlan(value: unknown): value is BillingPlan {
    return value === 'single' || value === 'team'
}

export function isBillingInterval(value: unknown): value is BillingInterval {
    return value === 'monthly' || value === 'yearly'
}

export function getBillingPlanOption(plan: BillingPlan): BillingPlanOption {
    return BILLING_PLAN_OPTIONS.find((option) => option.id === plan) ?? BILLING_PLAN_OPTIONS[0]!
}

export function getBillingPrice(plan: BillingPlan, interval: BillingInterval): BillingPriceDetails {
    return getBillingPlanOption(plan).pricing[interval]
}

export function formatBillingSelectionName(plan: BillingPlan, interval: BillingInterval): string {
    return `${getBillingPlanOption(plan).name} ${interval}`
}

export function formatTrialLabel(days: number): string {
    const normalizedDays = Math.max(Math.trunc(days || 0), 0)
    return normalizedDays > 0 ? `${normalizedDays}-day free trial` : 'No free trial'
}
