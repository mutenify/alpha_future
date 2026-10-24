"""Reproducible 1,000-company validation for canonical risk model v1."""

from __future__ import annotations

import csv
import json
import math
from collections import Counter
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np

from app.model import INDUSTRIES, PRODUCTS, decision, equilibrium_rate, max_payment, p_gap, payment, pd_from_score, point_a, point_b, score


SEED = 2026
SIMULATIONS = 10_000


def _auc(labels: np.ndarray, predictions: np.ndarray) -> float:
    order = np.argsort(predictions, kind="mergesort")
    ranks = np.empty(len(predictions), dtype=float)
    start = 0
    while start < len(order):
        end = start + 1
        while end < len(order) and predictions[order[end]] == predictions[order[start]]:
            end += 1
        ranks[order[start:end]] = (start + 1 + end) / 2.0
        start = end
    positives = labels == 1
    n_pos, n_neg = int(positives.sum()), int((~positives).sum())
    return float((ranks[positives].sum() - n_pos * (n_pos + 1) / 2) / (n_pos * n_neg))


def _roc(labels: np.ndarray, predictions: np.ndarray):
    thresholds = np.r_[np.inf, np.sort(np.unique(predictions))[::-1], -np.inf]
    positives, negatives = labels.sum(), len(labels) - labels.sum()
    tpr, fpr = [], []
    for threshold in thresholds:
        predicted = predictions >= threshold
        tpr.append(float((predicted & (labels == 1)).sum() / positives))
        fpr.append(float((predicted & (labels == 0)).sum() / negatives))
    return np.array(fpr), np.array(tpr)


def _portfolio_row(raw: dict[str, str], index: int, rng: np.random.Generator) -> dict:
    industry_code = ("10", "41", "47", "49", "56", "62", "OTHER")[index % 7]
    margin, typical_cv = INDUSTRIES[industry_code]
    annual_revenue = float(raw["revenue_rub"])
    monthly_mean = annual_revenue / 12.0
    seasonality = np.clip(1.0 + rng.normal(0, typical_cv, 12), 0.15, None)
    monthly = monthly_mean * seasonality / seasonality.mean()
    mu = float(monthly.mean() * margin)
    sigma = float(monthly.std(ddof=0) * margin)
    debt_load = float(np.clip(rng.beta(2.0, 3.5) * 1.05, 0, 1.2))
    debt_old = mu * debt_load
    overdues = int(rng.choice([0, 1, 2, 4], p=[0.62, 0.20, 0.13, 0.05]))
    max_overdue = 0 if overdues == 0 else int(rng.choice([7, 14, 30, 60, 90]))
    inquiries = int(rng.poisson(1.7))
    tax_debt = 0.0 if rng.random() < 0.65 else float(np.exp(float(raw["ln_taxes"])) / 12 * rng.uniform(0.02, 0.25))
    cash_balance = float(monthly_mean * rng.lognormal(-0.1, 0.8))
    trend = float(monthly[-3:].mean() / monthly[:3].mean() - 1.0)
    revenue_cv = float(monthly.std(ddof=0) / monthly.mean())
    values = {
        "age_months": float(raw["business_age_months"]), "overdues12m": overdues,
        "max_overdue_days": max_overdue, "inquiries3m": inquiries, "tax_debt": tax_debt,
        "revenue_trend": trend, "revenue_cv": revenue_cv, "debt_load": debt_load,
        "industry_code": industry_code,
    }
    points = score(values)
    pd = pd_from_score(points)
    product = tuple(PRODUCTS.values())[index % len(PRODUCTS)]
    term = int(rng.integers(product.min_term, product.max_term + 1))
    amount = float(rng.uniform(product.min_amount, min(product.max_amount, max(product.min_amount, annual_revenue * 0.8))))
    rate = equilibrium_rate(product.base_rate, pd, product.lgd)
    monthly_payment = payment(amount, term, rate)
    analytical = p_gap(mu, sigma, debt_old, monthly_payment, cash_balance, term)
    a = point_a(mu, sigma, debt_old, cash_balance, product, term, rate)
    b = point_b(mu, sigma, debt_old, cash_balance, product, amount, term, rate)
    fits = monthly_payment <= max_payment(mu, sigma, debt_old, cash_balance, term, product)
    outcome = decision(pd, fits, a is not None or b is not None)

    samples = rng.normal(mu, sigma, SIMULATIONS)
    empirical = float(np.mean(samples + cash_balance / term < debt_old + monthly_payment))
    standard_error = math.sqrt(max(analytical * (1.0 - analytical), 1e-12) / SIMULATIONS)

    mu_stress, sigma_stress = 0.8 * mu, 1.5 * sigma
    stress_gap = p_gap(mu_stress, sigma_stress, debt_old, monthly_payment, cash_balance, term)
    stress_a = point_a(mu_stress, sigma_stress, debt_old, cash_balance, product, term, rate)
    stress_b = point_b(mu_stress, sigma_stress, debt_old, cash_balance, product, amount, term, rate)
    stress_fits = monthly_payment <= max_payment(mu_stress, sigma_stress, debt_old, cash_balance, term, product)
    stress_outcome = decision(pd, stress_fits, stress_a is not None or stress_b is not None)
    return {
        "company_id": raw["company_id"], "inn": raw["inn"], "product_code": product.code,
        "amount": amount, "term_months": term, "score": points, "pd": pd, "mu": mu, "sigma": sigma,
        "cash_balance": cash_balance, "debt_old": debt_old, "payment": monthly_payment,
        "p_gap": analytical, "p_gap_mc": empirical, "mc_abs_error": abs(empirical - analytical),
        "mc_tolerance": 4 * standard_error + 0.005, "decision": outcome,
        "point_a_amount": "" if a is None else a["amount"], "point_a_term": "" if a is None else a["term_months"],
        "point_b_amount": "" if b is None else b["amount"], "point_b_term": "" if b is None else b["term_months"],
        "mu_stress": mu_stress, "sigma_stress": sigma_stress, "p_gap_stress": stress_gap,
        "decision_stress": stress_outcome,
    }


def _plots(rows: list[dict], labels: np.ndarray, image_dir: Path, auc: float, calibration: list[dict]) -> None:
    image_dir.mkdir(parents=True, exist_ok=True)
    plt.style.use("seaborn-v0_8-whitegrid")
    pd_values = np.array([r["pd"] for r in rows])
    fpr, tpr = _roc(labels, pd_values)
    fig, ax = plt.subplots(figsize=(7, 5)); ax.plot(fpr, tpr, lw=2, label=f"ROC-AUC = {auc:.3f}"); ax.plot([0, 1], [0, 1], "--", color="grey"); ax.set(xlabel="False positive rate", ylabel="True positive rate", title="Synthetic default ROC"); ax.legend(); fig.tight_layout(); fig.savefig(image_dir / "roc_auc.png", dpi=160); plt.close(fig)
    fig, ax = plt.subplots(figsize=(7, 5)); ax.plot([d["mean_pd"] for d in calibration], [d["default_rate"] for d in calibration], "o-", label="Observed synthetic defaults"); ax.plot([0, 1], [0, 1], "--", color="grey", label="Perfect calibration"); ax.set(xlabel="Mean predicted PD", ylabel="Observed default rate", title="PD calibration by decile"); ax.legend(); fig.tight_layout(); fig.savefig(image_dir / "pd_calibration.png", dpi=160); plt.close(fig)
    counts = Counter(r["decision"] for r in rows); names = ["APPROVED", "COUNTER_OFFER", "MANUAL_REVIEW", "REJECTED"]; fig, ax = plt.subplots(figsize=(8, 5)); bars=ax.bar(names, [counts[n] for n in names], color=["#2e7d32", "#1976d2", "#ed6c02", "#d32f2f"]); ax.bar_label(bars); ax.set(ylabel="Companies", title="Decision distribution (n=1000)"); ax.tick_params(axis="x", rotation=15); fig.tight_layout(); fig.savefig(image_dir / "decision_distribution.png", dpi=160); plt.close(fig)
    base = np.array([r["p_gap"] for r in rows]); stress = np.array([r["p_gap_stress"] for r in rows]); fig, ax = plt.subplots(figsize=(7, 5)); ax.scatter(base, stress, s=10, alpha=.45); ax.plot([0, 1], [0, 1], "--", color="grey"); ax.set(xlabel="Base PGap", ylabel="Stressed PGap", title="Stress: μ×0.8, σ×1.5"); fig.tight_layout(); fig.savefig(image_dir / "stress_test.png", dpi=160); plt.close(fig)


def run_validation(input_csv: Path, image_dir: Path, results_csv: Path) -> dict:
    rng = np.random.default_rng(SEED)
    with input_csv.open(encoding="utf-8") as source:
        source_rows = list(csv.DictReader(source))
    if len(source_rows) != 1_000:
        raise ValueError(f"expected 1000 companies, got {len(source_rows)}")
    rows = [_portfolio_row(raw, index, rng) for index, raw in enumerate(source_rows)]
    if any(r["mc_abs_error"] > r["mc_tolerance"] for r in rows):
        raise AssertionError("Monte Carlo deviation exceeds the four-sigma tolerance")
    pd_values = np.array([r["pd"] for r in rows])
    labels = rng.binomial(1, pd_values)
    auc = _auc(labels, pd_values)
    ordered = np.argsort(pd_values)
    calibration = []
    for decile, indexes in enumerate(np.array_split(ordered, 10), start=1):
        calibration.append({"decile": decile, "count": len(indexes), "mean_pd": float(pd_values[indexes].mean()), "default_rate": float(labels[indexes].mean())})
    _plots(rows, labels, image_dir, auc, calibration)
    with results_csv.open("w", encoding="utf-8", newline="") as target:
        writer = csv.DictWriter(target, fieldnames=list(rows[0])); writer.writeheader(); writer.writerows(rows)
    decisions = Counter(r["decision"] for r in rows)
    stress_decisions = Counter(r["decision_stress"] for r in rows)
    summary = {
        "seed": SEED, "companies": len(rows), "simulations_per_company": SIMULATIONS,
        "roc_auc": auc, "gini": 2 * auc - 1, "decisions": dict(decisions),
        "stress_decisions": dict(stress_decisions), "mean_p_gap": float(np.mean([r["p_gap"] for r in rows])),
        "mean_p_gap_stress": float(np.mean([r["p_gap_stress"] for r in rows])),
        "mean_mc_abs_error": float(np.mean([r["mc_abs_error"] for r in rows])),
        "max_mc_abs_error": float(np.max([r["mc_abs_error"] for r in rows])),
        "calibration": calibration,
        "warning": "Synthetic labels sampled from model PD; this is not independent validation on real defaults.",
    }
    results_csv.with_suffix(".summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    return summary


if __name__ == "__main__":
    root = Path(__file__).resolve().parents[1]
    print(json.dumps(run_validation(root / "data/synthetic_1000_companies.csv", root / "docs/images", root / "ml/validation_1000_results.csv"), ensure_ascii=False, indent=2))
