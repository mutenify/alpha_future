"""Generate the cross-language golden cases from fixed demo-company source data."""

from __future__ import annotations

import csv
import hashlib
import json
import math
import random
from datetime import date
from pathlib import Path

from app.model import INDUSTRIES, PRODUCTS, decision, equilibrium_rate, max_payment, p_gap, payment, pd_from_score, point_a, point_b, score


AS_OF = date(2026, 10, 1)
CASES = [
    (1, "client1", "7701234560", "WORKING_CAPITAL", 300_000, 12, "Пополнение оборотных средств"),
    (2, "client2", "7812345675", "WORKING_CAPITAL", 3_000_000, 12, "Закупка сырья"),
    (3, "client3", "5401234566", "INVESTMENT_LOAN", 5_000_000, 36, "Покупка оборудования"),
    (4, "client4", "6671234569", "WORKING_CAPITAL", 8_000_000, 24, "Расширение производства"),
    (5, "client5", "5251234564", "WORKING_CAPITAL", 1_000_000, 12, "Закупка материалов"),
    (6, "client6", "7700000503", "WORKING_CAPITAL", 1_000_000, 12, "Пополнение оборотных средств"),
    (7, "client7", "1651234561", "OVERDRAFT", 700_000, 6, "Кассовый разрыв"),
    (8, "client8", "2311234566", "WORKING_CAPITAL", 6_000_000, 18, "Сезонная закупка"),
    (9, "client9", "6161234561", "INVESTMENT_LOAN", 15_000_000, 48, "Строительная техника"),
    (10, "client10", "7707083893", "INVESTMENT_LOAN", 20_000_000, 24, "Разработка платформы"),
]


def _rng(key: str) -> random.Random:
    return random.Random(int(hashlib.sha256(key.encode()).hexdigest(), 16))


def _monthly_revenue(inn: str, annual: int) -> list[float]:
    rng = _rng(f"monthly:{inn}")
    average = annual / 12
    return [round(average * rng.uniform(0.82, 1.18) / 1000) * 1000 for _ in range(12)]


def _mean_std(values: list[float]) -> tuple[float, float]:
    mean = sum(values) / len(values)
    return mean, math.sqrt(sum((value - mean) ** 2 for value in values) / len(values))


def _profile(company: dict) -> dict:
    revenues = _monthly_revenue(company["inn"], company["fns"]["annualRevenue"])
    mean, std = _mean_std(revenues)
    industry_code = company["egrul"]["okved"][:2]
    if industry_code not in INDUSTRIES:
        industry_code = "OTHER"
    margin, _ = INDUSTRIES[industry_code]
    mu, sigma = round(mean * margin, 2), round(std * margin, 2)
    first, last = sum(revenues[:3]) / 3, sum(revenues[-3:]) / 3
    return {
        "age_months": (AS_OF.year - date.fromisoformat(company["egrul"]["registeredAt"]).year) * 12 + AS_OF.month - date.fromisoformat(company["egrul"]["registeredAt"]).month,
        "overdues12m": company["bki"]["overdues12m"], "max_overdue_days": company["bki"]["maxOverdueDays"],
        "inquiries3m": company["bki"]["inquiries3m"], "tax_debt": company["fns"]["taxDebt"],
        "revenue_trend": round(last / first - 1, 4), "revenue_cv": round(std / mean, 4),
        "debt_load": company["abs"]["currentDebtPayments"] / mu if mu else 0,
        "industry_code": industry_code, "mu": mu, "sigma": sigma,
        "debt_old": float(company["abs"]["currentDebtPayments"]), "cash_balance": float(company["abs"]["cashBalance"]),
    }


def generate() -> list[dict]:
    root = Path(__file__).resolve().parents[1]
    companies = json.loads((root / "ml/app/companies.json").read_text(encoding="utf-8"))
    rows = []
    for case_id, login, inn, product_code, amount, term, purpose in CASES:
        base = {"case_id": case_id, "login": login, "inn": inn, "product_code": product_code, "amount": amount, "term_months": term, "purpose": purpose}
        if inn == "7700000503":
            rows.append({**base, "expected_status": "MANUAL_REVIEW", "special": "SOURCE_UNAVAILABLE"})
            continue
        company = {"inn": inn, **companies[inn]}
        if company["egrul"].get("stopFactor"):
            rows.append({**base, "expected_status": "REJECTED", "special": company["egrul"]["stopFactor"]})
            continue
        profile = _profile(company)
        points = score(profile)
        pd = pd_from_score(points)
        product = PRODUCTS[product_code]
        rate = equilibrium_rate(product.base_rate, pd, product.lgd)
        monthly_payment = payment(amount, term, rate)
        gap = p_gap(profile["mu"], profile["sigma"], profile["debt_old"], monthly_payment, profile["cash_balance"], term)
        point_a_terms = point_a(profile["mu"], profile["sigma"], profile["debt_old"], profile["cash_balance"], product, term, rate)
        point_b_terms = point_b(profile["mu"], profile["sigma"], profile["debt_old"], profile["cash_balance"], product, amount, term, rate)
        fits = monthly_payment <= max_payment(profile["mu"], profile["sigma"], profile["debt_old"], profile["cash_balance"], term, product)
        outcome = decision(pd, fits, point_a_terms is not None or point_b_terms is not None)
        rows.append({
            **base, **profile, "score": points, "pd": pd, "rate": rate, "payment": monthly_payment,
            "p_gap": gap, "expected_status": outcome, "special": "",
            "point_a_amount": "" if point_a_terms is None else point_a_terms["amount"],
            "point_a_term": "" if point_a_terms is None else point_a_terms["term_months"],
            "point_a_payment": "" if point_a_terms is None else point_a_terms["monthly_payment"],
            "point_a_p_gap": "" if point_a_terms is None else point_a_terms["p_gap"],
            "point_b_amount": "" if point_b_terms is None else point_b_terms["amount"],
            "point_b_term": "" if point_b_terms is None else point_b_terms["term_months"],
            "point_b_payment": "" if point_b_terms is None else point_b_terms["monthly_payment"],
            "point_b_p_gap": "" if point_b_terms is None else point_b_terms["p_gap"],
        })
    return rows


def write(path: Path) -> list[dict]:
    rows = generate()
    fieldnames = list(dict.fromkeys(key for row in rows for key in row))
    with path.open("w", encoding="utf-8", newline="") as target:
        writer = csv.DictWriter(target, fieldnames=fieldnames); writer.writeheader(); writer.writerows(rows)
    return rows


if __name__ == "__main__":
    root = Path(__file__).resolve().parents[1]
    for row in write(root / "backend/internal/scoring/testdata/golden_cases.csv"):
        print(row["case_id"], row["login"], row["expected_status"], row.get("score", "-"), row.get("pd", "-"))
