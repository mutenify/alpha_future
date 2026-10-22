"""Generate a reproducible synthetic portfolio of 1,000 Russian companies."""

from __future__ import annotations

import csv
from pathlib import Path

import numpy as np


COMPANY_COUNT = 1_000
RANDOM_SEED = 2026
OUTPUT = Path(__file__).with_name("synthetic_1000_companies.csv")

# Order: ln(revenue), margin, ln(taxes), business age in months.
CORRELATION_MATRIX = np.array(
    [
        [1.00, 0.30, 0.78, 0.38],
        [0.30, 1.00, 0.24, 0.18],
        [0.78, 0.24, 1.00, 0.31],
        [0.38, 0.18, 0.31, 1.00],
    ],
    dtype=float,
)
MU = np.array([17.2, 0.22, 14.7, 60.0])
STANDARD_DEVIATIONS = np.array([1.15, 0.10, 1.10, 38.0])


def inn_checksum_is_valid(inn: str) -> bool:
    """Validate the official checksum for a 10- or 12-digit Russian INN."""
    if not inn.isdigit() or len(inn) not in (10, 12):
        return False
    digits = [int(char) for char in inn]
    if len(digits) == 10:
        weights = (2, 4, 10, 3, 5, 9, 4, 6, 8)
        return digits[9] == sum(a * b for a, b in zip(digits, weights)) % 11 % 10
    first_weights = (7, 2, 4, 10, 3, 5, 9, 4, 6, 8)
    second_weights = (3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8)
    first = sum(a * b for a, b in zip(digits[:10], first_weights)) % 11 % 10
    second = sum(a * b for a, b in zip(digits[:11], second_weights)) % 11 % 10
    return digits[10:] == [first, second]


def generate_inn(rng: np.random.Generator, legal_form: str) -> str:
    if legal_form not in ("OOO", "IP"):
        raise ValueError("legal_form must be 'OOO' or 'IP'")
    length = 10 if legal_form == "OOO" else 12
    prefix = [int(value) for value in rng.integers(0, 10, size=length - 1)]
    prefix[0] = int(rng.integers(1, 10))
    if length == 10:
        weights = (2, 4, 10, 3, 5, 9, 4, 6, 8)
        return "".join(map(str, prefix + [sum(a * b for a, b in zip(prefix, weights)) % 11 % 10]))
    first_weights = (7, 2, 4, 10, 3, 5, 9, 4, 6, 8)
    first = sum(a * b for a, b in zip(prefix[:10], first_weights)) % 11 % 10
    prefix[10] = first
    second_weights = (3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8)
    second = sum(a * b for a, b in zip(prefix, second_weights)) % 11 % 10
    return "".join(map(str, prefix + [second]))


def validate_correlation_matrix(matrix: np.ndarray) -> None:
    if matrix.shape != (4, 4):
        raise ValueError("correlation matrix must be 4x4")
    if not np.allclose(matrix, matrix.T):
        raise ValueError("correlation matrix must be symmetric")
    if not np.allclose(np.diag(matrix), 1.0):
        raise ValueError("correlation matrix must have ones on the diagonal")


def main() -> None:
    rng = np.random.default_rng(RANDOM_SEED)
    validate_correlation_matrix(CORRELATION_MATRIX)
    # The Cholesky call rejects a correlation matrix that is not positive definite.
    L = np.linalg.cholesky(CORRELATION_MATRIX)
    D_s = np.diag(STANDARD_DEVIATIONS)
    Z = rng.standard_normal((4, COMPANY_COUNT))
    Y = MU[:, None] + D_s @ L @ Z

    legal_forms = rng.choice(("OOO", "IP"), size=COMPANY_COUNT)
    inns: set[str] = set()
    rows = []
    for index, (ln_revenue, margin, ln_taxes, age) in enumerate(Y.T, start=1):
        legal_form = str(legal_forms[index - 1])
        inn = generate_inn(rng, legal_form)
        while inn in inns:
            inn = generate_inn(rng, legal_form)
        inns.add(inn)
        rows.append(
            (
                index,
                f"Синтетическая компания {index:04d}",
                legal_form,
                inn,
                f"{ln_revenue:.6f}",
                round(float(np.exp(ln_revenue)), 2),
                f"{np.clip(margin, 0.01, 0.80):.6f}",
                f"{ln_taxes:.6f}",
                round(float(np.exp(ln_taxes)), 2),
                max(1, round(float(age))),
            )
        )

    if len(inns) != COMPANY_COUNT or not all(inn_checksum_is_valid(inn) for inn in inns):
        raise RuntimeError("generated INNs are not unique or fail checksum validation")

    with OUTPUT.open("w", encoding="utf-8", newline="") as output_file:
        writer = csv.writer(output_file)
        writer.writerow(
            (
                "company_id",
                "company_name",
                "legal_form",
                "inn",
                "ln_revenue",
                "revenue_rub",
                "margin",
                "ln_taxes",
                "taxes_rub",
                "business_age_months",
            )
        )
        writer.writerows(rows)

    print(f"Generated {len(rows)} companies in {OUTPUT}; all INN checksums are valid.")


if __name__ == "__main__":
    main()
