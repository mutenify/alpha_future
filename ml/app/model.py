"""Canonical risk model v1, mirrored from migrations 000002 and 000004."""

from __future__ import annotations

import math
from dataclasses import dataclass
from statistics import NormalDist


BASE_SCORE = 600
BASE_ODDS = 50.0
PDO = 20
MODEL_VERSION = "v1"

SCORE_BINS = {
    "age_months": ((0, 12, 46), (12, 24, 51), (24, 60, 70), (60, None, 66)),
    "overdues12m": ((-1, 0, 74), (0, 1, 62), (1, 3, 40), (3, None, 50)),
    "max_overdue_days": ((-1, 0, 63), (0, 30, 62), (30, None, 62)),
    "inquiries3m": ((-1, 0, 64), (0, 2, 63), (2, None, 58)),
    "tax_debt": ((-1, 0, 63), (0, None, 62)),
    "revenue_trend": ((None, -0.1, 61), (-0.1, 0.1, 74), (0.1, None, 47)),
    "revenue_cv": ((0, 0.25, 63), (0.25, 0.4, 64), (0.4, None, 60)),
    "debt_load": ((-1, 0, 72), (0, 0.3, 62), (0.3, 0.6, 58), (0.6, None, 44)),
}
OKVED_POINTS = {"10": 71, "41": 38, "47": 77, "49": 71, "56": 56, "62": 62}
INDUSTRIES = {
    "10": (0.09, 0.30), "41": (0.10, 0.45), "47": (0.08, 0.25),
    "49": (0.12, 0.30), "56": (0.10, 0.35), "62": (0.20, 0.30),
    "OTHER": (0.10, 0.35),
}


@dataclass(frozen=True)
class Product:
    code: str
    min_amount: float
    max_amount: float
    min_term: int
    max_term: int
    base_rate: float
    lgd: float
    pd_auto: float
    pd_reject: float
    gap_q: float
    payment_k: float


PRODUCTS = {
    "WORKING_CAPITAL": Product("WORKING_CAPITAL", 300_000, 30_000_000, 3, 36, 19.50, 0.60, 0.10, 0.25, 0.05, 0.50),
    "INVESTMENT_LOAN": Product("INVESTMENT_LOAN", 1_000_000, 100_000_000, 12, 84, 18.00, 0.40, 0.10, 0.25, 0.03, 0.40),
    "OVERDRAFT": Product("OVERDRAFT", 100_000, 10_000_000, 1, 12, 21.00, 0.75, 0.10, 0.25, 0.08, 0.30),
}


def _points(name: str, value: float) -> int:
    for lower, upper, points in SCORE_BINS[name]:
        if (lower is None or value > lower) and (upper is None or value <= upper):
            return points
    return 0


def score(values: dict[str, float | str]) -> int:
    total = sum(_points(name, float(values[name])) for name in SCORE_BINS)
    return total + OKVED_POINTS.get(str(values["industry_code"]), 0)


def pd_from_score(points: int) -> float:
    return 1.0 / (1.0 + BASE_ODDS * 2.0 ** ((points - BASE_SCORE) / PDO))


def round2(value: float) -> float:
    return math.floor(value * 100.0 + 0.5) / 100.0


def equilibrium_rate(base_rate_pct: float, pd: float, lgd: float) -> float:
    return round2(((base_rate_pct / 100.0 + pd * lgd) / (1.0 - pd)) * 100.0)


def payment(principal: float, term_months: int, rate_pct: float) -> float:
    monthly_rate = rate_pct / 100.0 / 12.0
    if monthly_rate == 0:
        return principal / term_months
    return principal * monthly_rate / (1.0 - (1.0 + monthly_rate) ** -term_months)


def principal(payment_value: float, term_months: int, rate_pct: float) -> float:
    monthly_rate = rate_pct / 100.0 / 12.0
    if monthly_rate == 0:
        return payment_value * term_months
    return payment_value * (1.0 - (1.0 + monthly_rate) ** -term_months) / monthly_rate


def p_gap(mu: float, sigma: float, debt_old: float, loan_payment: float, cash_balance: float, term_months: int) -> float:
    cushion = cash_balance / term_months if term_months > 0 else 0.0
    threshold = debt_old + loan_payment - cushion
    if sigma == 0:
        return float(mu < threshold)
    return NormalDist().cdf((threshold - mu) / sigma)


def max_payment(mu: float, sigma: float, debt_old: float, cash_balance: float, term: int, product: Product) -> float:
    cushion = cash_balance / term if term > 0 else 0.0
    by_gap = mu + cushion - debt_old + sigma * NormalDist().inv_cdf(product.gap_q)
    by_share = product.payment_k * mu - debt_old
    return max(0.0, min(by_gap, by_share))


def terms(mu: float, sigma: float, debt_old: float, cash_balance: float, amount: float, term: int, rate: float) -> dict[str, float | int]:
    monthly = payment(amount, term, rate)
    return {
        "amount": round2(amount), "term_months": term, "monthly_payment": round2(monthly),
        "p_gap": p_gap(mu, sigma, debt_old, monthly, cash_balance, term),
    }


def point_a(mu: float, sigma: float, debt_old: float, cash_balance: float, product: Product, term: int, rate: float):
    amount = min(principal(max_payment(mu, sigma, debt_old, cash_balance, term, product), term, rate), product.max_amount)
    return None if amount < product.min_amount else terms(mu, sigma, debt_old, cash_balance, amount, term, rate)


def point_b(mu: float, sigma: float, debt_old: float, cash_balance: float, product: Product, amount: float, from_term: int, rate: float):
    for term in range(from_term + 1, product.max_term + 1):
        if payment(amount, term, rate) <= max_payment(mu, sigma, debt_old, cash_balance, term, product):
            return terms(mu, sigma, debt_old, cash_balance, amount, term, rate)
    return None


def decision(pd: float, fits: bool, has_alternative: bool = True) -> str:
    if pd > 0.25:
        return "REJECTED"
    if pd >= 0.10:
        return "MANUAL_REVIEW"
    if fits:
        return "APPROVED"
    return "COUNTER_OFFER" if has_alternative else "MANUAL_REVIEW"
